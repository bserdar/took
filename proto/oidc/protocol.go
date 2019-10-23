package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	jwt "gopkg.in/square/go-jose.v2/jwt"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto"
)

const stateRandomLength = 32

// Data contains the tokens
type Data struct {
	Last   string
	Tokens []TokenData
}

// TokenData contains the access and refresh token with username
type TokenData struct {
	Username     string
	AccessToken  string
	RefreshToken string
	Type         string
}

// Protocol contains the oidc config, default congfig, and tokens
type Protocol struct {
	Cfg      Config
	Defaults Config
	Tokens   Data
}

func (d Data) findUser(username string) *TokenData {
	for i, x := range d.Tokens {
		if x.Username == username {
			return &d.Tokens[i]
		}
	}
	return nil
}

// DecodeCfg converts map[string]interface{} into Config{}
func (p *Protocol) DecodeCfg(in interface{}) (interface{}, error) {
	if in != nil {
		out := Config{}
		cfg.Decode(in, &out)
		return out, nil
	}
	return nil, nil
}

// SetCfg sets the p.Cfg and p.Defaults from user and common configs
func (p *Protocol) SetCfg(user, common cfg.Remote) {
	if user.Configuration != nil {
		cfg.Decode(user.Configuration, &p.Cfg)
	}
	if user.Data != nil {
		cfg.Decode(user.Data, &p.Tokens)
	}
	if common.Configuration != nil {
		cfg.Decode(common.Configuration, &p.Defaults)
	}
}

// GetConfig merges default cfg with user cfg and returns a merged copy
func (p *Protocol) GetConfig() Config {
	ret := p.Cfg

	// First look at server profile reference
	profile := cfg.GetServerProfile(ret.Profile)
	if len(profile.Type) > 0 {
		if profile.Type != "oidc" && profile.Type != "oidc-auth" {
			panic("Server profile is not for oidc")
		}
		sp := ServerProfile{}
		cfg.Decode(profile.Configuration, &sp)
		ret.ServerProfile = ret.ServerProfile.Merge(sp)
	}
	ret = ret.Merge(p.Defaults)
	return ret
}

func init() {
	proto.Register("oidc-auth", func() proto.Protocol {
		return &Protocol{}
	})
	proto.Register("oidc", func() proto.Protocol {
		return &Protocol{}
	})
}

// FormatToken converts token to string based on the output options
func (t TokenData) FormatToken(out proto.OutputOption) string {
	switch out {
	case proto.OutputHeader:
		return fmt.Sprintf("Authorization: %s %s", http.CanonicalHeaderKey(t.Type),
			t.AccessToken)
	}
	return t.AccessToken
}

// GetToken gets a token
func (p *Protocol) GetToken(request proto.TokenRequest) (string, interface{}, error) {
	config := p.GetConfig()
	if config.Insecure {
		proto.InsecureTLS = true
	}
	// If there is a username, use that. Otherwise, use last
	userName := request.Username
	if userName == "" {
		userName = p.Tokens.Last
	}

	if userName == "" {
		log.Fatalf("Username is required for oidc auth")
		return "", nil, nil
	}
	var tok *TokenData
	tok = p.Tokens.findUser(userName)
	if tok == nil {
		p.Tokens.Tokens = append(p.Tokens.Tokens, TokenData{})
		tok = &p.Tokens.Tokens[len(p.Tokens.Tokens)-1]
		tok.Username = userName
	}
	p.Tokens.Last = tok.Username

	serverData, err := GetServerData(config.URL)
	if err != nil {
		return "", nil, err
	}
	if request.Refresh != proto.UseReAuth {
		if tok.AccessToken != "" {
			log.Debugf("There is an access token, validating")
			if p.Validate(tok.AccessToken, serverData) {
				log.Debug("Token is valid")
				// Token may be valid, but too close to expiration
				if !p.TooClose(tok.AccessToken, serverData) {
					log.Debug("But expiration is too close")
					if request.Refresh != proto.UseRefresh {
						return tok.FormatToken(request.Out), p.Tokens, nil
					}
				}
			}
			if tok.RefreshToken != "" {
				log.Debug("Refreshing token")
				err := p.Refresh(tok, serverData)
				if err == nil {
					return tok.FormatToken(request.Out), p.Tokens, nil
				}
			}
		}
	}

	conf := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scopes:       []string{"openid"},
		RedirectURL:  config.CallbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  p.GetAuthURL(serverData),
			TokenURL: p.GetTokenURL(serverData)}}

	// Generate a crytographically secure random token for the state.
	stateBytes := make([]byte, stateRandomLength)
	_, err = rand.Read(stateBytes)
	if err != nil {
		return "", nil, err
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	var token *oauth2.Token
	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, proto.GetHTTPClient())
	conf.Scopes = append(conf.Scopes, config.AdditionalScopes...)
	log.Debugf("Password grant: %v", config.PasswordGrant)
	if config.RefreshOnly != nil && *config.RefreshOnly {
		tok.RefreshToken = cfg.AskPasswordWithPrompt(fmt.Sprintf("Refresh token for %s: ", userName))
		err := p.Refresh(tok, serverData)
		if err != nil {
			return "", nil, err
		}
		return tok.FormatToken(request.Out), p.Tokens, nil
	} else if config.PasswordGrant != nil && *config.PasswordGrant {
		var password string
		if len(request.Password) > 0 {
			password = request.Password
		} else {
			password = cfg.AskPasswordWithPrompt(fmt.Sprintf("Password for %s: ", userName))
		}
		token, err = conf.PasswordCredentialsToken(ctx, userName, password)
		if err != nil {
			return "", nil, err
		}
	} else {
		authURL := conf.AuthCodeURL(state, oauth2.AccessTypeOnline)
		var redirectedURL *url.URL
		if config.Form != nil {
			redirectedURL = FormAuth(*config.Form, authURL, userName, request.Password)
			if redirectedURL == nil {
				fmt.Printf("Authentication failed\n")
			}
		}
		if redirectedURL == nil {
			inURL := cfg.Ask(fmt.Sprintf(`Go to this URL to authenticate %s: %s
After authentication, copy/paste the URL here:`, userName, authURL))
			redirectedURL, err = url.Parse(inURL)
			if err != nil {
				return "", nil, err
			}
			if state != redirectedURL.Query().Get("state") {
				return "", nil, fmt.Errorf("Invalid state")
			}
		}
		token, err = conf.Exchange(ctx, redirectedURL.Query().Get("code"))
		if err != nil {
			return "", nil, err
		}
	}

	tok.AccessToken = token.AccessToken
	tok.RefreshToken = token.RefreshToken
	tok.Type = token.TokenType

	return tok.FormatToken(request.Out), p.Tokens, nil
}

// Refresh refreshes the token
func (p *Protocol) Refresh(tok *TokenData, s ServerData) error {
	cfg := p.GetConfig()
	t, err := RefreshToken(cfg.ClientID, cfg.ClientSecret, tok.RefreshToken, p.GetTokenURL(s))
	if err != nil {
		return err
	}
	tok.AccessToken = t.AccessToken
	tok.RefreshToken = t.RefreshToken
	tok.Type = t.TokenType
	return nil
}

// GetTokenURL retutrns the token URL on the auth server
func (p *Protocol) GetTokenURL(s ServerData) string {
	cfg := p.GetConfig()
	if len(cfg.TokenAPI) == 0 {
		return s.TokenEndpoint
	}
	return combine(cfg.URL, cfg.TokenAPI)
}

// GetAuthURL returns the auth URL on the auth server
func (p *Protocol) GetAuthURL(s ServerData) string {
	cfg := p.GetConfig()
	if len(cfg.AuthAPI) == 0 {
		return s.AuthorizationEndpoint
	}
	return combine(cfg.URL, cfg.AuthAPI)
}

func combine(base, suffix string) string {
	if strings.HasPrefix(suffix, "/") {
		suffix = suffix[1:]
	}
	if !strings.HasSuffix(base, "/") {
		base = base + "/"
	}
	return base + suffix
}

// TooClose returns true if the token expiration is too close: 1m if
// token lifetime is more than 1m, or token lifetime if not
func (p *Protocol) TooClose(accessToken string, serverData ServerData) bool {
	t, err := jwt.ParseSigned(accessToken)
	if err == nil {
		var c jwt.Claims
		t.UnsafeClaimsWithoutVerification(&c)
		return tooClose(c.Expiry.Time(), time.Now())
	}
	return false
}

func tooClose(expiry, now time.Time) bool {
	return expiry.Sub(now) < 30*time.Second
}

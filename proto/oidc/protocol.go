package oidc

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/bserdar/took/proto"
)

type Config struct {
	ClientId     string
	ClientSecret string
	URL          string
	CallbackURL  string
	TokenAPI     string
	AuthAPI      string
	Form         *HTMLFormConfig
}

// Data contains the tokens
type Data struct {
	Last   string
	Tokens []TokenData
}

type TokenData struct {
	Username     string
	AccessToken  string
	RefreshToken string
	Type         string
}

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

// Merge sets the unset fields of c from defaults
func (c Config) Merge(defaults Config) Config {
	wdef := func(s, def string) string {
		if len(s) > 0 {
			return s
		}
		return def
	}
	return Config{ClientId: wdef(c.ClientId, defaults.ClientId),
		ClientSecret: wdef(c.ClientSecret, defaults.ClientSecret),
		URL:          wdef(c.URL, defaults.URL),
		CallbackURL:  wdef(c.CallbackURL, defaults.CallbackURL),
		TokenAPI:     wdef(c.TokenAPI, defaults.TokenAPI),
		AuthAPI:      wdef(c.AuthAPI, defaults.AuthAPI)}
}

// GetConfigInstance returns a pointer to a new config
func (p *Protocol) GetConfigInstance() interface{} { return &p.Cfg }

func (p *Protocol) GetConfigDefaultsInstance() interface{} { return &p.Defaults }

func (p *Protocol) GetDataInstance() interface{} { return &p.Tokens }

func (p *Protocol) ConfigWithDefaults() Config {
	return p.Cfg.Merge(p.Defaults)
}

func init() {
	proto.Register("oidc-auth", func() proto.Protocol {
		return &Protocol{}
	})
}

func (t TokenData) FormatToken(out proto.OutputOption) string {
	switch out {
	case proto.OutputHeader:
		return fmt.Sprintf("Authorization: %s %s", t.Type, t.AccessToken)
	}
	return t.AccessToken
}

// GetToken gets a token
func (p *Protocol) GetToken(request proto.TokenRequest) (string, error) {
	cfg := p.ConfigWithDefaults()
	// If there is a username, use that. Otherwise, use last
	userName := request.Username
	if userName == "" {
		userName = p.Tokens.Last
	}

	if userName == "" {
		log.Fatalf("Username is required for oidc quth")
		return "", nil
	}
	var tok *TokenData
	tok = p.Tokens.findUser(userName)
	if tok == nil {
		p.Tokens.Tokens = append(p.Tokens.Tokens, TokenData{})
		tok = &p.Tokens.Tokens[len(p.Tokens.Tokens)-1]
		tok.Username = userName
	}
	p.Tokens.Last = tok.Username

	if request.Refresh != proto.UseReAuth {
		if tok.AccessToken != "" {
			log.Debugf("There is an access token, validating")
			if Validate(tok.AccessToken, cfg.URL) {
				log.Debug("Token is valid")
				if request.Refresh != proto.UseRefresh {
					return tok.FormatToken(request.Out), nil
				}
			}
			if tok.RefreshToken != "" {
				log.Debug("Refreshing token")
				err := p.Refresh(tok)
				if err == nil {
					return tok.FormatToken(request.Out), nil
				}
			}
		}
	}

	conf := &oauth2.Config{
		ClientID:     cfg.ClientId,
		ClientSecret: cfg.ClientSecret,
		Scopes:       []string{"openid"},
		RedirectURL:  cfg.CallbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  p.GetAuthURL(),
			TokenURL: p.GetTokenURL()}}
	state := fmt.Sprintf("%x", rand.Uint64())
	authUrl := conf.AuthCodeURL(state, oauth2.AccessTypeOnline)
	var redirectedUrl *url.URL
	if p.Cfg.Form != nil {
		redirectedUrl = FormAuth(*p.Cfg.Form, authUrl)
	}
	if redirectedUrl == nil {
		fmt.Printf("Go to this URL to authenticate %s: %s\n", userName, authUrl)
		inUrl, err := proto.Ask("After authentication, copy/paste the URL here:")
		if err != nil {
			log.Fatalf("%s", err)
		}
		redirectedUrl, err = url.Parse(inUrl)
		if err != nil {
			log.Fatal(err.Error())
		}
		if state != redirectedUrl.Query().Get("state") {
			log.Fatal("Invalid state")
		}
	}

	token, err := conf.Exchange(context.Background(), redirectedUrl.Query().Get("code"))
	if err != nil {
		log.Fatal(err)
	}

	tok.AccessToken = token.AccessToken
	tok.RefreshToken = token.RefreshToken
	tok.Type = token.TokenType

	return tok.FormatToken(request.Out), nil
}

func (p *Protocol) Refresh(tok *TokenData) error {
	cfg := p.ConfigWithDefaults()
	t, err := RefreshToken(cfg.ClientId, cfg.ClientSecret, tok.RefreshToken, p.GetTokenURL())
	if err != nil {
		return err
	}
	tok.AccessToken = t.AccessToken
	tok.RefreshToken = t.RefreshToken
	tok.Type = t.TokenType
	return nil
}

func (p *Protocol) GetTokenURL() string {
	cfg := p.ConfigWithDefaults()
	token := cfg.TokenAPI
	if token == "" {
		token = "protocol/openid-connect/token"
	}
	return combine(cfg.URL, token)
}

func (p *Protocol) GetAuthURL() string {
	cfg := p.ConfigWithDefaults()
	auth := cfg.AuthAPI
	if auth == "" {
		auth = "protocol/openid-connect/auth"
	}
	return combine(cfg.URL, auth)
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

package oidc_da

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/bserdar/took/proto"
	"github.com/bserdar/took/proto/oidc"
)

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
	Cfg      oidc.Config
	Defaults oidc.Config
	Tokens   Data
}

// GetConfigInstance returns a pointer to a new config
func (p *Protocol) GetConfigInstance() interface{} { return &p.Cfg }

func (p *Protocol) GetConfigDefaultsInstance() interface{} { return &p.Defaults }

func (p *Protocol) GetDataInstance() interface{} { return &p.Tokens }

func init() {
	proto.Register("oidc-direct-access", func() proto.Protocol {
		return &Protocol{}
	})
}

func (d Data) findUser(username string) *TokenData {
	for i, x := range d.Tokens {
		if x.Username == username {
			return &d.Tokens[i]
		}
	}
	return nil
}

func (t TokenData) FormatToken(out proto.OutputOption) string {
	switch out {
	case proto.OutputHeader:
		return fmt.Sprintf("Authorization: %s %s", t.Type, t.AccessToken)
	}
	return t.AccessToken
}

func (p *Protocol) ConfigWithDefaults() oidc.Config {
	return p.Cfg.Merge(p.Defaults)
}

// GetToken gets a token
func (p *Protocol) GetToken(request proto.TokenRequest) (string, error) {
	cfg := p.ConfigWithDefaults()
	// If there is a username, use that. Otherwise, use last
	userName := request.Username
	log.Debugf("Request username: %s", userName)
	if userName == "" {
		userName = p.Tokens.Last
		log.Debugf("Last username: %s", userName)
	}
	if userName == "" {
		userName, _ = proto.AskUsername()
		if userName == "" {
			return "", errors.New("No user")
		}
	}

	tok := p.Tokens.findUser(userName)
	if tok == nil {
		p.Tokens.Tokens = append(p.Tokens.Tokens, TokenData{})
		tok = &p.Tokens.Tokens[len(p.Tokens.Tokens)-1]
		tok.Username = userName
	}
	p.Tokens.Last = tok.Username

	if request.Refresh != proto.UseReAuth {
		if tok.AccessToken != "" {
			log.Debugf("There is an access token, validating")
			if oidc.Validate(tok.AccessToken, cfg.URL) {
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

	password, _ := proto.AskPassword()
	err := p.GetNewToken(tok, password)
	if err != nil {
		return "", err
	}
	return tok.FormatToken(request.Out), nil
}

// GetNewTokens gets a new token from the server
func (p *Protocol) GetTokenAPI() string {
	cfg := p.ConfigWithDefaults()
	url := cfg.URL
	tok := cfg.TokenAPI
	if tok == "" {
		tok = "protocol/openid-connect/token"
	}
	if strings.HasPrefix(tok, "/") {
		tok = tok[1:]
	}
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return url + tok
}

func (p *Protocol) GetNewToken(tok *TokenData, password string) error {
	cfg := p.ConfigWithDefaults()
	values := url.Values{}
	values.Set("client_id", cfg.ClientId)
	values.Set("client_secret", cfg.ClientSecret)
	values.Set("username", tok.Username)
	values.Set("password", password)
	values.Set("grant_type", "password")
	log.Debugf("Get new token, url:%s values: %v", p.GetTokenAPI, values)
	resp, err := http.PostForm(p.GetTokenAPI(), values)
	if err != nil {
		log.Debugf("Get new token returns: %s", err)
		return err
	}
	if resp.StatusCode != 200 {
		log.Debugf("Get new token returns: %s", resp.Status)
		return fmt.Errorf("Cannot get token: %s", resp.Status)
	}
	defer resp.Body.Close()
	var d oauth2.Token
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		return err
	}
	log.Debugf("Tokens: %v", d)
	tok.AccessToken = d.AccessToken
	tok.RefreshToken = d.RefreshToken
	tok.Type = d.TokenType
	return nil
}

func (p *Protocol) Refresh(tok *TokenData) error {
	cfg := p.ConfigWithDefaults()
	t, err := oidc.RefreshToken(cfg.ClientId, cfg.ClientSecret, tok.RefreshToken, p.GetTokenAPI())
	if err != nil {
		return err
	}
	tok.AccessToken = t.AccessToken
	tok.RefreshToken = t.RefreshToken
	tok.Type = t.TokenType
	return nil
}

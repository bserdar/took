package oidc_da

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
	//	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"

	"github.com/bserdar/took/proto"
	"github.com/bserdar/took/proto/oidc"
)

// Config is the OIDC-Connect direct access configuration
type Config struct {
	ClientId     string
	ClientSecret string
	URL          string
	TokenAPI     string
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
	Cfg    Config
	Tokens Data
}

// GetConfigInstance returns a pointer to a new config
func (p *Protocol) GetConfigInstance() interface{} { return &p.Cfg }

func (p *Protocol) GetDataInstance() interface{} { return &p.Tokens }

var instance Protocol

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

// GetToken gets a token
func (p *Protocol) GetToken(request proto.TokenRequest) (string, error) {
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
			if oidc.Validate(tok.AccessToken, p.Cfg.URL) {
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
	err = p.GetNewToken(tok, password)
	if err != nil {
		return "", err
	}
	return tok.FormatToken(request.Out), nil
}

// GetNewTokens gets a new token from the server
func (p *Protocol) GetTokenAPI() string {
	url := p.Cfg.URL
	tok := p.Cfg.TokenAPI
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
	values := url.Values{}
	values.Set("client_id", p.Cfg.ClientId)
	values.Set("client_secret", p.Cfg.ClientSecret)
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
	values := url.Values{}
	values.Set("client_id", p.Cfg.ClientId)
	values.Set("client_secret", p.Cfg.ClientSecret)
	values.Set("refresh_token", tok.RefreshToken)
	values.Set("grant_type", "refresh_token")
	log.Debugf("Refresh %s %v", p.GetTokenAPI(), values)
	resp, err := http.PostForm(p.GetTokenAPI(), values)
	if err != nil {
		log.Debugf("Refresh token returns: %s", err)
		return err
	}
	if resp.StatusCode != 200 {
		log.Debugf("Refresh token returns: %s", resp.Status)
		return fmt.Errorf("Cannot refresh token: %s", resp.Status)
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

package oidc_da

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"

	"github.com/bserdar/took/proto"
)

// Config is the OIDC-Connect direct access configuration
type Config struct {
	ClientId     string
	ClientSecret string
	URL          string
	TokenAPI     string
	UserName     string
}

// Data contains the tokens
type Data struct {
	AccessToken  string
	RefreshToken string
	Type         string
}

type ServerData struct {
	Realm          string `json:"realm"`
	PublicKey      string `json:"public_key"`
	TokenService   string `json:"token_service"`
	AccountService string `json:"account_service"`
	PK             interface{}
}

type jsonData struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Type         string `json:"token_type,omitempty"`
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

func (p *Protocol) FormatToken() string {
	return fmt.Sprintf("%s %s", p.Tokens.Type, p.Tokens.AccessToken)
}

// GetToken gets a token
func (p *Protocol) GetToken(forceNew bool) (string, error) {
	// Do we already have a token
	svc, err := GetServiceInfo(p.Cfg.URL)
	if err != nil {
		return "", err
	}
	log.Debugf("Service info: %v", svc)
	if !forceNew && p.Tokens.AccessToken != "" {
		log.Debugf("There is an access token, validating")
		token, _ := jwt.Parse(p.Tokens.AccessToken, func(*jwt.Token) (interface{}, error) {
			return svc.PK, nil
		})
		if token != nil {
			if token.Valid {
				log.Debug("Token is valid")
				return p.FormatToken(), nil
			}
			log.Debug("Token is not valid")
			if p.Tokens.RefreshToken != "" {
				log.Debug("Refreshing token")
				tokens, err := p.Refresh()
				if err != nil {
					return "", err
				}
				p.Tokens = tokens
				return p.FormatToken(), nil
			}
		}
	}

	username, _ := proto.AskUsername()
	password, _ := proto.AskPassword()
	tokens, err := p.GetNewToken(username, password)
	if err != nil {
		return "", err
	}
	p.Tokens = tokens
	return p.FormatToken(), nil
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

func (p *Protocol) GetNewToken(username, password string) (Data, error) {
	values := url.Values{}
	values.Set("client_id", p.Cfg.ClientId)
	values.Set("client_secret", p.Cfg.ClientSecret)
	values.Set("username", username)
	values.Set("password", password)
	values.Set("grant_type", "password")
	log.Debugf("Get new token, url:%s values: %v", p.GetTokenAPI, values)
	resp, err := http.PostForm(p.GetTokenAPI(), values)
	if err != nil {
		log.Debugf("Get new token returns: %s", err)
		return Data{}, err
	}
	if resp.StatusCode != 200 {
		log.Debugf("Get new token returns: %s", resp.Status)
		return Data{}, fmt.Errorf("Cannot get token: %s", resp.Status)
	}
	defer resp.Body.Close()
	var d jsonData
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		return Data{}, err
	}
	log.Debugf("Tokens: %v", d)
	return Data{AccessToken: d.AccessToken,
		RefreshToken: d.RefreshToken,
		Type:         d.Type}, nil
}

func GetServiceInfo(url string) (ServerData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return ServerData{}, err
	}
	defer resp.Body.Close()
	var d ServerData
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		return d, err
	}

	b, err := base64.StdEncoding.DecodeString(d.PublicKey)
	if err != nil {
		return d, err
	}
	d.PK, err = x509.ParsePKIXPublicKey(b)
	if err != nil {
		return d, err
	}
	return d, nil

}

func (p *Protocol) Refresh() (Data, error) {
	values := url.Values{}
	values.Set("client_id", p.Cfg.ClientId)
	values.Set("client_secret", p.Cfg.ClientSecret)
	values.Set("refresh_token", p.Tokens.RefreshToken)
	values.Set("grant_type", "refresh_token")
	log.Debugf("Refresh %s %v", p.GetTokenAPI(), values)
	resp, err := http.PostForm(p.GetTokenAPI(), values)
	if err != nil {
		log.Debugf("Refresh token returns: %s", err)
		return Data{}, err
	}
	if resp.StatusCode != 200 {
		log.Debugf("Refresh token returns: %s", resp.Status)
		return Data{}, fmt.Errorf("Cannot refresh token: %s", resp.Status)
	}
	defer resp.Body.Close()
	var d jsonData
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		return Data{}, err
	}
	log.Debugf("Tokens: %v", d)
	return Data{AccessToken: d.AccessToken,
		RefreshToken: d.RefreshToken,
		Type:         d.Type}, nil

}

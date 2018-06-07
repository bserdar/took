package oidc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// RefreshToken gets a new token using the refresh token
func RefreshToken(clientId, clientSecret, refreshToken, tokenURL string) (oauth2.Token, error) {
	values := url.Values{}
	values.Set("client_id", clientId)
	values.Set("client_secret", clientSecret)
	values.Set("refresh_token", refreshToken)
	values.Set("grant_type", "refresh_token")
	log.Debugf("Refresh %s %v", tokenURL, values)
	resp, err := http.PostForm(tokenURL, values)
	if err != nil {
		log.Debugf("Refresh token returns: %s", err)
		return oauth2.Token{}, err
	}
	if resp.StatusCode != 200 {
		log.Debugf("Refresh token returns: %s", resp.Status)
		return oauth2.Token{}, fmt.Errorf("Cannot refresh token: %s", resp.Status)
	}
	defer resp.Body.Close()
	var d oauth2.Token
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		return oauth2.Token{}, err
	}
	log.Debugf("Tokens: %v", d)
	return d, nil
}

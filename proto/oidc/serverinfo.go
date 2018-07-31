package oidc

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"

	"github.com/bserdar/took/proto"
)

// ServerData contains the OIDC server information
type ServerData struct {
	Realm          string `json:"realm"`
	PublicKey      string `json:"public_key"`
	TokenService   string `json:"token_service"`
	AccountService string `json:"account_service"`
	PK             interface{}
}

// GetServerData retrieves server data from the auth server
func GetServerData(url string) (ServerData, error) {
	resp, err := proto.HTTPGet(url)
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

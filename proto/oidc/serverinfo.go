package oidc

import (
	"encoding/json"
	"fmt"

	"github.com/bserdar/took/proto"
	log "github.com/sirupsen/logrus"
)

// ServerData contains the OIDC server information
type ServerData struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	IntrospectionEndpoint string `json:"token_introspection_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
	JWKSUri               string `json:"jwks_uri"`
}

// GetServerData retrieves server data from the auth server
func GetServerData(url string) (ServerData, error) {
	cfgUrl := combine(url, ".well-known/openid-configuration")
	log.Debugf("Getting server info from %s", cfgUrl)
	resp, err := proto.HTTPGet(cfgUrl)
	if err != nil {
		return ServerData{}, err
	}
	defer resp.Body.Close()
	var d ServerData
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		err = fmt.Errorf("Cannot get SSO server information from %s: %s", cfgUrl, err.Error())
	}
	return d, err
}

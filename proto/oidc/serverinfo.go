package oidc

import (
	"encoding/json"

	"github.com/bserdar/took/proto"
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
	resp, err := proto.HTTPGet(combine(url, ".well-known/openid-configuration"))
	if err != nil {
		return ServerData{}, err
	}
	defer resp.Body.Close()
	var d ServerData
	err = json.NewDecoder(resp.Body).Decode(&d)
	return d, err
}

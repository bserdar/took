package oidc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/bserdar/took/proto"
)

// Validate checks if a token is valid
func (p *Protocol) Validate(accessToken string, serverData ServerData) bool {
	cli := proto.GetHTTPClient()
	req, _ := http.NewRequest(http.MethodPost, serverData.IntrospectionEndpoint,
		strings.NewReader(fmt.Sprintf("token=%s", accessToken)))
	req.SetBasicAuth(p.Cfg.ClientID, p.Cfg.ClientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	log.Debugf("Sending introspection request to %s", serverData.IntrospectionEndpoint)
	response, err := cli.Do(req)
	if err != nil {
		log.Debugf("Introspection error: %s", err.Error())
	} else {
		defer response.Body.Close()
		var m map[string]interface{}
		json.NewDecoder(response.Body).Decode(&m)
		log.Debugf("Response: %v", m)
		if active, ok := m["active"]; ok {
			if b, ok := active.(bool); ok {
				return b
			}
		}
	}
	return false
}

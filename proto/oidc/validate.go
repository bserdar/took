package oidc

import (
	"time"

	log "github.com/sirupsen/logrus"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

// Validate checks if a token is valid
func Validate(accessToken, serverUrl string) bool {
	token, e := jwt.ParseSigned(accessToken)
	log.Debugf("Validation error: %v", e)
	if token != nil {
		svc, err := GetServerData(serverUrl)
		if err != nil {
			log.Fatalf("Cannot get server info for %s: %s", serverUrl, err)
			return false
		}
		claims := jwt.Claims{}
		token.Claims(svc.PK, &claims)
		log.Debugf("Token: %+v claims: %+v", token, claims)
		log.Debugf("Expiration: %d now: %d", claims.Expiry, time.Now().Unix())
		e := claims.ValidateWithLeeway(jwt.Expected{Time: time.Now()}, -30*time.Second)
		if e == nil {
			log.Debug("Token is valid")
			return true
		}
	}
	return false
}

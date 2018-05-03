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
			return "", err
		}
		claims := jwt.Claims{}
		token.Claims(svc.PK, &claims)
		log.Debugf("Token: %v claims: %v", token, claims)
		e := claims.Validate(jwt.Expected{Time: time.Now()})
		if e == nil {
			log.Debug("Token is valid")
			return true
		}
	}
	return false
}

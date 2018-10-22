package proto

import (
	"crypto/tls"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

// InsecureTLS set to true means TLS calls won't check certs
var InsecureTLS = false

// GetHTTPClient returns an HTTP client instance based on configuration
func GetHTTPClient() *http.Client {
	if InsecureTLS {
		return &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	}
	return &http.Client{}
}

// HTTPGet executes a GET using the HTTP client obtained from GetHTTPClient
func HTTPGet(url string) (*http.Response, error) {
	return GetHTTPClient().Get(url)
}

// HTTPPostForm posts form
func HTTPPostForm(url string, data url.Values) (*http.Response, error) {
	log.Debugf("Post %s %s", url, data.Encode())
	return GetHTTPClient().PostForm(url, data)
}

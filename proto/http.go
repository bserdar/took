package proto

import (
	"crypto/tls"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

// InsecureTLS set to true means TLS calls won't check certs
var InsecureTLS = false

// GetHTTPClient is initializes to DefaultGetHTTPClient
var GetHTTPClient = DefaultGetHTTPClient

// HTTPGet is initialized to DefaultHTTPGet
var HTTPGet = DefaultHTTPGet

// HTTPPostForm is initialized to DefaultHTTPPostForm
var HTTPPostForm = DefaultHTTPPostForm

// DefaultGetHTTPClient returns an HTTP client instance based on configuration
func DefaultGetHTTPClient() *http.Client {
	if InsecureTLS {
		return &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	}
	return &http.Client{}
}

// DefaultHTTPGet executes a GET using the HTTP client obtained from GetHTTPClient
func DefaultHTTPGet(url string) (*http.Response, error) {
	return GetHTTPClient().Get(url)
}

// DefaultHTTPPostForm posts form
func DefaultHTTPPostForm(url string, data url.Values) (*http.Response, error) {
	log.Debugf("Post %s %s", url, data.Encode())
	return GetHTTPClient().PostForm(url, data)
}

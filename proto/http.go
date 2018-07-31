package proto

import (
	"crypto/tls"
	"net/http"
	"net/url"
)

// If InsecureTLS is set to true, TLS calls won't check certs
var InsecureTLS = false

// GetHTTPClient retusn an HTTP client instance based on configuration
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

func HTTPPostForm(url string, data url.Values) (*http.Response, error) {
	return GetHTTPClient().PostForm(url, data)
}

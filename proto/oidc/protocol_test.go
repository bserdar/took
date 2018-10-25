package oidc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto"
)

type testReturn struct {
	returnCode int
	returnBody string
	headers    map[string]string
}

type testProtocolHandler struct {
	response        map[string]testReturn
	defaultResponse testReturn
}

func (h testProtocolHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("url: %s\n", req.URL)
	t, ok := h.response[req.URL.String()]
	if !ok {
		t = h.defaultResponse
	}

	for k, v := range t.headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(t.returnCode)
	w.Write([]byte(t.returnBody))
}

func TestGetToken_PasswordGrant(t *testing.T) {
	handler := testProtocolHandler{response: make(map[string]testReturn)}
	server := httptest.NewServer(&handler)
	defer server.Close()

	p := Protocol{}
	tr := true
	p.Cfg = Config{ServerProfile: ServerProfile{URL: server.URL, PasswordGrant: &tr},
		ClientID:     "id",
		ClientSecret: "secret",
		CallbackURL:  "http://callback"}

	cfg.AskPasswordWithPrompt = func(s string) string { return "pwd" }
	handler.response["/.well-known/openid-configuration"] =
		testReturn{returnCode: 200, returnBody: fmt.Sprintf(`{"authorization_endpoint":"%s/auth","token_endpoint":"%s/token","token_introspection_endpoint":"%s/verify"}`, server.URL, server.URL, server.URL)}

	handler.response["/token"] = testReturn{returnCode: 200, headers: map[string]string{"Content-Type": "application/json"}, returnBody: `{"access_token":"a","token_type":"bearer","refresh_token":"r"}`}

	ret, _, err := p.GetToken(proto.TokenRequest{Username: "user"})
	if err != nil {
		t.Errorf("Cannot get token: %v", err)
	}
	if ret != "a" {
		t.Errorf("Wrong token: %s", ret)
	}
}

func TestGetToken(t *testing.T) {
	handler := testProtocolHandler{response: make(map[string]testReturn)}
	server := httptest.NewServer(&handler)
	defer server.Close()

	p := Protocol{}
	p.Cfg = Config{ServerProfile: ServerProfile{URL: server.URL},
		ClientID:     "id",
		ClientSecret: "secret",
		CallbackURL:  "http://callback"}

	// This will be called with "Go to this url..."
	cfg.Ask = func(s string) string {
		// Get the http:// part
		ix := strings.Index(s, "http://")
		if ix == -1 {
			t.Errorf("Cannot find url in prompt")
			return ""
		}
		lf := strings.IndexRune(s, '\n')
		u, err := url.Parse(s[ix:lf])
		if err != nil {
			t.Errorf("Cannot parse url %s: %v", s[ix:lf], err)
			return ""
		}

		ret := fmt.Sprintf("http://callback?code=%s&state=%s", u.Query().Get("code"), u.Query().Get("state"))
		fmt.Printf("redirect: %s\n", ret)
		return ret
	}

	handler.response["/.well-known/openid-configuration"] =
		testReturn{returnCode: 200, returnBody: fmt.Sprintf(`{"authorization_endpoint":"%s/auth","token_endpoint":"%s/token","token_introspection_endpoint":"%s/verify"}`, server.URL, server.URL, server.URL)}

	handler.response["/token"] = testReturn{returnCode: 200, headers: map[string]string{"Content-Type": "application/json"}, returnBody: `{"access_token":"a","token_type":"bearer","refresh_token":"r"}`}

	ret, _, err := p.GetToken(proto.TokenRequest{Username: "user"})
	if err != nil {
		t.Errorf("Cannot get token: %v", err)
	}
	if ret != "a" {
		t.Errorf("Wrong token: %s", ret)
	}
}

func TestRefreshNeeded(t *testing.T) {
	handler := testProtocolHandler{response: make(map[string]testReturn)}
	server := httptest.NewServer(&handler)
	defer server.Close()

	p := Protocol{}
	p.Cfg = Config{ServerProfile: ServerProfile{URL: server.URL},
		ClientID:     "id",
		ClientSecret: "secret",
		CallbackURL:  "http://callback"}
	p.Tokens = Data{Last: "last",
		Tokens: []TokenData{{Username: "last", Type: "oidc", AccessToken: "a", RefreshToken: "r"}}}

	// This will be called with "Go to this url..."
	cfg.Ask = func(s string) string {
		// Get the http:// part
		ix := strings.Index(s, "http://")
		if ix == -1 {
			t.Errorf("Cannot find url in prompt")
			return ""
		}
		lf := strings.IndexRune(s, '\n')
		u, err := url.Parse(s[ix:lf])
		if err != nil {
			t.Errorf("Cannot parse url %s: %v", s[ix:lf], err)
			return ""
		}

		ret := fmt.Sprintf("http://callback?code=%s&state=%s", u.Query().Get("code"), u.Query().Get("state"))
		fmt.Printf("redirect: %s\n", ret)
		return ret
	}

	handler.response["/.well-known/openid-configuration"] =
		testReturn{returnCode: 200, returnBody: fmt.Sprintf(`{"authorization_endpoint":"%s/auth","token_endpoint":"%s/token","token_introspection_endpoint":"%s/verify"}`, server.URL, server.URL, server.URL)}

	handler.response["/verify"] = testReturn{returnCode: 200, headers: map[string]string{"Content-Type": "application/json"}, returnBody: `{"active":false}`}
	handler.response["/token"] = testReturn{returnCode: 200, headers: map[string]string{"Content-Type": "application/json"}, returnBody: `{"access_token":"a","token_type":"bearer","refresh_token":"r"}`}

	ret, _, err := p.GetToken(proto.TokenRequest{})
	if err != nil {
		t.Errorf("Cannot get token: %v", err)
	}
	if ret != "a" {
		t.Errorf("Wrong token: %s", ret)
	}
}

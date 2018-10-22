package oidc

import (
	"strings"
	"testing"

	yml "gopkg.in/yaml.v2"
)

func TestMarshal(t *testing.T) {
	p := ServerProfile{}
	data, _ := yml.Marshal(p)
	if strings.Contains(string(data), "url") {
		t.Errorf("Got %s", string(data))
	}
}

func TestMerge(t *testing.T) {
	tr := true
	p := ServerProfile{URL: "serverUrl",
		Insecure:      true,
		PasswordGrant: &tr}

	x := ServerProfile{URL: "",
		TokenAPI: "token",
		Insecure: false}
	x = x.Merge(p)

	if x.URL != p.URL ||
		!x.Insecure ||
		!*x.PasswordGrant ||
		x.TokenAPI != x.TokenAPI {
		t.Errorf("Got %+v", x)
	}
}

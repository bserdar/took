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

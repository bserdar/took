package oidc

import (
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto"
)

var multipleForms = `<html>
<body>
<form id="1" action="x">
<input type="text" name="field1" value="v1"/>
<input type="text" name="field3" value="v2"/>
</form>
<form id="2" action="x">
<input type="text" name="field1" value="v3"/>
<input type="text" name="field2" value="v4"/>
</form>
</body>`

var emptyForm = `<html>
<form id="2" action="http://action">
<input type="text" name="field1"/>
<input type="text" name="field2"/>
</form>
</body>`

func TestItrForms(t *testing.T) {
	node, err := html.Parse(strings.NewReader(multipleForms))
	if err != nil {
		t.Error(err)
		return
	}
	_, values := itrForms("", []string{"field1", "field2"}, node)
	if values.Get("field1") != "v3" ||
		values.Get("field2") != "v4" {
		t.Errorf("Wrong field value: %v", values)
	}
}

func TestFillForm(t *testing.T) {
	config := HTMLFormConfig{PasswordField: "field1", UsernameField: "field2",
		Fields: []FieldConfig{{Input: "field1", Password: true},
			{Input: "field2"}}}

	node, err := html.Parse(strings.NewReader(multipleForms))
	if err != nil {
		t.Error(err)
		return
	}
	action, values, err := FillForm(config, node, "user", "pass")
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if action != "x" {
		t.Errorf("Expection action x, got :%s", action)
	}
	if values.Get("field1") != "pass" ||
		values.Get("field2") != "user" {
		t.Errorf("Wrong values: %v", values)
	}

	node, err = html.Parse(strings.NewReader(emptyForm))
	if err != nil {
		t.Error(err)
		return
	}
	action, values, err = FillForm(config, node, "user", "pass")
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if values.Get("field1") != "pass" ||
		values.Get("field2") != "user" {
		t.Errorf("Wrong values: %v", values)
	}
}

type testFormAuthHandler struct {
	returnCode int
	returnBody string
	headers    map[string]string
}

func (h testFormAuthHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for k, v := range h.headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(h.returnCode)
	w.Write([]byte(h.returnBody))
}

func TestFormAuth(t *testing.T) {
	config := HTMLFormConfig{PasswordField: "field1", UsernameField: "field2",
		Fields: []FieldConfig{{Input: "field1", Password: true},
			{Input: "field2"}}}
	handler := testFormAuthHandler{headers: make(map[string]string)}
	server := httptest.NewServer(&handler)
	defer server.Close()

	proto.HTTPGet = func(url string) (*http.Response, error) {
		return &http.Response{Status: "200 OK",
			StatusCode: 200,
			Body:       ioutil.NopCloser(strings.NewReader(strings.Replace(emptyForm, "http://action", server.URL, -1)))}, nil
	}
	cfg.AskPasswordWithPrompt = func(prompt string) string { return "pass" }

	handler.returnCode = http.StatusMovedPermanently
	handler.headers["Location"] = "http://redirect"

	u := FormAuth(config, server.URL, "", "")
	proto.HTTPGet = proto.DefaultHTTPGet
	if u.String() != "http://redirect" {
		t.Errorf("Wrong redirect: %v", u)
	}
}

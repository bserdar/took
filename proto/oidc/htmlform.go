package oidc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto"
)

// HTMLFormConfig defines an HTP login form structure
type HTMLFormConfig struct {
	// Form ID
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	// Which field in Fields is the password field
	PasswordField string `json:"passwordField,omitempty" yaml:"passwordField,omitempty"`
	// Which field in Fields is the username field
	UsernameField string        `json:"usernameField,omitempty" yaml:"usernameField,omitempty"`
	Fields        []FieldConfig `json:"fields,omitempty" yaml:"fields,omitempty"`
}

// FieldConfig describes an HTML field in the HTML form
type FieldConfig struct {
	Input string `json:"input" yaml:"input"`
	// If non-empty, will ask for value
	Prompt   string `json:"prompt,omitempty" yaml:"prompt,omitempty"`
	Password bool   `json:"password" yaml:"password"`
	// If non-empty, the default value
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

// ReadPage reads the contents of the page
func ReadPage(url string) (*html.Node, []*http.Cookie, error) {
	resp, err := proto.HTTPGet(url)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	node, err := html.Parse(resp.Body)
	return node, resp.Cookies(), err
}

func findAttr(attr string, n *html.Node) string {
	for _, a := range n.Attr {
		if strings.ToLower(a.Key) == attr {
			return a.Val
		}
	}
	return ""
}

func getFields(node *html.Node) url.Values {
	ret := url.Values{}
	if node.Type == html.ElementNode && node.DataAtom == atom.Input {
		inputType := findAttr("type", node)
		name := findAttr("name", node)
		value := findAttr("value", node)
		if inputType == "text" || inputType == "password" || inputType == "hidden" {
			ret.Set(name, value)
		}
	}
	merge := func(m url.Values) {
		for k, v := range m {
			ret.Set(k, v[0])
		}
	}
	if node.FirstChild != nil {
		merge(getFields(node.FirstChild))
	}
	for trc := node.NextSibling; trc != nil; trc = trc.NextSibling {
		merge(getFields(trc))
	}
	return ret
}

func containsAll(fields url.Values, required []string) bool {
	for _, x := range required {
		if _, ok := fields[x]; !ok {
			return false
		}
	}
	return true
}

func itrForms(formID string, requiredFields []string, node *html.Node) (string, url.Values) {
	if node.Type == html.ElementNode && node.DataAtom == atom.Form {
		if len(formID) == 0 || findAttr("id", node) == formID {
			// Find the first form with these fields
			action := findAttr("action", node)
			fields := getFields(node.FirstChild)
			if len(formID) > 0 || containsAll(fields, requiredFields) {
				return action, fields
			}
		}
	}
	if node.FirstChild != nil {
		s, f := itrForms(formID, requiredFields, node.FirstChild)
		if s != "" {
			return s, f
		}
	}
	for trc := node.NextSibling; trc != nil; trc = trc.NextSibling {
		s, f := itrForms(formID, requiredFields, trc)
		if s != "" {
			return s, f
		}
	}
	return "", nil
}

// FillForm processes the form, prompts the user for field values, and returns the form to be submitted
func FillForm(config HTMLFormConfig, page *html.Node, userName, password string) (action string, values url.Values, err error) {
	requiredFields := make([]string, 0)
	for _, f := range config.Fields {
		requiredFields = append(requiredFields, f.Input)
	}
	action, values = itrForms(config.ID, requiredFields, page)
	if action != "" {
		for _, field := range config.Fields {
			if field.Input == config.UsernameField && len(userName) > 0 {
				// This is the username
				values.Set(field.Input, userName)
			} else if field.Input == config.PasswordField && len(password) > 0 {
				// This is the password
				values.Set(field.Input, password)
			} else {
				ask := cfg.Ask
				if field.Password {
					ask = cfg.AskPasswordWithPrompt
				}
				if field.Prompt != "" {
					if field.Value == "" {
						var v string
						v = ask(fmt.Sprintf("%s:", field.Prompt))
						values.Set(field.Input, v)
					} else {
						defaultValue := field.Value
						if field.Password {
							defaultValue = "***"
						}
						var val string
						val = ask(fmt.Sprintf("%s (%s):", field.Prompt, defaultValue))
						if len(val) == 0 {
							val = field.Value
						}
						values.Set(field.Input, val)
					}
				} else {
					values.Set(field.Input, field.Value)
				}
			}
		}
	}
	return action, values, nil
}

// FormAuth retrieves a login form from the authURL, parses it, asks
// credentials, submits the form, and if everything goes fine, returns
// the redirect URL
func FormAuth(cfg HTMLFormConfig, authURL string, userName, password string) *url.URL {
	var redirectedURL *url.URL
	log.Debugf("Reading login page at %s", authURL)
	node, cookies, err := ReadPage(authURL)
	if err == nil && node != nil {
		action, values, err := FillForm(cfg, node, userName, password)
		log.Debugf("action=%s err=%s", action, err)
		if err == nil && action != "" && values != nil {
			formData := values.Encode()
			request, _ := http.NewRequest(http.MethodPost, action, ioutil.NopCloser(strings.NewReader(formData)))
			for _, c := range cookies {
				request.AddCookie(c)
			}
			log.Debugf("posting...")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.ContentLength = int64(len(formData))
			cli := proto.GetHTTPClient()
			cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				redirectedURL = req.URL
				return errors.New("Redirect")
			}
			response, err := cli.Do(request)
			log.Debugf("Response: %+v %v", response, err)
			if err == nil {
				defer response.Body.Close()
			}
		}
	} else {
		log.Debugf("err:%s", err)
	}
	return redirectedURL
}

package oidc

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/bserdar/took/proto"
)

type HTMLFormConfig struct {
	ID     string
	Fields []FieldConfig
}

type FieldConfig struct {
	InputName string
	// If non-empty, will ask for value
	Prompt   string
	Password bool
	// If non-empty, the default value
	Value string
}

// ReadPage reads the contents of the page
func ReadPage(url string) (*html.Node, []*http.Cookie, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	node, err := html.Parse(resp.Body)
	return node, resp.Cookies(), err
}

const (
	st_seekingForm int = iota
	st_inForm
)

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
			fields := getFields(node)
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
func FillForm(cfg HTMLFormConfig, page *html.Node) (action string, values url.Values, err error) {
	requiredFields := make([]string, 0)
	for _, f := range cfg.Fields {
		requiredFields = append(requiredFields, f.InputName)
	}
	action, values = itrForms(cfg.ID, requiredFields, page)
	if action != "" {
		for _, field := range cfg.Fields {
			ask := proto.Ask
			if field.Password {
				ask = proto.AskPasswordWithPrompt
			}
			if field.Prompt != "" {
				if field.Value == "" {
					var v string
					v, err = ask(fmt.Sprintf("%s:", field.Prompt))
					if err != nil {
						return
					}
					values.Set(field.InputName, v)
				} else {
					defaultValue := field.Value
					if field.Password {
						defaultValue = "***"
					}
					var val string
					val, err = ask(fmt.Sprintf("%s (%s):", field.Prompt, defaultValue))
					if err != nil {
						return
					}
					if len(val) == 0 {
						val = field.Value
					}
					values.Set(field.InputName, val)
				}
			} else {
				values.Set(field.InputName, field.Value)
			}
		}
	}
	return action, values, nil
}

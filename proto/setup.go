package proto

import (
	"fmt"
	"strings"
)

type SetupStep struct {
	Prompt       string
	DefaultValue string
	// This will be called only if remoteCfg is non-nill
	GetDefault func(remoteCfg interface{}) string
	Parse      func(string) error
}

func (s SetupStep) Run(remoteCfg interface{}) {
	var def string

	if remoteCfg != nil {
		if s.GetDefault != nil {
			def = s.GetDefault(remoteCfg)
		}
	}
	if len(def) == 0 {
		def = s.DefaultValue
	}

retry:
	var value string
	if len(def) != 0 {
		var prompt string
		if strings.HasSuffix(s.Prompt, ":") {
			prompt = s.Prompt[:len(s.Prompt)-1]
		} else {
			prompt = s.Prompt
		}
		value = Ask(fmt.Sprintf("%s (%s):", prompt, def))
	} else {
		value = Ask(s.Prompt)
	}
	if len(value) == 0 {
		value = def
	}
	err := s.Parse(value)
	if err != nil {
		fmt.Println(err)
		goto retry
	}
}

package cmd

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto"
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
		value = proto.Ask(fmt.Sprintf("%s (%s):", prompt, def))
	} else {
		value = proto.Ask(s.Prompt)
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

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup a new authentication configuration",
	Long: `Setup a new authentication configuration. 

 took setup [newName [protocol]]
`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var cfgName string
		commonCfg := cfg.ReadCommonConfig()

		askName := func() {
			if len(commonCfg.Remotes) > 0 {
				buf := bytes.Buffer{}
				buf.WriteString("These are the known configurations:\n")
				for k := range commonCfg.Remotes {
					buf.WriteString(fmt.Sprintf("%s\n", k))
				}
				buf.WriteString("Enter configuration to add a new entry, or enter a new configuration name:")
				cfgName = proto.Ask(buf.String())
			} else {
				cfgName = proto.Ask("Enter new configuration name:")
			}
		}

		if len(args) > 0 {
			cfgName = args[0]
		} else {
			askName()
		}

	recheckName:
		if _, ok := UserCfg.Remotes[cfgName]; ok {
			fmt.Printf("%s already exists\n", cfgName)
			askName()
			goto recheckName
		}

		var protocolName string
		if len(args) > 1 {
			protocolName = args[1]
		} else {
			protocols := proto.ProtocolNames()
			if len(protocols) > 1 {
				protocolName = proto.Ask(fmt.Sprintf("Protocol (supported values: %v):", protocols))
			} else {
				protocolName = protocols[0]
			}
		}

		protocol := proto.Get(protocolName)
		if protocol == nil {
			panic("Invalid protocol")
		}

		var rmt interface{}
		if commonCfg.Remotes != nil {
			if r, ok := commonCfg.Remotes[cfgName]; ok {
				if r.Configuration != nil {
					defaults := protocol.GetConfigDefaultsInstance()
					err := mapstructure.Decode(r.Configuration, defaults)
					if err != nil {
						log.Fatalf("Error reading common configuration: %s", err)
					}
					rmt = defaults
				}
			}
		}

		var wiz []SetupStep
		var command *cobra.Command
		switch protocolName {
		case "oidc-auth", "oidc":
			wiz = oidcConnectWizard
			command = oidcConnectCmd
			oidcCfg.Name = cfgName
		}
		if wiz == nil {
			panic("unknown protocol")
		}
		for _, s := range wiz {
			s.Run(rmt)
		}
		command.Run(command, []string{})
	}}

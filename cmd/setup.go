package cmd

import (
	"bytes"
	"fmt"
	"log"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto"
)

func init() {
	RootCmd.AddCommand(setupCmd)
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

		wiz, command := protocol.InitSetupWizard(cfgName)
		if wiz == nil {
			panic("unknown protocol")
		}
		for _, s := range wiz {
			s.Run(rmt)
		}
		command.Run(command, []string{})
	}}

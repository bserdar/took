package cmd

import (
	"bytes"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto"
)

func init() {
	RootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup a new authentication configuration based on a server profile",
	Long: `Setup a new authentication configuration based on a server profile.

`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg.DecryptUserConfig()
		var serverProfileName string
	askProfile:
		if len(cfg.CommonCfg.ServerProfiles)+len(cfg.UserCfg.ServerProfiles) > 0 {
			buf := bytes.Buffer{}
			buf.WriteString("These are the known server profiles:\n")
			for k, v := range cfg.CommonCfg.ServerProfiles {
				buf.WriteString(fmt.Sprintf("%s (%s)\n", k, v.Type))
			}
			for k, v := range cfg.UserCfg.ServerProfiles {
				buf.WriteString(fmt.Sprintf("%s (%s)\n", k, v.Type))
			}
			buf.WriteString("Enter the server profile for which you want to add a new authentication configuration:")
			serverProfileName = proto.Ask(buf.String())
		} else {
			log.Fatalf("There are no known server profiles")
		}

		serverProfile := cfg.GetServerProfile(serverProfileName)
		if len(serverProfile.Type) == 0 {
			goto askProfile
		}

	askName:
		cfgName := proto.Ask("Enter name of the new authentication configuration:")
		if _, ok := cfg.UserCfg.Remotes[cfgName]; ok {
			fmt.Printf("%s already exists\n", cfgName)
			goto askName
		}

		protocol := proto.Get(serverProfile.Type)
		if protocol == nil {
			panic("Invalid protocol")
		}

		var rmt interface{}
		if cfg.CommonCfg.Remotes != nil {
			if r, ok := cfg.CommonCfg.Remotes[cfgName]; ok {
				if r.Configuration != nil {
					defaults, err := protocol.DecodeCfg(r.Configuration)
					if err != nil {
						log.Fatalf("Error reading common configuration: %s", err)
					}
					rmt = defaults
				}
			}
		}

		wiz, command := protocol.InitSetupWizard(cfgName, serverProfileName, serverProfile)
		if wiz == nil {
			panic("unknown protocol")
		}
		for _, s := range wiz {
			s.Run(rmt)
		}
		command.Run(command, []string{})
	}}

package cmd

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto"
)

func init() {
	RootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup [profile]",
	Short: "Setup a new authentication configuration based on a server profile",
	Long: `Setup a new authentication configuration based on a server profile.

`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		InitConfig()
		cfg.DecryptUserConfig(cfg.UserCfgFile)
		var serverProfileName string
		profileArg := ""
		if len(args) > 0 {
			profileArg = args[0]
		}
	askProfile:
		if len(profileArg) == 0 {
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
				serverProfileName = cfg.Ask(buf.String())
			} else {
				log.Fatalf("There are no known server profiles")
			}
		} else {
			serverProfileName = profileArg
		}

		serverProfile := cfg.GetServerProfile(serverProfileName)
		if len(serverProfile.Type) == 0 {
			fmt.Printf("%s not found\n", serverProfileName)
			profileArg = ""
			goto askProfile
		}

	askName:
		cfgName := strings.TrimSpace(cfg.Ask("Enter name of the new authentication configuration:"))
		if _, ok := cfg.UserCfg.Remotes[cfgName]; ok {
			fmt.Printf("%s already exists\n", cfgName)
			goto askName
		}

		protocol := proto.Get(serverProfile.Type)
		if protocol == nil {
			log.Fatal("Invalid protocol")
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
			log.Fatal("unknown protocol")
		}
		for _, s := range wiz {
			s.Run(rmt)
		}
		command.Run(command, []string{})
	}}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
)

func init() {
	RootCmd.AddCommand(remotesCmd)
	RootCmd.AddCommand(profilesCmd)
}

var remotesCmd = &cobra.Command{
	Use:   "cfg",
	Short: "List authentication configurations",
	Long:  `List authentication configurations`,
	Run: func(cmd *cobra.Command, args []string) {
		InitConfig()
		for k, v := range cfg.UserCfg.Remotes {
			fmt.Printf("%s (%s)\n", k, v.Type)
		}
		for k, v := range cfg.CommonCfg.Remotes {
			if _, ok := cfg.UserCfg.Remotes[k]; !ok {
				fmt.Printf("%s (%s)\n", k, v.Type)
			}
		}
	}}

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List server profiles",
	Long:  `List server profiles`,
	Run: func(cmd *cobra.Command, args []string) {
		InitConfig()
		for k, v := range cfg.UserCfg.ServerProfiles {
			fmt.Printf("%s (%s)\n", k, v.Type)
		}
		for k, v := range cfg.CommonCfg.ServerProfiles {
			if _, ok := cfg.UserCfg.ServerProfiles[k]; !ok {
				fmt.Printf("%s (%s)\n", k, v.Type)
			}
		}
	}}

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
)

var forceDelete bool

func init() {
	RootCmd.AddCommand(DeleteCmd)
	DeleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Delete without asking")
}

// DeleteCmd is the took delete command
var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an authentication configuration",
	Long:  `Delete an authentication configuration`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		InitConfig()
		cfg.DecryptUserConfig()
		_, ok := cfg.UserCfg.Remotes[args[0]]
		if !ok {
			log.Fatalf("%s not found", args[0])
		}
		if !forceDelete {
			answer := cfg.Ask(fmt.Sprintf("Delete %s (y/N): ", args[0]))
			if answer != "y" && answer != "N" {
				return
			}
		}
		delete(cfg.UserCfg.Remotes, args[0])
		WriteUserConfig()
	}}

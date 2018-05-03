package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new authentication configuration",
	Long:  `Add a new authentication configuration`}

func remoteExists(name string) bool {
	return viper.IsSet(fmt.Sprintf("remotes.%s", name))
}

func verifyRemoteUnique(name string) {
	if remoteExists(name) {
		log.Fatalf("Remote %s already exists", name)
	}
}

func setRemoteType(name, t string) {
	viper.Set(fmt.Sprintf("remotes.%s.type", name), t)
}

func setRemoteConfig(name string, config interface{}) {
	viper.Set(fmt.Sprintf("remotes.%s.cfg", name), config)
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(VerCmd)
}

// VerCmd is the took version command
var VerCmd = &cobra.Command{
	Use:   "version",
	Short: "Show took version",
	Long:  `Show took version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	}}

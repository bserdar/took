package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(AddCmd)
	RootCmd.AddCommand(ModCmd)
}

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new authentication configuration",
	Long:  `Add a new authentication configuration`}

var ModCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing configuration",
	Long:  "Update an existing configuration"}

package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(modCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new authentication configuration",
	Long:  `Add a new authentication configuration`}

var modCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing configuration",
	Long:  "Update an existing configuration"}

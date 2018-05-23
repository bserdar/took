package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto/oidc"
)

type oidcConnectDA struct {
	Name string
	Cfg  oidc.Config
}

var oidcConnectDACfg oidcConnectDA

func init() {
	addCmd.AddCommand(oidcConnectDACmd)
	modCmd.AddCommand(oidcConnectDAUpdateCmd)

	doFlags := func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&oidcConnectDACfg.Name, "name", "n", "", "Name of the configuration (required)")
		cmd.MarkFlagRequired("name")

		cmd.Flags().StringVarP(&oidcConnectDACfg.Cfg.ClientId, "clientId", "c", "", "Client ID (required)")
		cmd.MarkFlagRequired("clientId")
		cmd.Flags().StringVarP(&oidcConnectDACfg.Cfg.ClientSecret, "secret", "s", "", "Client Secret (required)")
		cmd.MarkFlagRequired("secret")
		cmd.Flags().StringVarP(&oidcConnectDACfg.Cfg.URL, "url", "u", "", "Server URL (required)")
		cmd.MarkFlagRequired("url")
		cmd.Flags().StringVarP(&oidcConnectDACfg.Cfg.TokenAPI, "token-api", "a", "", "Token API (defaults to protocol/openid-connect/token)")
	}
	doFlags(oidcConnectDACmd)
	doFlags(oidcConnectDAUpdateCmd)
}

var oidcConnectDACmd = &cobra.Command{
	Use:   "oidc-direct-access",
	Short: "Add a new oidc-direct-access configuration",
	Long:  `Add a new oidc-direct-access configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, ok := UserCfg.Remotes[oidcConnectDACfg.Name]; ok {
			log.Fatalf("Remote %s already exists", oidcConnectDACfg.Name)
		}
		UserCfg.Remotes[oidcConnectDACfg.Name] = cfg.Remote{Type: "oidc-direct-access", Configuration: oidcConnectDACfg.Cfg}
		writeUserConfig()
	}}

var oidcConnectDAUpdateCmd = &cobra.Command{
	Use:   "oidc-direct-access",
	Short: "Add a new oidc-direct-access configuration",
	Long:  `Add a new oidc-direct-access configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, ok := UserCfg.Remotes[oidcConnectDACfg.Name]; !ok {
			log.Fatalf("Remote %s does not exist", oidcConnectDACfg.Name)
		}
		UserCfg.Remotes[oidcConnectDACfg.Name] = cfg.Remote{Type: "oidc-direct-access", Configuration: oidcConnectDACfg.Cfg}
		writeUserConfig()
	}}

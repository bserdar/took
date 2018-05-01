package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	oidc "github.com/bserdar/took/proto/oidc_da"
)

type oidcConnectDA struct {
	Name string
	Cfg  oidc.Config
}

var oidcConnectDACfg oidcConnectDA

func init() {
	addCmd.AddCommand(oidcConnectDACmd)

	oidcConnectDACmd.Flags().StringVarP(&oidcConnectDACfg.Name, "name", "n", "", "Name of the configuration (required)")
	oidcConnectDACmd.MarkFlagRequired("name")

	oidcConnectDACmd.Flags().StringVarP(&oidcConnectDACfg.Cfg.ClientId, "clientId", "c", "", "Client ID (required)")
	oidcConnectDACmd.MarkFlagRequired("clientId")
	oidcConnectDACmd.Flags().StringVarP(&oidcConnectDACfg.Cfg.ClientSecret, "secret", "s", "", "Client Secret (required)")
	oidcConnectDACmd.MarkFlagRequired("secret")
	oidcConnectDACmd.Flags().StringVarP(&oidcConnectDACfg.Cfg.URL, "url", "u", "", "Server URL (required)")
	oidcConnectDACmd.MarkFlagRequired("url")
	oidcConnectDACmd.Flags().StringVarP(&oidcConnectDACfg.Cfg.TokenAPI, "token-api", "a", "", "Token API (defaults to protocol/openid-connect/token)")
}

var oidcConnectDACmd = &cobra.Command{
	Use:   "oidc-direct-access",
	Short: "Add a new oidc-direct-access configuration",
	Long:  `Add a new oidc-direct-access configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.IsSet(fmt.Sprintf("remotes.%s", oidcConnectDACfg.Name)) {
			log.Fatalf("Remote %s already exists", oidcConnectDACfg.Name)
		}
		viper.Set(fmt.Sprintf("remotes.%s.type", oidcConnectDACfg.Name), "oidc-direct-access")
		viper.Set(fmt.Sprintf("remotes.%s.cfg", oidcConnectDACfg.Name), oidcConnectDACfg.Cfg)
		e := viper.WriteConfig()
		if e != nil {
			panic(e)
		}
	}}

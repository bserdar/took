package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto/oidc"
)

type oidcConnect struct {
	Name string
	Cfg  oidc.Config
}

var oidcCfg oidcConnect

func init() {
	addCmd.AddCommand(oidcConnectCmd)

	oidcConnectCmd.Flags().StringVarP(&oidcCfg.Name, "name", "n", "", "Name of the configuration (required)")
	oidcConnectCmd.MarkFlagRequired("name")

	oidcConnectCmd.Flags().StringVarP(&oidcCfg.Cfg.ClientId, "clientId", "c", "", "Client ID (required)")
	oidcConnectCmd.MarkFlagRequired("clientId")
	oidcConnectCmd.Flags().StringVarP(&oidcCfg.Cfg.ClientSecret, "secret", "s", "", "Client Secret (required)")
	oidcConnectCmd.MarkFlagRequired("secret")
	oidcConnectCmd.Flags().StringVarP(&oidcCfg.Cfg.URL, "url", "u", "", "Server URL (required)")
	oidcConnectCmd.MarkFlagRequired("url")
	oidcConnectCmd.Flags().StringVarP(&oidcCfg.Cfg.CallbackURL, "callback-url", "b", "", "Callback URL (required)")
	oidcConnectCmd.MarkFlagRequired("callback-url")
	oidcConnectCmd.Flags().StringVarP(&oidcCfg.Cfg.TokenAPI, "token-api", "a", "", "Token API (defaults to protocol/openid-connect/token)")
	oidcConnectCmd.Flags().StringVarP(&oidcCfg.Cfg.AuthAPI, "auth-api", "t", "", "Auth API (defaults to protocol/openid-connect/auth)")
}

var oidcConnectCmd = &cobra.Command{
	Use:   "oidc",
	Short: "Add a new oidc configuration using authorization code flow",
	Long:  `Add a new oidc configuration using authorization code flow`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, ok := UserCfg.Remotes[oidcCfg.Name]; ok {
			log.Fatalf("Remote %s already exists", oidcCfg.Name)
		}
		UserCfg.Remotes[oidcCfg.Name] = cfg.Remote{Type: "oidc-auth", Configuration: oidcCfg.Cfg}
		writeUserConfig()
	}}

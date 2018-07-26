package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto/oidc"
)

type oidcConnect struct {
	Name string
	Cfg  oidc.Config
	form string
}

var oidcCfg oidcConnect

var oidcConnectWizard = []SetupStep{
	{Prompt: "Server URL:", Parse: func(in string) error {
		if len(in) == 0 {
			return fmt.Errorf("Server URL is required")
		}
		oidcCfg.Cfg.URL = in
		return nil
	}, GetDefault: func(remoteCfg interface{}) string { return remoteCfg.(*oidc.Config).URL }},
	{Prompt: "Client ID:", Parse: func(in string) error {
		if len(in) == 0 {
			return fmt.Errorf("Client id is required")
		}
		oidcCfg.Cfg.ClientId = in
		return nil
	}, GetDefault: func(remoteCfg interface{}) string { return remoteCfg.(*oidc.Config).ClientId }},
	{Prompt: "Client secret:", Parse: func(in string) error {
		if len(in) == 0 {
			return fmt.Errorf("Client secret is required")
		}
		oidcCfg.Cfg.ClientSecret = in
		return nil
	}, GetDefault: func(remoteCfg interface{}) string { return remoteCfg.(*oidc.Config).ClientSecret }},
	{Prompt: "Callback URL:", Parse: func(in string) error {
		if len(in) == 0 {
			return fmt.Errorf("Callback URL is required")
		}
		oidcCfg.Cfg.CallbackURL = in
		return nil
	}, GetDefault: func(remoteCfg interface{}) string { return remoteCfg.(*oidc.Config).CallbackURL }}}

func init() {
	addCmd.AddCommand(oidcConnectCmd)
	modCmd.AddCommand(oidcConnectUpdateCmd)

	doFlags := func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&oidcCfg.Name, "name", "n", "", "Name of the configuration (required)")
		cmd.MarkFlagRequired("name")

		cmd.Flags().StringVarP(&oidcCfg.Cfg.ClientId, "clientId", "c", "", "Client ID (required)")
		cmd.MarkFlagRequired("clientId")
		cmd.Flags().StringVarP(&oidcCfg.Cfg.ClientSecret, "secret", "s", "", "Client Secret (required)")
		cmd.MarkFlagRequired("secret")
		cmd.Flags().StringVarP(&oidcCfg.Cfg.URL, "url", "u", "", "Server URL (required)")
		cmd.MarkFlagRequired("url")
		cmd.Flags().StringVarP(&oidcCfg.Cfg.CallbackURL, "callback-url", "b", "", "Callback URL (required)")
		cmd.MarkFlagRequired("callback-url")
		cmd.Flags().StringVarP(&oidcCfg.Cfg.TokenAPI, "token-api", "a", "", "Token API (defaults to protocol/openid-connect/token)")
		cmd.Flags().StringVarP(&oidcCfg.Cfg.AuthAPI, "auth-api", "t", "", "Auth API (defaults to protocol/openid-connect/auth)")
		cmd.Flags().BoolVarP(&oidcCfg.Cfg.PasswordGrant, "pwd", "p", false, "Password grant")
		cmd.Flags().StringVarP(&oidcCfg.form, "form", "F", "", `Login form parameters, json document
  { "id":<formId>,
    "username": <name of the fields[] element for username>,
    "password": <name of the fields[] element for password>,
    "fields": [
      {
        "input":<Name of the input field>,
        "prompt": <String to prompt. If empty, value of the field in form is used>,
        "password": true|false,
        "value": <default value, omit field if none>
      }
    ]
  }`)
	}

	doFlags(oidcConnectCmd)
	doFlags(oidcConnectUpdateCmd)
}

var oidcConnectUpdateCmd = &cobra.Command{
	Use:   "oidc",
	Short: "Update an oidc configuration",
	Long:  `Update an oidc configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, ok := UserCfg.Remotes[oidcCfg.Name]; !ok {
			log.Fatalf("Remote %s does not exist", oidcCfg.Name)
		}
		parseOidc()
	}}

var oidcConnectCmd = &cobra.Command{
	Use:   "oidc",
	Short: "Add a new oidc configuration using authorization code flow",
	Long:  `Add a new oidc configuration using authorization code flow`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, ok := UserCfg.Remotes[oidcCfg.Name]; ok {
			log.Fatalf("Remote %s already exists", oidcCfg.Name)
		}
		parseOidc()
	}}

func parseOidc() {
	var formCfg oidc.HTMLFormConfig
	if len(oidcCfg.form) > 0 {
		err := json.Unmarshal([]byte(oidcCfg.form), &formCfg)
		if err != nil {
			log.Fatal(err)
		}
		oidcCfg.Cfg.Form = &formCfg
	}
	UserCfg.Remotes[oidcCfg.Name] = cfg.Remote{Type: "oidc-auth", Configuration: oidcCfg.Cfg}
	writeUserConfig()
}

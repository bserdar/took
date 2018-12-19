package oidc

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/cmd"
	"github.com/bserdar/took/proto"
)

type oidcConnect struct {
	Name   string
	Cfg    Config
	form   string
	scopes string
	flow   string
}

var oidcCfg oidcConnect

var oidcConnectWizard = []proto.SetupStep{
	{Prompt: "Client ID:", Parse: func(in string) error {
		in = strings.TrimSpace(in)
		if len(in) == 0 {
			return fmt.Errorf("Client id is required")
		}
		oidcCfg.Cfg.ClientID = in
		return nil
	}, GetDefault: func(remoteCfg interface{}) string { return remoteCfg.(*Config).ClientID }},
	{Prompt: "Client secret:", Parse: func(in string) error {
		in = strings.TrimSpace(in)
		if len(in) == 0 {
			return fmt.Errorf("Client secret is required")
		}
		oidcCfg.Cfg.ClientSecret = in
		return nil
	}, GetDefault: func(remoteCfg interface{}) string { return remoteCfg.(*Config).ClientSecret }},
	{Prompt: "Callback URL:", Parse: func(in string) error {
		in = strings.TrimSpace(in)
		oidcCfg.Cfg.CallbackURL = in
		return nil
	}, GetDefault: func(remoteCfg interface{}) string { return remoteCfg.(*Config).CallbackURL }},
	{Prompt: "OIDC flow (auth - authorization code flow, pwd - password grant flow, leave empty to use server profile default):",
		Parse: func(in string) error {
			in = strings.TrimSpace(in)
			if in == "pwd" || in == "auth" || in == "" {
				oidcCfg.flow = in
			} else {
				return fmt.Errorf("Invalid entry %s, enter auth, pwd,or leave empty", in)
			}
			return nil
		}}}

func init() {
	//	oidcConnectCmd.SetUsageFunc(func(c *cobra.Command) error {
	//		c.OutOrStderr().Write([]byte(`required flags are "callback-url", "clientId", "name", "secret", and one of "server" or "url"`))
	//		return nil
	//	})
	cmd.AddCmd.AddCommand(oidcConnectCmd)
	cmd.ModCmd.AddCommand(oidcConnectUpdateCmd)

	doFlags := func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&oidcCfg.Name, "name", "n", "", "Name of the configuration (required)")
		cmd.MarkFlagRequired("name")

		cmd.Flags().StringVarP(&oidcCfg.Cfg.ClientID, "clientId", "c", "", "Client ID (required)")
		cmd.MarkFlagRequired("clientId")
		cmd.Flags().StringVarP(&oidcCfg.Cfg.ClientSecret, "secret", "s", "", "Client Secret (required)")
		cmd.MarkFlagRequired("secret")
		cmd.Flags().StringVarP(&oidcCfg.Cfg.CallbackURL, "callback-url", "b", "", "Callback URL")
		cmd.Flags().StringVarP(&oidcCfg.Cfg.Profile, "server", "e", "", "Server profile to use. Either a server profile or the server URL must be given")
		cmd.Flags().StringVarP(&oidcCfg.Cfg.URL, "url", "u", "", "Server URL. Either a server profile or server URL must be given")
		cmd.Flags().StringVarP(&oidcCfg.Cfg.TokenAPI, "token-api", "a", "", "Token API (defaults to protocol/openid-connect/token)")
		cmd.Flags().StringVarP(&oidcCfg.Cfg.AuthAPI, "auth-api", "t", "", "Auth API (defaults to protocol/openid-connect/auth)")
		cmd.Flags().StringVarP(&oidcCfg.scopes, "scopes", "o", "", "Additional scopes to request from server (-o scope1,scope2,scope3)")
		cmd.Flags().StringVarP(&oidcCfg.flow, "flow", "f", "auth", "Use authorization code flow (auth) or password grant flow (pwd)")
		if cfg.InsecureAllowed() {
			cmd.Flags().BoolVarP(&oidcCfg.Cfg.Insecure, "insecure", "k", false, "Do not validate server certificates")
		}
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
	Run: func(c *cobra.Command, args []string) {
		cfg.DecryptUserConfig()
		if _, ok := cfg.UserCfg.Remotes[oidcCfg.Name]; !ok {
			log.Fatalf("Remote %s does not exist", oidcCfg.Name)
		}
		parseOidc(c)
	}}

var oidcConnectCmd = &cobra.Command{
	Use:   "oidc",
	Short: "Add a new oidc configuration",
	Long: `Add a new oidc configuration.

You will need the following information to create an OIDC configuration. These are supplied by the API provider:

   ClientId/ClientSecret: This pair uniquely identifies your application to the authentication server
   CallbackURL: This is the CallbackURL you must configure when you signed up with the API provider.
                This URL does not have to be a valid website. Once authenticated, the authentication 
                server returns an HTTP redirect to this CallbackURL.  This redirect also contains 
                codes that you can exchange for tokens, or tokens themselves.
                Took never attempts to follow this redirection, it simply extracts authentication information 
                from the redirection response. However, if you use a browser to extract this information, make
                sure this URL does not redirect you somewhere else, otherwise you will not be able to get your
                authentication code.

You can use a predefined server profile to define authentication configurations:

   took add oidc -n <name> -e <profile> -c <clientId> -s <clientSecret> -b <callbackURL>

Or you can add configuration by defining the server URL and other server attributes:

   took add oidc -n <name> -u <serverURL> (other optional server arguments) -c <clientId> -s <clientSecret> -b <callbackURL> 

 `,
	Run: func(c *cobra.Command, args []string) {
		cfg.DecryptUserConfig()
		if _, ok := cfg.UserCfg.Remotes[oidcCfg.Name]; ok {
			log.Fatalf("Remote %s already exists", oidcCfg.Name)
		}
		parseOidc(c)
	}}

func parseOidc(c *cobra.Command) {
	// There must be either a valid server profile, or server URL
	if len(oidcCfg.Cfg.URL) == 0 {
		if len(oidcCfg.Cfg.Profile) == 0 {
			log.Fatal("Either a server URL or a server profile must be given")
		}
		profile := cfg.UserCfg.GetServerProfile(oidcCfg.Cfg.Profile)
		if len(profile.Type) > 0 {
			if profile.Type != "oidc" && profile.Type != "oidc-auth" {
				log.Fatal("Server profile is not for oidc")
			}
		} else {
			profile = cfg.CommonCfg.GetServerProfile(oidcCfg.Cfg.Profile)
			if len(profile.Type) > 0 {
				if profile.Type != "oidc" && profile.Type != "oidc-auth" {
					log.Fatal("Server profile is not for oidc")
				}
			}
		}
	}
	if oidcCfg.flow == "auth" {
		x := false
		oidcCfg.Cfg.PasswordGrant = &x
	} else if oidcCfg.flow == "pwd" {
		x := true
		oidcCfg.Cfg.PasswordGrant = &x
	} else if oidcCfg.flow != "" {
		log.Fatalf("Invalid flow: %s Use 'auth' or 'pwd'", oidcCfg.flow)
	}

	var formCfg HTMLFormConfig
	if len(oidcCfg.form) > 0 {
		err := json.Unmarshal([]byte(oidcCfg.form), &formCfg)
		if err != nil {
			log.Fatal(err)
		}
		oidcCfg.Cfg.Form = &formCfg
	}
	if len(oidcCfg.scopes) > 0 {
		oidcCfg.Cfg.AdditionalScopes = strings.Split(oidcCfg.scopes, ",")
	}
	cfg.UserCfg.Remotes[oidcCfg.Name] = cfg.Remote{Type: "oidc-auth", Configuration: oidcCfg.Cfg}
	cmd.WriteUserConfig()
}

// InitSetupWizard initializes the setup wizard for oidc
func (p *Protocol) InitSetupWizard(name string, profileName string, profile cfg.Profile) ([]proto.SetupStep, *cobra.Command) {
	oidcCfg.Name = name
	oidcCfg.Cfg.Profile = profileName
	return oidcConnectWizard, oidcConnectCmd
}

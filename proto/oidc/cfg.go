package oidc

// ServerProfile defines an OIDC auth server
type ServerProfile struct {
	URL              string
	TokenAPI         string
	AuthAPI          string
	Form             *HTMLFormConfig
	Insecure         bool
	PasswordGrant    *bool    `yaml:"passwordgrant,omitempty"`
	AdditionalScopes []string `yaml:"additionalscopes,omitempty"`
}

// Merge sets any unset field in s from in, and returns the merged copy
func (s ServerProfile) Merge(in ServerProfile) ServerProfile {
	ret := ServerProfile{URL: wdef(s.URL, in.URL),
		TokenAPI: wdef(s.TokenAPI, in.TokenAPI),
		AuthAPI:  wdef(s.AuthAPI, in.AuthAPI)}
	ret.Insecure = s.Insecure || in.Insecure
	ret.PasswordGrant = s.PasswordGrant
	if ret.PasswordGrant == nil {
		ret.PasswordGrant = in.PasswordGrant
	}
	ret.Form = s.Form
	if ret.Form == nil {
		ret.Form = in.Form
	}
	ret.AdditionalScopes = append(s.AdditionalScopes, in.AdditionalScopes...)

	return ret
}

// Config includes the server profile and contains user creds
type Config struct {
	ServerProfile `yaml:",inline" mapstructure:",squash"`
	Profile       string
	ClientID      string `yaml:"clientid" mapstructure:"clientid"`
	ClientSecret  string
	CallbackURL   string
}

// Merge sets the unset fields of c from defaults
func (c Config) Merge(defaults Config) Config {
	ret := Config{ClientID: wdef(c.ClientID, defaults.ClientID),
		ClientSecret: wdef(c.ClientSecret, defaults.ClientSecret),
		CallbackURL:  wdef(c.CallbackURL, defaults.CallbackURL)}
	ret.ServerProfile = c.ServerProfile.Merge(defaults.ServerProfile)
	return ret
}
func wdef(s, def string) string {
	if len(s) > 0 {
		return s
	}
	return def
}

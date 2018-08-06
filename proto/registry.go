package proto

import (
	"github.com/bserdar/took/cfg"

	"github.com/spf13/cobra"
)

type RefreshOption int

// Constants for how to get the token
const (
	UseDefault RefreshOption = iota
	UseRefresh
	UseReAuth
)

type OutputOption int

// Output options
const (
	OutputToken OutputOption = iota
	OutputHeader
)

type TokenRequest struct {
	Refresh  RefreshOption
	Out      OutputOption
	Username string
	Password string
}

// Protocol defines a protocol
type Protocol interface {

	// DecodeCfg decodes the input map configuration into protocol specific structure
	DecodeCfg(config interface{}) (interface{}, error)

	// SetCfg sets the configuration for the protocol instance. The
	// passed in userCfg and commonCfg can be nil
	SetCfg(userCfg, commonCfg cfg.Remote)

	// GetToken returns the token with the given configuration and
	// data blocks. Returns the new copy of data block for
	// configuration
	GetToken(TokenRequest) (string, interface{}, error)

	// InitSetupWizard should initialize the internal configuration to
	// setup configuration 'name', and return the setup steps and the
	// cobra command
	InitSetupWizard(name string) ([]SetupStep, *cobra.Command)
}

var protocols = make(map[string]func() Protocol)

// Register registers a protocol
func Register(name string, factory func() Protocol) {
	protocols[name] = factory
}

// Get retrieves a protocol by name
func Get(name string) Protocol {
	p, ok := protocols[name]
	if ok {
		return p()
	}
	return nil
}

// ProtocolNames returns supported protocol names
func ProtocolNames() []string {
	ret := make([]string, len(protocols))
	i := 0
	for k := range protocols {
		ret[i] = k
		i++
	}
	return ret
}

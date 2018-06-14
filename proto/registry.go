package proto

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
	// GetDataInstance returns a new data block into which the data
	// part of the configuration will be unmarshaled
	GetDataInstance() interface{}
	// GetConfigInstance returns a new configuration instance into which the configuration will be unmarshaled
	GetConfigInstance() interface{}
	// GetConfigDefaultInstance returns an instance of configuration into which defaults will be unmarshaled
	GetConfigDefaultsInstance() interface{}
	// GetToken returns the token with the given configuration and
	// data blocks. Returns the new copy of data block for
	// configuration
	GetToken(TokenRequest) (string, interface{}, error)
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

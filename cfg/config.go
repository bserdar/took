package cfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	yml "gopkg.in/yaml.v2"

	"github.com/bserdar/took/crypta"
	"github.com/bserdar/took/crypta/rpc"
)

// UserCfg is the user's config
var UserCfg Configuration

// CommonCfg is the common config from /etc/tool.yaml
var CommonCfg Configuration

// DefaultEncTimeout is 10 minutes, the default idle timeout for the took decrypt agent
var DefaultEncTimeout = 10 * time.Minute

// Configuration declares the structure of the config file
type Configuration struct {
	AuthKey        string             `yaml:"key,omitempty"`
	Remotes        map[string]Remote  `yaml:"remotes,omitempty"`
	ServerProfiles map[string]Profile `yaml:"serverProfiles,omitempty"`
}

// GetServerProfile returns a server profile by name. Returns empty profile if not found
func (c Configuration) GetServerProfile(name string) Profile {
	if len(name) > 0 {
		if c.ServerProfiles != nil {
			ret, _ := c.ServerProfiles[name]
			return ret
		}
	}
	return Profile{}
}

// GetServerProfile returns the server profile from user cfg. If there
// is none, then it returns the profile from common cfg. If common cfg
// does not have a profile, it'll return empty profile
func GetServerProfile(name string) Profile {
	p := UserCfg.GetServerProfile(name)
	if len(p.Type) > 0 {
		return p
	}
	return CommonCfg.GetServerProfile(name)
}

// Profile is combines the type and the type-specific configuration of a server profile
type Profile struct {
	Type          string      `yaml:"type"`
	Configuration interface{} `yaml:"cfg,omitempty"`
}

// Remote defines a remote auth configuration
type Remote struct {
	// Type is the auth protocol
	Type string `yaml:"type"`
	// Configuration is the prototocol specific configuration
	Configuration interface{} `yaml:"cfg,omitempty"`
	// Data contains the protocol specific token information
	Data interface{} `yaml:"data,omitempty"`
	// ECfg is the encrypted configuration. Only one of Configuration or Ecfg is nonempty
	ECfg string `yaml:"ecfg,omitempty"`
	// EData is the encrypted data. Only one of Data or EData is nonempty
	EData string `yaml:"edata,omitempty"`
}

// ReadConfig reads the cfgFile.
func ReadConfig(cfgFile string) Configuration {
	c, err := readConfig(cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func readConfig(cfgFile string) (Configuration, error) {
	f, err := os.Open(cfgFile)
	if err != nil {
		return Configuration{}, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return Configuration{}, err
	}
	var config Configuration
	err = yml.Unmarshal(data, &config)
	if err != nil {
		return Configuration{}, err
	}
	if config.Remotes == nil {
		config.Remotes = make(map[string]Remote)
	}

	return config, nil
}

func init() {
	CommonCfg = ReadCommonConfig()
}

// ReadCommonConfig reads the common configuration file under /etc/took.yaml
func ReadCommonConfig() Configuration {
	cfg, _ := readConfig("/etc/took.yaml")
	return cfg
}

func decrypt(cli *rpc.RequestProcessorClient, in string) map[string]interface{} {
	out, err := cli.Decrypt(in)
	if err != nil {
		log.Fatal(err)
	}
	var m map[string]interface{}
	err = json.Unmarshal([]byte(out), &m)
	if err != nil {
		log.Fatal(err)
	}
	return m
}

func encrypt(cli *rpc.RequestProcessorClient, in interface{}) (string, *rpc.RequestProcessorClient) {
	s, _ := json.Marshal(in)
	if cli == nil {
		cli = MustConnectEncServer()
	}
	ret, _ := cli.Encrypt(string(s))
	return ret, cli
}

func decryptRemote(cli *rpc.RequestProcessorClient, in Remote) Remote {
	if len(in.ECfg) > 0 {
		in.Configuration = decrypt(cli, in.ECfg)
		in.ECfg = ""
	}
	if len(in.EData) > 0 {
		in.Data = decrypt(cli, in.EData)
		in.EData = ""
	}
	return in
}

func encryptRemote(cli *rpc.RequestProcessorClient, in Remote) (Remote, *rpc.RequestProcessorClient) {
	if in.Configuration != nil {
		in.ECfg, cli = encrypt(cli, in.Configuration)
		in.Configuration = nil
	}
	if in.Data != nil {
		in.EData, cli = encrypt(cli, in.Data)
		in.Data = nil
	}
	return in, cli
}

// ReadUserConfig reads the user configuration file and sets UserCfg
func ReadUserConfig(file string) {
	UserCfg = ReadConfig(file)
}

// DecryptUserConfig decrypts the user config if it is
// encrypted. After this call, Data and Configuration members of the
// configuration are set, and EData and ECfg are set to empty
func DecryptUserConfig() {
	if len(UserCfg.AuthKey) > 0 {
		cli, err := ConnectEncServer()
		if err != nil {
			log.Debugf("Cannot connect took agent: %s", err.Error())
			AskPasswordStartDecrypt(DefaultEncTimeout)
			cli, err = ConnectEncServer()
			if err != nil {
				log.Fatal(err)
			}
		}
		m := make(map[string]Remote)
		for k, v := range UserCfg.Remotes {
			m[k] = decryptRemote(cli, v)
		}
		UserCfg.Remotes = m
	}
}

// WriteUserConfig writes the user config file
func WriteUserConfig(cfgFile string) error {
	cfg := UserCfg
	if len(UserCfg.AuthKey) > 0 {
		var cli *rpc.RequestProcessorClient
		m := make(map[string]Remote)
		for k, v := range cfg.Remotes {
			m[k], cli = encryptRemote(cli, v)
		}
		cfg.Remotes = m
	}
	return WriteConfig(cfgFile, cfg)
}

// WriteConfig writes configuration to the file
func WriteConfig(cfgFile string, cfg Configuration) error {
	f, e := os.Create(cfgFile)
	if e != nil {
		return e
	}
	f.Chmod(0600)
	defer f.Close()
	enc := yml.NewEncoder(f)
	return enc.Encode(cfg)
}

// Decode a map[] into a structure
func Decode(in, out interface{}) {
	d, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: out})
	err := d.Decode(in)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error decoding configuration: %s", err))
	}
}

// ConvertMap converts map[interface{}]interface{} into map[string]interface{}
func ConvertMap(in interface{}) interface{} {
	if in == nil {
		return nil
	}
	if mp, ok := in.(map[interface{}]interface{}); ok {
		out := make(map[string]interface{})
		for k, v := range mp {
			out[fmt.Sprint(k)] = ConvertMap(v)
		}
		return out
	}
	if mp, ok := in.(map[string]interface{}); ok {
		out := make(map[string]interface{})
		for k, v := range mp {
			out[k] = ConvertMap(v)
		}
		return out
	}
	if a, ok := in.([]interface{}); ok {
		for i, x := range a {
			a[i] = ConvertMap(x)
		}
		return a
	}
	return in
}

// ConnectEncServer attempts to connect the decryption agent
func ConnectEncServer() (*rpc.RequestProcessorClient, error) {
	socketName, e := homedir.Expand("~/.took.s")
	if e != nil {
		return nil, e
	}
	cli, err := rpc.NewRequestProcessorClient("unix", socketName)
	if err != nil {
		log.Debugf("Cannot connect to server: %s", err.Error())
		return nil, errors.New("You need to use took decrypt to decrypt the tokens")
	}

	return cli, nil
}

// MustConnectEncServer connects to decryption agent and exists if fails
func MustConnectEncServer() *rpc.RequestProcessorClient {
	cli, err := ConnectEncServer()
	if err != nil {
		log.Fatal(err)
	}
	return cli
}

// InsecureAllowed returns true if program path has -insecure in it.
func InsecureAllowed() bool {
	return strings.Contains(os.Args[0], "-insecure")
}

// ValidateInsecureURL validates that a URL is https, or insecure URLs are allowd
func ValidateInsecureURL(url string) {
	if strings.HasPrefix(strings.ToLower(url), "http://") {
		if !InsecureAllowed() {
			fmt.Println("You are using http:// URLs instead of https://. This is not secure. You have to run took as took-insecure to use unencrypted URLs")
			os.Exit(1)
		}
	}
}

// AskPasswordStartDecrypt asks password and starts the decrypt server with the given timout
func AskPasswordStartDecrypt(timeout time.Duration) {
	StartDecrypt(AskPasswordWithPrompt("Configuration/Token encryption password: "), timeout)
}

//StartDecrypt starts another copy of took with decrypt x flag, and passes the password. Panics on fail
func StartDecrypt(password string, timeout time.Duration) {
	_, err := crypta.NewServer(password, UserCfg.AuthKey)
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "decrypt", "x", "-t", timeout.String())
	wr, _ := cmd.StdinPipe()
	cmd.Start()
	wr.Write([]byte(password))
	wr.Write([]byte("\n"))
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		_, err := ConnectEncServer()
		if err == nil {
			return
		}
	}
	log.Fatal("Cannot run decrypt server")
}

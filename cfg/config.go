package cfg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	yml "gopkg.in/yaml.v2"

	"github.com/bserdar/took/crypta/rpc"
)

var UserCfg, CommonCfg Configuration

type Configuration struct {
	AuthKey        string             `yaml:"key,omitempty"`
	Remotes        map[string]Remote  `yaml:"remotes,omitempty"`
	ServerProfiles map[string]Profile `yaml:"serverProfiles,omitempty"`
}

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

type Profile struct {
	Type          string      `yaml:"type"`
	Configuration interface{} `yaml:"cfg,omitempty"`
}

type Remote struct {
	Type          string      `yaml:"type"`
	Configuration interface{} `yaml:"cfg,omitempty"`
	Data          interface{} `yaml:"data,omitempty"`
	ECfg          string      `yaml:"ecfg,omitempty"`
	EData         string      `yaml:"edata,omitempty"`
}

// ReadConfig reads the cfgFile.
func ReadConfig(cfgFile string) Configuration {
	c, err := readConfig(cfgFile)
	if err != nil {
		panic(err)
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
		panic(err)
	}
	var m map[string]interface{}
	json.Unmarshal([]byte(out), &m)
	return m
}

func encrypt(cli *rpc.RequestProcessorClient, in interface{}) string {
	s, _ := json.Marshal(in)
	ret, _ := cli.Encrypt(string(s))
	return ret
}

func decryptRemote(cli *rpc.RequestProcessorClient, in Remote) Remote {
	if len(in.ECfg) > 0 {
		in.Configuration = decrypt(cli, in.ECfg)
	}
	if len(in.EData) > 0 {
		in.Data = decrypt(cli, in.EData)
	}
	return in
}

func encryptRemote(cli *rpc.RequestProcessorClient, in Remote) Remote {
	if in.Configuration != nil {
		in.ECfg = encrypt(cli, in.Configuration)
		in.Configuration = nil
	}
	if in.Data != nil {
		in.EData = encrypt(cli, in.Data)
		in.Data = nil
	}
	return in
}

func ReadUserConfig(file string) {
	UserCfg = ReadConfig(file)
}

func DecryptUserConfig() {
	if len(UserCfg.AuthKey) > 0 {
		cli := ConnectEncServer()
		m := make(map[string]Remote)
		for k, v := range UserCfg.Remotes {
			m[k] = decryptRemote(cli, v)
		}
		UserCfg.Remotes = m
	}
}

func WriteUserConfig(cfgFile string) error {
	cfg := UserCfg
	if len(UserCfg.AuthKey) > 0 {
		cli := ConnectEncServer()
		m := make(map[string]Remote)
		for k, v := range cfg.Remotes {
			m[k] = encryptRemote(cli, v)
		}
		cfg.Remotes = m
	}
	return WriteConfig(cfgFile, cfg)
}

// Write configuration to the file
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

// Decodes a map[] into a structure
func Decode(in, out interface{}) {
	d, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: out})
	err := d.Decode(in)
	if err != nil {
		panic(fmt.Sprintf("Error decoding configuration: %s", err))
	}
}

// Converts map[interface{}]interface{} into map[string]interface{}
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
	} else if sp, ok := in.(map[string]interface{}); ok {
		out := make(map[string]interface{})
		for k, v := range sp {
			out[k] = ConvertMap(v)
		}
		return out
	}
	return in
}

func ConnectEncServer() *rpc.RequestProcessorClient {
	socketName, e := homedir.Expand("~/.took.s")
	if e != nil {
		panic(e)
	}
	cli, err := rpc.NewRequestProcessorClient("unix", socketName)
	if err != nil {
		panic("You need to use took decrypt to decrypt the tokens")
	}

	return cli
}

func InsecureAllowed() bool {
	return strings.Contains(os.Args[0], "-insecure")
}

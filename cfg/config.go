package cfg

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mitchellh/mapstructure"
	yml "gopkg.in/yaml.v2"
)

var UserCfg, CommonCfg Configuration

type Configuration struct {
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

func ReadUserConfig(file string) {
	UserCfg = ReadConfig(file)
}

func WriteUserConfig(cfgFile string) error {
	return WriteConfig(cfgFile, UserCfg)
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

package cfg

import (
	"io/ioutil"
	"os"

	yml "gopkg.in/yaml.v2"
)

type Configuration struct {
	Remotes map[string]Remote `yaml:"remotes"`
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

// ReadCommonConfig reads the common configuration file under /etc/took.yaml
func ReadCommonConfig() Configuration {
	cfg, _ := readConfig("/etc/took.yaml")
	return cfg
}

// Write configuration to the file
func WriteConfig(cfgFile string, cfg Configuration) error {
	f, e := os.Create(cfgFile)
	if e != nil {
		return e
	}
	defer f.Close()
	enc := yml.NewEncoder(f)
	return enc.Encode(cfg)
}

package magickzip

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config is magickzip yaml format conf
type Config struct {
	EnableWebp bool
	Root       string
	Structure  map[interface{}]interface{}
	Resize     map[interface{}]interface{}
	Modulate   map[interface{}]interface{}
}

// LoadConfig load yml file and return Config
func LoadConfig(file string) (*Config, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	err = yaml.Unmarshal(buf, conf)
	if err != nil {
		return nil, err
	}

	err = conf.validate()
	if err != nil {
		return nil, err
	}

	// conf.Structure's key is only one
	for k := range conf.Structure {
		conf.Root = k.(string)
	}

	return conf, nil
}

func (conf *Config) validate() error {
	if len(conf.Structure) > 1 || len(conf.Structure) == 0 {
		return fmt.Errorf("'structure' has only one tag")
	}
	return nil
}

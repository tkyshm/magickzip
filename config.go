package magickzip

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Structure Structure
	Resize    map[string]interface{}
	Modulate  map[string]interface{}
}

type Structure struct {
	Tree map[string]interface{}
}

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

	return conf, nil
}

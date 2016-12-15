package agent

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Server 	string `yaml:"server"`
	Checks 	[]string `yaml:"checks"`
}

var config Config

func init() {
	config = Config{
		Server: "http://localhost:3000",
		Checks: []string{},
	}
}

func GetConfig() *Config {
	return &config
}

func LoadConfig(file string) error {
	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(fileData, &config)
	if err != nil {
		return err
	}

	return nil
}

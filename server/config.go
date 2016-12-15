package server

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Port  int          `yaml:"port"`
	Users []ConfigUser `yaml:"users"`
}

type ConfigUser struct {
	Name        string   `yaml:"name"`
	SigningKeys []string `yaml:"keys"`
}

var config Config

func init() {
	config = Config{
		Port:  3000,
		Users: []ConfigUser{},
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

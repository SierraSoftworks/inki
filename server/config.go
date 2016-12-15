package server

import (
	"io/ioutil"

	"strings"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/openpgp"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Port  int          `yaml:"port"`
	Users []ConfigUser `yaml:"users"`
}

func (c *Config) GetUser(name string) *ConfigUser {
	for _, u := range c.Users {
		if u.Name == name {
			return &u
		}
	}

	return nil
}

type ConfigUser struct {
	Name    string `yaml:"name"`
	KeyRing string `yaml:"keyring"`
}

func (u *ConfigUser) GetKeyRing() (openpgp.KeyRing, error) {
	r := strings.NewReader(u.KeyRing)
	el, err := openpgp.ReadArmoredKeyRing(r)
	if err != nil {
		log.WithField("keyring", u.KeyRing).WithError(err).Error("Failed to parse user's keyring")
		return nil, err
	}

	return el, nil
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

package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/user"

	"os"

	"path/filepath"

	"github.com/SierraSoftworks/inki/crypto"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var AuthorizedKeysCommand = cli.Command{
	Category: "Agent",
	Name:     "authorized-keys",
	Usage:    "Get the correctly formatted authorized keys file for the current user",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Usage:  "The configuration file that the Inki client will use",
			Value:  "~/.inki.yml",
			EnvVar: "INKI_CLIENT_CONFIG",
		},
	},
	Before: func(c *cli.Context) error {
		log.SetOutput(os.Stderr)

		cPath, err := filepath.Abs(c.String("config"))
		if err != nil {
			return err
		}

		err = LoadConfig(cPath)
		if c.IsSet("config") {
			if err != nil {
				log.WithError(err).Errorf("Failed to read configuration file '%s'", c.String("config, c"))
				return err
			}
		} else if err != nil {
			log.Warn("No configuration file provided, using empty defaults")
		}

		err = RunChecks()
		if err != nil {
			return err
		}

		return nil
	},
	Action: func(c *cli.Context) error {
		username := ""

		if c.NArg() > 0 {
			username = c.Args().First()
		}

		if username == "" {
			u, err := user.Current()
			if err != nil {
				log.WithError(err).Error("Failed to get current user's details")
				return err
			}

			username = u.Username
		}

		url := fmt.Sprintf("%s/api/v1/user/%s/keys", GetConfig().Server, username)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.WithError(err).Error("Failed to prepare request for user keys")
			return fmt.Errorf("Failed to prepare request for user keys")
		}

		log.WithFields(log.Fields{
			"server": GetConfig().Server,
			"user":   username,
		}).Info("Fetching authorized keys for the current user")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.WithError(err).Error("Failed to make request for user keys")
			return fmt.Errorf("Request for user keys failed to server '%s'", url)
		}

		if res.StatusCode != 200 {
			log.WithFields(log.Fields{
				"user":   username,
				"server": GetConfig().Server,
				"status": res.StatusCode,
			}).Error("Failed to get list of keys for user")
			return fmt.Errorf("Failed to get list of keys for '%s'", username)
		}

		keys := []crypto.Key{}
		if err := json.NewDecoder(res.Body).Decode(&keys); err != nil {
			log.WithError(err).Error("Failed to parse response from server")
			return fmt.Errorf("Failed to parse response from server")
		}

		for _, k := range keys {
			if err := k.Validate(); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"user":   k.User,
					"expire": k.Expires,
					"key":    k.PublicKey,
				}).Warn("Key did not pass validation")
			} else {
				fmt.Printf("%s\n", k.PublicKey)
			}
		}

		return nil
	},
}

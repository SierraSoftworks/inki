package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"os"

	"github.com/SierraSoftworks/inki/crypto"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var listKeysCommand = cli.Command{
	Name:  "list",
	Usage: "Gets the list of keys currently registered on the server",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "authorized-keys, a",
			Usage: "Format the resulting data in a way that is compatible with authorized_keys",
		},
		cli.BoolFlag{
			Name:  "expired, x",
			Usage: "Includes keys which have expired in the output",
		},
	},
	Before: func(c *cli.Context) error {
		log.SetOutput(os.Stderr)
		return nil
	},
	Action: func(c *cli.Context) error {
		if c.NArg() < 1 {
			return fmt.Errorf("Missing user and host argument")
		}

		u, err := url.Parse(c.Args().First())
		if err != nil {
			log.WithError(err).Error("Failed to parse host URL")
			return fmt.Errorf("Failed to parse user and host argument")
		}

		if u.User == nil || u.User.String() == "" {
			log.Error("Host URL did not contain a username")
			return fmt.Errorf("Host address did not contain a username")
		}

		if u.Scheme == "" {
			u.Scheme = "http"
		}

		server := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		url := server
		if u.User.Username() != "" {
			url = fmt.Sprintf("%s/api/v1/user/%s/keys", url, u.User.Username())
		} else {
			url = fmt.Sprintf("%s/api/v1/keys", url)
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.WithError(err).Error("Failed to prepare request for keys")
			return fmt.Errorf("Failed to prepare request for keys")
		}

		log.WithFields(log.Fields{
			"server": server,
			"user":   u.User.Username(),
		}).Info("Fetching authorized keys")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.WithError(err).Error("Failed to make request for user keys")
			return fmt.Errorf("Request for user keys failed to server '%s'", url)
		}

		if res.StatusCode != 200 {
			log.WithFields(log.Fields{
				"user":   u.User.Username(),
				"server": server,
				"status": res.StatusCode,
			}).Error("Failed to get list of keys")
			return fmt.Errorf("Failed to get list of keys")
		}

		keys := []crypto.Key{}
		if err := json.NewDecoder(res.Body).Decode(&keys); err != nil {
			log.WithError(err).Error("Failed to parse response from server")
			return fmt.Errorf("Failed to parse response from server")
		}

		allowExpired := c.IsSet("expired")
		if c.IsSet("authorized-keys") {
			for _, k := range keys {
				err := k.Validate()
				if allowExpired || err == nil {
					fmt.Printf("%s\n", k.PublicKey)
				}
			}
		} else {
			fmt.Println("Authorized keys:")
			for _, k := range keys {
				err := k.Validate()
				if allowExpired || err == nil {
					fmt.Printf(" - Username:     %s\n", k.User)
					fmt.Printf("   Fingerprint:  %s\n", k.Fingerprint())
					fmt.Printf("   Expires:      %s\n", k.Expires)
					fmt.Println()
				}
			}
		}

		return nil
	},
}

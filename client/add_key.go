package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"os"

	"bytes"
	"net/url"

	"io/ioutil"

	"github.com/SierraSoftworks/inki/crypto"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"
	"golang.org/x/crypto/ssh/terminal"
)

var addKeyCommand = cli.Command{
	Name:      "add",
	Usage:     "Adds an SSH key to the Inki key server",
	UsageText: "user@inki-server",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "pgp-key, p",
			Usage: "The PGP private key you wish to use to sign this request",
		},
		cli.StringFlag{
			Name:  "file, f",
			Usage: "The SSH public key file which you would like to submit",
		},
		cli.DurationFlag{
			Name:  "expire, x",
			Usage: "The amount of time that the key should be valid for",
			Value: time.Hour,
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

		keyData := bytes.NewBuffer([]byte{})
		if c.IsSet("file") {
			kd, err := ioutil.ReadFile(c.String("file"))
			if err != nil {
				log.
					WithError(err).
					WithField("file", c.String("file")).
					Debug("Failed to open key file for reading")
				return fmt.Errorf("Failed to read key file")
			}

			keyData = bytes.NewBuffer(kd)
		} else {
			_, err := keyData.ReadFrom(os.Stdin)
			if err != nil {
				log.
					WithError(err).
					Debug("Failed to read key data from stdin")
				return fmt.Errorf("Failed to read key from stdin")
			}
		}

		key := &crypto.Key{
			User:      u.User.Username(),
			PublicKey: keyData.String(),
			Expires:   time.Now().Add(c.Duration("expire")),
		}

		p, err := ioutil.ReadFile(c.String("pgp-key"))
		if err != nil {
			log.
				WithError(err).
				WithField("file", c.String("pgp-key")).
				Debug("Failed to read the pgp-key file")
			return fmt.Errorf("Failed to read the pgp-key you provided")
		}

		kr, err := openpgp.ReadArmoredKeyRing(bytes.NewBuffer(p))
		if err != nil {
			log.WithError(err).
				WithField("file", c.String("pgp-key")).
				Debug("Failed to decode the pgp-key file")
			return fmt.Errorf("Failed to decode the pgp-key you provided")
		}

		pk := kr[0].PrivateKey
		if pk.Encrypted {
			if !c.IsSet("file") {
				log.
					Debug("Private key is encrypted and stdin has been used to read the SSH key")
				return fmt.Errorf("Private key is encrypted and stdin was used to read the SSH key")
			}

			fmt.Print("Enter PGP key password: ")
			pw, err := terminal.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				log.
					WithError(err).
					Debug("Failed to request password from user")
				return fmt.Errorf("Failed to request password input")
			}

			err = pk.Decrypt(pw)
			if err != nil {
				log.
					WithError(err).
					Debug("Failed to decrypt the PGP private key")
				return fmt.Errorf("Failed to decrypt the PGP private key, please check that your password is correct")
			}
		}

		reqData := bytes.NewBuffer([]byte{})
		reqStream, err := clearsign.Encode(reqData, pk, nil)
		if err != nil {
			log.
				WithError(err).
				Debug("Failed to prepare signing packet")
			return fmt.Errorf("Failed to prepare signing packet")
		}

		err = json.NewEncoder(reqStream).Encode(key)
		if err != nil {
			log.
				WithError(err).
				Debug("Failed to encode key request")
			return fmt.Errorf("Failed to encode key request")
		}

		reqStream.Close()

		url := fmt.Sprintf("%s://%s/api/v1/keys", u.Scheme, u.Host)
		req, err := http.NewRequest("POST", url, reqData)
		if err != nil {
			log.
				WithError(err).
				Debug("Failed to prepare request")
			return fmt.Errorf("Failed to prepare request to server")
		}

		log.WithFields(log.Fields{
			"user":   u.User.Username(),
			"key":    key.PublicKey,
			"expire": key.Expires,
		}).Info("Submitting new key for user")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.
				WithError(err).
				WithFields(log.Fields{
					"server": GetConfig().Server,
				}).
				Debug("Failed to send key request to server")
			return fmt.Errorf("Request to send key request to server '%s'", u.Host)
		}

		if res.StatusCode != 200 {
			log.
				WithFields(log.Fields{
					"server": u.Host,
					"status": res.StatusCode,
				}).
				Debug("Failed to send key request to server")
			return fmt.Errorf("Failed to send key request to server: %s", res.Status)
		}

		keys := []crypto.Key{}
		if err := json.NewDecoder(res.Body).Decode(&keys); err != nil {
			log.
				WithError(err).
				Debug("Failed to parse response from server")
			return fmt.Errorf("Failed to parse response from server")
		}

		fmt.Println("Added keys:")
		for _, k := range keys {
			fmt.Printf(" - Username:     %s\n", k.User)
			fmt.Printf("   Fingerprint:  %s\n", k.Fingerprint())
			fmt.Printf("   Expires:      %s\n", k.Expires)
			fmt.Println()
		}

		return nil
	},
}

package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:  "server",
	Usage: "Run the Inki server daemon",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Usage:  "The configuration file to be used by this daemon",
			EnvVar: "INKI_CONFIG",
		},
	},
	Before: func(c *cli.Context) error {
		if c.IsSet("config, c") {
			err := LoadConfig(c.String("config, c"))
			if err != nil {
				log.WithError(err).Errorf("Failed to read configuration file '%s'", c.String("config, c"))
				return err
			}
		} else {
			log.Warn("No configuration file provided, using empty defaults")
		}

		return nil
	},
	Action: func(c *cli.Context) error {
		return nil
	},
}

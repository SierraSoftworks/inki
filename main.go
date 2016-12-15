package main

import (
	"os"
	"strings"

	"github.com/SierraSoftworks/inki/client"
	"github.com/SierraSoftworks/inki/server"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var version = "v1.0.0"

func main() {
	app := cli.NewApp()

	app.Name = "Inki"
	app.Author = "Benjamin Pannell"
	app.Email = "admin@sierrasoftworks.com"
	app.Version = version
	app.UsageText = "An SSH key distribution tool"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level, L",
			Value: "WARN",
			Usage: "Log level to use (ERROR, WARN, INFO, DEBUG)",
		},
	}

	app.Before = func(c *cli.Context) error {
		logLevel := strings.ToUpper(c.String("log-level"))
		switch logLevel {
		case "ERROR":
			log.SetLevel(log.ErrorLevel)
		case "WARN":
			log.SetLevel(log.WarnLevel)
		case "INFO":
			log.SetLevel(log.InfoLevel)
		case "DEBUG":
			log.SetLevel(log.DebugLevel)
		default:
			log.SetLevel(log.WarnLevel)
		}

		return nil
	}

	app.Commands = []cli.Command{
		server.Command,
		client.KeysCommands,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.WithError(err).Fatal("Failed to run application")
	}
}

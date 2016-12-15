package main

import (
	"github.com/SierraSoftworks/inki/server"
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

	app.Commands = []cli.Command{
		server.Command,
	}

	app.RunAndExitOnError()
}

package client

import "github.com/urfave/cli"

var KeysCommands = cli.Command{
	Category: "Client",
	Name:     "key",
	Usage:    "Tools to manage keys on your Inki server",
	Subcommands: []cli.Command{
		addKeyCommand,
		listKeysCommand,
	},
}

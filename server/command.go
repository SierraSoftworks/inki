package server

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/rs/cors"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Category: "Server",
	Name:     "server",
	Usage:    "Run the Inki server daemon",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Usage:  "The configuration file to be used by this daemon",
			EnvVar: "INKI_CONFIG",
			Value:  "/etc/inki/server.yml",
		},
		cli.IntFlag{
			Name:   "port, P",
			Usage:  "The port on which the Inki server should listen",
			EnvVar: "PORT",
			Value:  3000,
		},
	},
	Before: func(c *cli.Context) error {
		if c.IsSet("config") {
			err := LoadConfig(c.String("config"))
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
		port := c.Int("port")
		log.WithField("port", port).Info("Starting server")

		mux := http.NewServeMux()
		mux.Handle("/api/", http.StripPrefix("/api", Router()))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(404)
			w.Write([]byte(`{"code": 404, "error": "Not Found", "message": "The method you attempted to make use of could not be found on our system."}`))
		})

		return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), cors.New(cors.Options{
			AllowCredentials: true,
			AllowedOrigins:   []string{"*"},
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
			Debug:            false,
		}).Handler(mux))
	},
}

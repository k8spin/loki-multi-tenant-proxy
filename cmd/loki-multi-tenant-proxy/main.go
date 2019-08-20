package main

import (
	"os"

	"github.com/urfave/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := cli.NewApp()
	app.Name = "Loki Multitenant Proxy"
	app.Usage = "Makes your Loki server multi tenant"
	app.Version = version
	app.Author = "Ángel Barrera - @angelbarrera92"
	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "Runs the Loki multi tenant proxy",
			Action: nil,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port",
					Usage: "Port to expose this loki proxy",
					Value: 3501,
				}, cli.StringFlag{
					Name:  "loki-server",
					Usage: "Loki server endpoint",
					Value: "http://localhost:3500",
				}, cli.StringFlag{
					Name:  "auth-config",
					Usage: "AuthN yaml configuration file path",
					Value: "authn.yaml",
				},
			},
		},
	}
	app.Run(os.Args)
}

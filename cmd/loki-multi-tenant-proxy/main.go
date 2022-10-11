package main

import (
	"os"

	proxy "github.com/giantswarm/loki-multi-tenant-proxy/internal/app/loki-multi-tenant-proxy"
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
	app.Authors = []cli.Author{
		{Name: "Angel Barrera", Email: "angel@k8spin.cloud"},
		{Name: "Pau Rosello", Email: "pau@k8spin.cloud"},
	}
	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "Runs the Loki multi tenant proxy",
			Action: proxy.Serve,
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
				}, cli.BoolFlag{
					Name:  "keep-orgid",
					Usage: "Don't change OrgID header (proxy is only used for authent)",
				},
			},
		},
	}
	app.Run(os.Args)
}

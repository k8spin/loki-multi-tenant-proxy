package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/giantswarm/loki-multi-tenant-proxy/internal/pkg"
	"github.com/urfave/cli/v2"
)

// Serve serves
func Serve(c *cli.Context) error {
	lokiServerURL, _ := url.Parse(c.String("loki-server"))
	serveAt := fmt.Sprintf(":%d", c.Int("port"))
	authConfigLocation := c.String("auth-config")
	authConfig, _ := pkg.ParseConfig(&authConfigLocation)
	authConfig.KeepOrgID = c.Bool("keep-orgid")

	http.HandleFunc("/", createHandler(lokiServerURL, authConfig))
	if err := http.ListenAndServe(serveAt, nil); err != nil {
		log.Fatalf("Loki multi tenant proxy can not start %v", err)
		return err
	}
	return nil
}

func createHandler(lokiServerURL *url.URL, authConfig *pkg.Authn) http.HandlerFunc {
	reverseProxy := httputil.NewSingleHostReverseProxy(lokiServerURL)
	return LogRequest(BasicAuth(ReverseLoki(reverseProxy, lokiServerURL), authConfig))
}

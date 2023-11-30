package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/giantswarm/loki-multi-tenant-proxy/internal/pkg"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// Serve serves
func Serve(c *cli.Context) error {
	lokiServerURL, _ := url.Parse(c.String("loki-server"))
	addr := fmt.Sprintf(":%d", c.Int("port"))
	authConfigLocation := c.String("auth-config")
	authConfig, _ := pkg.ParseConfig(&authConfigLocation)
	authConfig.KeepOrgID = c.Bool("keep-orgid")

	logger := zap.Must(zap.NewProduction())
	defer logger.Sync()

	errorLogger, err := zap.NewStdLogAt(logger, zap.ErrorLevel)
	if err != nil {
		logger.Fatal("Can not create error logger", zap.Error(err))
	}

	http.HandleFunc("/", createHandler(lokiServerURL, authConfig, logger, errorLogger))
	server := &http.Server{Addr: addr, ErrorLog: errorLogger}
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Loki multi tenant proxy can not start", zap.Error(err))
		return err
	}
	logger.Info("Starting HTTP server", zap.String("addr", addr))
	return nil
}

func createHandler(lokiServerURL *url.URL, authConfig *pkg.Authn, logger *zap.Logger, errorLogger *log.Logger) http.HandlerFunc {
	reverseProxy := httputil.NewSingleHostReverseProxy(lokiServerURL)
	reverseProxy.ErrorLog = errorLogger
	return Logger(BasicAuth(ReverseLoki(reverseProxy, lokiServerURL), authConfig), logger)
}

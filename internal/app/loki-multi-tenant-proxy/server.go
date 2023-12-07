package proxy

import (
	"fmt"
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

	logLevel := c.String("log-level")
	if logLevel == "" {
		logLevel = "INFO"
	}

	var logger *zap.Logger
	{
		zapConfig := zap.NewProductionConfig()
		level, err := zap.ParseAtomicLevel(logLevel)
		if err != nil {
			panic(fmt.Sprintf("Could not parse log level %v", zap.Error(err)))
		}
		zapConfig.Level = level

		logger = zap.Must(zapConfig.Build())
		defer logger.Sync()
	}

	errorLogger, err := zap.NewStdLogAt(logger, zap.ErrorLevel)
	if err != nil {
		logger.Fatal("Can not create error logger", zap.Error(err))
	}

	var reverseProxy *httputil.ReverseProxy
	{
		reverseProxy = &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(lokiServerURL)
				r.Out.Host = lokiServerURL.Host
				r.Out.Header.Set("X-Forwarded-Host", lokiServerURL.Host)
				orgID := r.In.Context().Value(OrgIDKey)

				if orgID != "" {
					r.Out.Header.Set("X-Scope-OrgID", orgID.(string))
				}
			},
			ErrorLog: errorLogger,
		}
	}

	handlers := Logger(
		BasicAuth(
			ReverseLoki(reverseProxy),
			authConfig,
		),
		logger,
	)

	http.HandleFunc("/", handlers)
	server := &http.Server{Addr: addr, ErrorLog: errorLogger}
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Loki multi tenant proxy can not start", zap.Error(err))
		return err
	}
	logger.Info("Starting HTTP server", zap.String("addr", addr))
	return nil
}

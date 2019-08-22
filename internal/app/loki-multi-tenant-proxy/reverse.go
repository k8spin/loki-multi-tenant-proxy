package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/angelbarrera92/loki-multi-tenant-proxy/internal/pkg"
)

// ReverseLoki a
func ReverseLoki(reverseProxy *httputil.ReverseProxy, lokiServerURL *url.URL, users *pkg.Authn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		modifyRequest(r, lokiServerURL, users)
		reverseProxy.ServeHTTP(w, r)
	}
}

func modifyRequest(r *http.Request, lokiServerURL *url.URL, users *pkg.Authn) {
	r.URL.Scheme = lokiServerURL.Scheme
	r.URL.Host = lokiServerURL.Host
	r.Host = lokiServerURL.Host
	userName, _, _ := r.BasicAuth()
	orgID, _ := pkg.GetOrgID(userName, users)
	r.Header.Set("X-Forwarded-Host", r.Host)
	r.Header.Set("X-Scope-OrgID", orgID)
}

package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ReverseLoki a
func ReverseLoki(reverseProxy *httputil.ReverseProxy, lokiServerURL *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		modifyRequest(r, lokiServerURL)
		reverseProxy.ServeHTTP(w, r)
	}
}

func modifyRequest(r *http.Request, lokiServerURL *url.URL) {
	r.URL.Scheme = lokiServerURL.Scheme
	r.URL.Host = lokiServerURL.Host
	r.Host = lokiServerURL.Host
	orgID := r.Context().Value(OrgIDKey)
	r.Header.Set("X-Forwarded-Host", r.Host)
	if orgID != "" {
		r.Header.Set("X-Scope-OrgID", orgID.(string))
	}
}

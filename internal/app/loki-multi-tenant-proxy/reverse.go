package proxy

import (
	"net/http"
	"net/http/httputil"
)

// ReverseLoki reverse proxies to Loki
func ReverseLoki(reverseProxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reverseProxy.ServeHTTP(w, r)
	}
}

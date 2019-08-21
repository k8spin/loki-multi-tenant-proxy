package proxy

import (
	"crypto/subtle"
	"net/http"

	"github.com/angelbarrera92/loki-multi-tenant-proxy/internal/pkg"
)

const realm = "Loki multi-tenant proxy"

// BasicAuth can be used as a middleware chain to authenticate users before proxying a request
func BasicAuth(handler http.HandlerFunc, users *pkg.Authn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !isAuthorized(user, pass, users) {
			writeUnauthorisedResponse(w)
			return
		}
		handler(w, r)
	}
}

func isAuthorized(user string, pass string, users *pkg.Authn) bool {
	for _, v := range users.Users {
		if subtle.ConstantTimeCompare([]byte(user), []byte(v.Username)) == 1 && subtle.ConstantTimeCompare([]byte(pass), []byte(v.Password)) == 1 {
			return true
		}
	}
	return false
}

func writeUnauthorisedResponse(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.WriteHeader(401)
	w.Write([]byte("Unauthorised\n"))
}

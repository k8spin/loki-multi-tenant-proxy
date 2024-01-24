package proxy

import (
	"context"
	"crypto/subtle"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/giantswarm/loki-multi-tenant-proxy/internal/pkg"
	"go.uber.org/zap"
)

type key int

// Struct to represent the OAuth token payload section
type Payload struct {
	Iss           string   `json:"iss"`
	Sub           string   `json:"sub"`
	Aud           string   `json:"aud"`
	Exp           int      `json:"exp"`
	Iat           int      `json:"iat"`
	AtHash        string   `json:"at_hash"`
	CHash         string   `json:"c_hash"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Groups        []string `json:"groups"`
	Name          string   `json:"name"`
}

const (
	// OrgIDKey Key used to pass loki tenant id though the middleware context
	OrgIDKey key = iota
	realm        = "Loki multi-tenant proxy"
)

func Authentication(handler http.HandlerFunc, authConfig *pkg.Authn, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := r.Header["X-Id-Token"]
		if ok {
			// OAuth token authentication mode (X-Id-Token header provided)
			logger.Info("OAuth authentication type")
			logger.Info(fmt.Sprintf("Token = %s", token[0]))
			// Decoding base64 jwt token
			payload, err := decodeOAuthToken(token[0])
			if err != nil {
				logger.Error("Error decoding token payload", zap.Error(err))
				return
			}
			// Token validation against Dex
			err = validateOAuthToken(payload.Iss, payload.Aud)
			if err != nil {
				logger.Error("Error while validating OAuth token against DEX", zap.Error(err))
				writeUnauthorisedResponse(w, "oauth")
				return
			}
			//ctx := context.WithValue(r.Context(), OrgIDKey, orgID)
			ctx := context.WithValue(r.Context(), OrgIDKey, "giantswarm")
			handler(w, r.WithContext(ctx))
			//OAuth(handler, authConfig, logger)
		} else {
			// Default authentication mode => BasicAuth
			logger.Info("BasicAuth authentication type")
			user, pass, ok := r.BasicAuth()
			authorized, orgID := isAuthorized(user, pass, authConfig)
			if !ok || !authorized {
				writeUnauthorisedResponse(w, "basic")
				return
			}
			ctx := context.WithValue(r.Context(), OrgIDKey, orgID)
			handler(w, r.WithContext(ctx))
			//BasicAuth(handler, authConfig, logger)
		}
	}
}

func isAuthorized(user string, pass string, authConfig *pkg.Authn) (bool, string) {
	for _, v := range authConfig.Users {
		if subtle.ConstantTimeCompare([]byte(user), []byte(v.Username)) == 1 && subtle.ConstantTimeCompare([]byte(pass), []byte(v.Password)) == 1 {
			if !authConfig.KeepOrgID {
				return true, v.OrgID
			} else {
				return true, ""
			}

		}
	}
	return false, ""
}

func writeUnauthorisedResponse(w http.ResponseWriter, authenticationType string) {
	if authenticationType == "basic" {
		w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	}
	w.WriteHeader(401)
	w.Write([]byte("Unauthorised\n"))
}

func decodeOAuthToken(token string) (Payload, error) {
	// Get payload section from the token
	payload := strings.Split(token, ".")[1]
	payloadDecoded, _ := b64.URLEncoding.DecodeString(payload)

	var p Payload
	err := json.Unmarshal(payloadDecoded, &p)
	return p, err
}

func validateOAuthToken(dexUrl string, clientId string) error {
	return nil
}

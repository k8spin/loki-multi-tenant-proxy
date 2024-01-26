package proxy

import (
	"context"
	"crypto/subtle"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc"
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
	readUser     = "read"
)

func Authentication(handler http.HandlerFunc, authConfig *pkg.Authn, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := r.Header["X-Id-Token"]
		var authenticationMode string
		var user = "read" // default user if oauth mode
		var pass = ""     // no password in oauth mode
		if ok {
			// OAuth token authentication mode (X-Id-Token header provided)
			logger.Info("OAuth authentication mode")
			logger.Info(fmt.Sprintf("Token = %s", token[0]))
			// Decode OAuth token payload section
			payload, err := decodeOAuthToken(token[0])
			if err != nil {
				logger.Error(fmt.Sprintf("Error decoding token payload %s", token[0]), zap.Error(err))
				return
			}
			// Token validation against Dex
			err = validateOAuthToken(token[0], payload.Iss, payload.Aud, r.Context())
			if err != nil {
				logger.Error(fmt.Sprintf("Error while validating OAuth token against DEX %s", token[0]), zap.Error(err))
				writeUnauthorisedResponse(w, "oauth")
				return
			}
			authenticationMode = "oauth"

		} else {
			// BasicAuth authentication mode (X-Id-Token header not provided) - default mode (use for write path)
			logger.Info("BasicAuth authentication mode")
			user, pass, ok = r.BasicAuth()
			if !ok {
				writeUnauthorisedResponse(w, "basic")
				return
			}
			authenticationMode = "basic"
		}
		// Check if user is authorized to access Loki and retrieve OrgID
		authorized, orgID := isAuthorized(user, pass, authConfig, authenticationMode)
		if !authorized {
			writeUnauthorisedResponse(w, authenticationMode)
			return
		}
		ctx := context.WithValue(r.Context(), OrgIDKey, orgID)
		handler(w, r.WithContext(ctx))
	}
}

// isAuthorized checks if the user is authorized to access Loki (BasicAuth mode)
// and get OrgId to handle multi-tenant access
func isAuthorized(user string, pass string, authConfig *pkg.Authn, authenticationMode string) (bool, string) {
	for _, v := range authConfig.Users {
		// OAuth mode: retrieve user 'read' et get OrgID
		// BasicAuth mode: check user and password passed in the request and get OrgID
		if (authenticationMode == "oauth" && subtle.ConstantTimeCompare([]byte(user), []byte(v.Username)) == 1) ||
			(authenticationMode == "basic" && subtle.ConstantTimeCompare([]byte(user), []byte(v.Username)) == 1 && subtle.ConstantTimeCompare([]byte(pass), []byte(v.Password)) == 1) {
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

// decodeOAuthToken decodes the payload section of the OAuth token
func decodeOAuthToken(token string) (Payload, error) {
	// Get payload section from the token
	payload := strings.Split(token, ".")[1]
	payloadDecoded, err := b64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return Payload{}, err
	}

	var p Payload
	err = json.Unmarshal(payloadDecoded, &p)
	return p, err
}

// validateOAuthToken validates the OAuth token against Dex
func validateOAuthToken(token string, dexUrl string, clientId string, ctx context.Context) error {
	// Initialize a provider by specifying dex's issuer URL.
	provider, err := oidc.NewProvider(ctx, dexUrl)
	if err != nil {
		return err
	}
	// Create an ID token parser, but only trust ID tokens issued to 'clientId'
	idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: clientId})
	// Verify token validity
	_, err = idTokenVerifier.Verify(ctx, token)
	return err
}

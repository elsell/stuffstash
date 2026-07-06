package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
)

type MobileAuthOptions struct {
	Issuer      string
	ClientID    string
	RedirectURI string
	Scopes      []string
}

type mobileAuthMetadataResponse struct {
	Issuer      string   `json:"issuer"`
	ClientID    string   `json:"clientId"`
	RedirectURI string   `json:"redirectUri"`
	Scopes      []string `json:"scopes"`
}

const supportedMobileRedirectURI = "stuffstash://auth/callback"

func handleMobileAuthMetadata(options MobileAuthOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metadata, ok := mobileAuthMetadata(options)
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(&shared.ErrorEnvelope{
				BodyError: shared.ErrorBody{
					Code:    "mobile_auth_unavailable",
					Message: "Mobile sign-in is not configured.",
					Details: []shared.ErrorDetail{},
				},
				Meta: shared.Meta{},
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(shared.SuccessEnvelope[mobileAuthMetadataResponse]{
			Data: metadata,
			Meta: shared.Meta{},
		})
	}
}

func mobileAuthMetadata(options MobileAuthOptions) (mobileAuthMetadataResponse, bool) {
	issuer := strings.TrimRight(strings.TrimSpace(options.Issuer), "/")
	clientID := strings.TrimSpace(options.ClientID)
	redirectURI := strings.TrimSpace(options.RedirectURI)
	scopes := cleanScopes(options.Scopes)
	if issuer == "" || clientID == "" || redirectURI != supportedMobileRedirectURI || len(scopes) == 0 {
		return mobileAuthMetadataResponse{}, false
	}

	return mobileAuthMetadataResponse{
		Issuer:      issuer,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scopes:      scopes,
	}, true
}

func cleanScopes(values []string) []string {
	scopes := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		scope := strings.TrimSpace(value)
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}
	return scopes
}

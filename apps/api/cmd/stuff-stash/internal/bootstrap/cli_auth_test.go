package bootstrap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/config"
)

func TestCLIAuthDeviceMethodRequiresPolicyAndProviderSupport(t *testing.T) {
	for _, advertised := range []bool{false, true} {
		for _, enabled := range []bool{false, true} {
			var issuer string
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/.well-known/openid-configuration" {
					http.NotFound(w, r)
					return
				}
				metadata := map[string]any{"issuer": issuer, "authorization_endpoint": issuer + "/auth", "token_endpoint": issuer + "/token", "jwks_uri": issuer + "/keys", "response_types_supported": []string{"code"}, "subject_types_supported": []string{"public"}, "id_token_signing_alg_values_supported": []string{"RS256"}}
				if advertised {
					metadata["device_authorization_endpoint"] = issuer + "/device"
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(metadata)
			}))
			issuer = provider.URL
			metadata, err := buildCLIAuthMetadata(context.Background(), config.Config{AuthMode: "oidc", OIDCIssuer: issuer, OIDCCLIClientID: "cli", OIDCCLIScopes: []string{"openid"}, OIDCCLIDeviceAuthEnabled: enabled})
			provider.Close()
			if err != nil {
				t.Fatal(err)
			}
			expected := 1
			if advertised && enabled {
				expected = 2
			}
			if metadata == nil || len(metadata.LoginMethods) != expected {
				t.Fatalf("advertised=%v enabled=%v metadata=%+v", advertised, enabled, metadata)
			}
		}
	}
	metadata, err := buildCLIAuthMetadata(context.Background(), config.Config{AuthMode: "local-dev", OIDCCLIClientID: "cli"})
	if err != nil || metadata != nil {
		t.Fatal("local development must not advertise production CLI auth")
	}
}

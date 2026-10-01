package httpserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/mcpserver"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
)

func TestMCPSignedOIDCBoundaryRejectsInvalidTokens(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "mcp-test"))
	if err != nil {
		t.Fatal(err)
	}
	var issuerURL string
	issuer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]any{"issuer": issuerURL, "authorization_endpoint": issuerURL + "/authorize", "token_endpoint": issuerURL + "/token", "jwks_uri": issuerURL + "/keys", "id_token_signing_alg_values_supported": []string{"RS256"}})
		case "/keys":
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "mcp-test", Algorithm: "RS256", Use: "sig"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer issuer.Close()
	issuerURL = issuer.URL
	authenticator, err := auth.NewOIDCAuthenticatorFromIssuer(context.Background(), issuerURL, "stuffstash-client")
	if err != nil {
		t.Fatal(err)
	}
	token := func(iss, aud string, expiry time.Time) string {
		t.Helper()
		value, err := jwt.Signed(signer).Claims(jwt.Claims{Issuer: iss, Subject: "owner", Audience: jwt.Audience{aud}, IssuedAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), Expiry: jwt.NewNumericDate(expiry)}).Serialize()
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	valid := token(issuerURL, "stuffstash-client", time.Now().Add(time.Hour))
	identity, err := authenticator.Authenticate(context.Background(), "Bearer "+valid)
	if err != nil {
		t.Fatal(err)
	}
	application := newSeededTestAppWithBlobAuthorizerAndImportSource(t, seededState{tenants: []seedTenant{{id: "home", name: "Home", owner: identity.ID.String()}}, inventories: []seedInventory{{id: "main", tenantID: "home", name: "Main", owner: identity.ID.String()}}}, nil, memory.NewAuthorizer(), nil, authenticator)
	handler, err := mcpserver.New(application, mcpserver.Options{PublicURL: "https://stash.example.test/mcp", AuthMode: "oidc", Issuer: issuerURL})
	if err != nil {
		t.Fatal(err)
	}
	server := NewServerWithOptions(":0", application, Options{MCPHandler: handler})
	unsigned := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`)) + "." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"owner"}`)) + "."
	for _, tc := range []struct {
		name, token string
		status      int
	}{
		{"verified", valid, 200},
		{"expired", token(issuerURL, "stuffstash-client", time.Now().Add(-time.Hour)), 401},
		{"wrong issuer", token("https://another-issuer.invalid", "stuffstash-client", time.Now().Add(time.Hour)), 401},
		{"wrong audience", token(issuerURL, "another-client", time.Now().Add(time.Hour)), 401},
		{"unsigned", unsigned, 401}, {"malformed", "not-a-jwt", 401}, {"dev token", "dev:owner", 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "https://stash.example.test/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
			request.Header.Set("Authorization", "Bearer "+tc.token)
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Accept", "application/json, text/event-stream")
			request.Header.Set("MCP-Protocol-Version", "2025-11-25")
			response := httptest.NewRecorder()
			server.Handler.ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("status %d want %d: %s", response.Code, tc.status, response.Body.String())
			}
			if tc.status == 401 && (strings.Contains(response.Body.String(), "list_tenants") || strings.Contains(response.Body.String(), tc.token)) {
				t.Fatal("unauthenticated response leaked protocol content or token")
			}
		})
	}
}

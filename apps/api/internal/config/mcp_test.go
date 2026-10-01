package config

import "testing"

func TestMCPRequiresExplicitMatchingAuthenticationAndSafePublicEndpoint(t *testing.T) {
	for _, tc := range []struct {
		name, enabled, mode, apiMode, endpoint, issuer string
		fail                                           bool
	}{
		{name: "disabled", apiMode: "local-dev"},
		{name: "bad enable", enabled: "sometimes", fail: true},
		{name: "no implicit local mode", enabled: "true", apiMode: "local-dev", endpoint: "http://localhost:8080/mcp", fail: true},
		{name: "mismatched mode", enabled: "true", mode: "local-dev", apiMode: "oidc", endpoint: "https://stash.example/mcp", fail: true},
		{name: "unknown mode", enabled: "true", mode: "apikey", apiMode: "apikey", endpoint: "https://stash.example/mcp", fail: true},
		{name: "local loopback", enabled: "true", mode: "local-dev", apiMode: "local-dev", endpoint: "http://127.0.0.1:8080/mcp"},
		{name: "local remote HTTP", enabled: "true", mode: "local-dev", apiMode: "local-dev", endpoint: "http://stash.example/mcp", fail: true},
		{name: "production HTTP", enabled: "true", mode: "oidc", apiMode: "oidc", endpoint: "http://localhost:8080/mcp", issuer: "https://issuer.example", fail: true},
		{name: "production missing issuer", enabled: "true", mode: "oidc", apiMode: "oidc", endpoint: "https://stash.example/mcp", fail: true},
		{name: "production HTTPS", enabled: "true", mode: "oidc", apiMode: "oidc", endpoint: "https://stash.example/mcp", issuer: "https://issuer.example"},
		{name: "credential URL", enabled: "true", mode: "local-dev", apiMode: "local-dev", endpoint: "https://user:pass@stash.example/mcp", fail: true},
		{name: "query URL", enabled: "true", mode: "local-dev", apiMode: "local-dev", endpoint: "https://stash.example/mcp?token=x", fail: true},
		{name: "incorrect path", enabled: "true", mode: "local-dev", apiMode: "local-dev", endpoint: "https://stash.example/api/mcp", fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("STUFF_STASH_MCP_ENABLED", tc.enabled)
			t.Setenv("STUFF_STASH_MCP_AUTH_MODE", tc.mode)
			t.Setenv("STUFF_STASH_MCP_PUBLIC_URL", tc.endpoint)
			result, err := LoadMCP(tc.apiMode, tc.issuer)
			if (err != nil) != tc.fail {
				t.Fatalf("error %v, want failure %v", err, tc.fail)
			}
			if err == nil && result.Enabled != (tc.enabled == "true") {
				t.Fatal("wrong enable state")
			}
		})
	}
}

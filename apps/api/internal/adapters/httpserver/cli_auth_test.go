package httpserver

import (
	"net/http"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/dto"
)

func TestCLIAuthDiscoveryIsPublicAndUnavailableWithoutConfiguration(t *testing.T) {
	for _, configured := range []bool{false, true} {
		options := Options{}
		if configured {
			options.CLIAuth = &dto.CLIAuthMetadata{Issuer: "https://issuer.example.test", ClientID: "cli", Scopes: []string{"openid"}, LoginMethods: []string{"authorization_code"}, LoopbackRedirect: dto.CLILoopbackRedirect{Host: "127.0.0.1", PathPrefix: "/callback/", EphemeralPort: true}}
		}
		server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), options)
		response := performRequest(server, http.MethodGet, "/auth/cli/config", "", nil)
		if !configured {
			if response.Code != http.StatusServiceUnavailable {
				t.Fatalf("disabled discovery returned %d", response.Code)
			}
			continue
		}
		if response.Code != http.StatusOK {
			t.Fatalf("discovery: %d %s", response.Code, response.Body.String())
		}
		var body struct {
			Data dto.CLIAuthMetadata `json:"data"`
		}
		decodeBody(t, response, &body)
		if body.Data.ClientID != "cli" || body.Data.Issuer != "https://issuer.example.test" || len(body.Data.LoginMethods) != 1 || body.Data.LoopbackRedirect.Host != "127.0.0.1" {
			t.Fatalf("unexpected public metadata: %+v", body.Data)
		}
	}
}

func coverCLIAuthScenario(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	options := Options{CLIAuth: &dto.CLIAuthMetadata{Issuer: "https://issuer.example.test", ClientID: "cli", Scopes: []string{"openid"}, LoginMethods: []string{"authorization_code"}, LoopbackRedirect: dto.CLILoopbackRedirect{Host: "127.0.0.1", PathPrefix: "/callback/", EphemeralPort: true}}}
	authorization := ""
	if adversarial {
		authorization = "Bearer malformed"
	}
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), options)
	response := coverage.request(t, server, http.MethodGet, "/auth/cli/config", "/auth/cli/config", authorization, nil, http.StatusOK)
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("auth discovery must not cache stale client settings")
	}
}

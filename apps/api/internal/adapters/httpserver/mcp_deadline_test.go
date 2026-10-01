package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/mcpserver"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestMCPDeadlineCancelsBlockedApplicationRead(t *testing.T) {
	access := &mcpBlockedAuthorizer{Authorizer: memory.NewAuthorizer(), entered: make(chan bool, 1)}
	application := newSeededTestAppWithAuthorizer(t, seededState{tenants: []seedTenant{{id: "home", name: "Home", owner: "owner"}}, inventories: []seedInventory{{id: "main", tenantID: "home", name: "Main", owner: "owner"}}}, access)
	handler, err := mcpserver.New(application, mcpserver.Options{PublicURL: "http://localhost/mcp", AuthMode: "local-dev", RequestTimeout: 200 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "http://localhost/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_asset","arguments":{"tenantId":"home","inventoryId":"main","assetId":"asset"}}}`))
	request.Header.Set("Authorization", "Bearer dev:owner")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	request.Header.Set("MCP-Protocol-Version", "2025-11-25")
	done := make(chan struct{})
	go func() { defer close(done); handler.ServeHTTP(httptest.NewRecorder(), request) }()
	select {
	case deadline := <-access.entered:
		if !deadline {
			t.Error("application received no deadline")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("tool never reached the application boundary")
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP request did not stop after its deadline")
	}
}

type mcpBlockedAuthorizer struct {
	ports.Authorizer
	entered chan bool
}

func (a *mcpBlockedAuthorizer) CheckInventory(ctx context.Context, _ identity.Principal, _ ports.InventoryPermission, _ inventory.InventoryID) error {
	_, hasDeadline := ctx.Deadline()
	a.entered <- hasDeadline
	<-ctx.Done()
	return ctx.Err()
}

func TestMCPBrowserCanReadAuthenticationDiscoveryChallenge(t *testing.T) {
	application, _ := mcpApplication(t)
	handler, err := mcpserver.New(application, mcpserver.Options{PublicURL: "https://stash.example.test/mcp", AuthMode: "local-dev", AllowedOrigins: []string{"https://web.example.test"}})
	if err != nil {
		t.Fatal(err)
	}
	server := NewServerWithOptions(":0", application, Options{MCPHandler: handler, CORSAllowedOrigins: []string{"https://web.example.test"}})
	request := httptest.NewRequest(http.MethodPost, "https://stash.example.test/mcp", strings.NewReader(`{}`))
	request.Header.Set("Origin", "https://web.example.test")
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	if response.Code != 401 || !strings.Contains(response.Header().Get("Access-Control-Expose-Headers"), "WWW-Authenticate") || !strings.Contains(response.Header().Get("WWW-Authenticate"), "resource_metadata=") {
		t.Fatalf("browser cannot discover authentication: %d %v", response.Code, response.Header())
	}
}

func TestMCPUsesSharedRateLimitAndPublicMetadata(t *testing.T) {
	application, _ := mcpApplication(t)
	handler, err := mcpserver.New(application, mcpserver.Options{PublicURL: "https://stash.example.test/mcp", AuthMode: "oidc", Issuer: "https://issuer.example.test"})
	if err != nil {
		t.Fatal(err)
	}
	server := NewServerWithOptions(":0", application, Options{MCPHandler: handler, RateLimiter: NewTokenBucketRateLimiter(1, time.Hour, 1)})
	request := httptest.NewRequest(http.MethodGet, "https://stash.example.test/.well-known/oauth-protected-resource/mcp", nil)
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"resource":"https://stash.example.test/mcp"`) || !strings.Contains(response.Body.String(), "https://issuer.example.test") {
		t.Fatalf("metadata: %d %s", response.Code, response.Body.String())
	}
	second := httptest.NewRecorder()
	server.Handler.ServeHTTP(second, request.Clone(context.Background()))
	if second.Code != 429 {
		t.Fatalf("MCP metadata escaped shared limiter: %d", second.Code)
	}
	disabled := NewServer(":0", application)
	hidden := httptest.NewRecorder()
	disabled.Handler.ServeHTTP(hidden, httptest.NewRequest(http.MethodGet, "https://stash.example.test/mcp", nil))
	if hidden.Code != 404 {
		t.Fatalf("disabled MCP endpoint published: %d", hidden.Code)
	}
}

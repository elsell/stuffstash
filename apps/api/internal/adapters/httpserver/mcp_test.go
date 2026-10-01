package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stuffstash/stuff-stash/internal/adapters/mcpserver"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func TestMCPClientReadsRESTAssetWithCurrentPrincipal(t *testing.T) {
	application, access := mcpApplication(t)
	api := NewServer(":0", application)
	created := performRequest(api, http.MethodPost, "/tenants/home/inventories/main/assets", "Bearer dev:owner", map[string]any{"kind": "item", "title": "MCP camping tent"})
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %s", created.Body.String())
	}
	assetID := decodeAsset(t, created).Data.ID
	parentCreated := performRequest(api, http.MethodPost, "/tenants/home/inventories/main/assets", "Bearer dev:owner", map[string]any{"kind": "location", "title": "Garage"})
	if parentCreated.Code != http.StatusCreated {
		t.Fatal(parentCreated.Body.String())
	}
	parentID := decodeAsset(t, parentCreated).Data.ID
	for n := 0; n < 2; n++ {
		result := performRequest(api, http.MethodPost, "/tenants/home/inventories/main/assets", "Bearer dev:owner", map[string]any{"kind": "item", "title": fmt.Sprintf("Garage tool %d", n), "parentAssetId": parentID})
		if result.Code != 201 {
			t.Fatal(result.Body.String())
		}
	}
	checkout := performRequest(api, http.MethodPost, "/tenants/home/inventories/main/assets/"+assetID+"/checkout", "Bearer dev:owner", map[string]any{"details": "Weekend camping"})
	if checkout.Code != 201 {
		t.Fatal(checkout.Body.String())
	}
	hosted := httptest.NewUnstartedServer(nil)
	endpoint := "http://" + hosted.Listener.Addr().String() + "/mcp"
	handler, err := mcpserver.New(application, mcpserver.Options{PublicURL: endpoint, AuthMode: "local-dev", MaxBodyBytes: 1024 * 1024})
	if err != nil {
		t.Fatal(err)
	}
	hosted.Config.Handler = NewServerWithOptions(":0", application, Options{MCPHandler: handler}).Handler
	hosted.Start()
	defer hosted.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, user := range []string{"viewer", "editor"} {
		t.Run(user, func(t *testing.T) {
			var token atomic.Value
			token.Store("Bearer dev:" + user)
			client := mcp.NewClient(&mcp.Implementation{Name: "mcp-gap-acceptance", Version: "1"}, nil)
			session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: endpoint, DisableStandaloneSSE: true, MaxRetries: -1, HTTPClient: &http.Client{Transport: mcpBearerTransport{token: &token}}}, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			listed, err := session.ListTools(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			names := []string{}
			for _, tool := range listed.Tools {
				names = append(names, tool.Name)
			}
			sort.Strings(names)
			want := []string{"get_asset", "list_asset_checkout_history", "list_checked_out_assets", "list_inventories", "list_location_assets", "list_root_assets", "list_tenants", "search_assets"}
			if strings.Join(names, ",") != strings.Join(want, ",") {
				t.Fatalf("catalog: %v", names)
			}
			result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "get_asset", Arguments: map[string]any{"tenantId": "home", "inventoryId": "main", "assetId": assetID}})
			if err != nil || result.IsError {
				t.Fatalf("get: %v %v", result, err)
			}
			encoded, _ := json.Marshal(result)
			if !bytes.Contains(encoded, []byte("MCP camping tent")) {
				t.Fatalf("missing asset: %s", encoded)
			}
			for _, tool := range []struct {
				name string
				args map[string]any
			}{
				{"list_tenants", map[string]any{}},
				{"list_inventories", map[string]any{"tenantId": "home"}},
				{"search_assets", map[string]any{"tenantId": "home", "inventoryId": "main", "query": "MCP camping tent", "mode": "exact"}},
				{"list_root_assets", map[string]any{"tenantId": "home", "inventoryId": "main", "limit": 1}},
				{"list_location_assets", map[string]any{"tenantId": "home", "inventoryId": "main", "assetId": parentID, "limit": 1}},
				{"list_checked_out_assets", map[string]any{"tenantId": "home", "inventoryId": "main"}},
				{"list_asset_checkout_history", map[string]any{"tenantId": "home", "inventoryId": "main", "assetId": assetID}},
			} {
				read, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tool.name, Arguments: tool.args})
				if err != nil || read.IsError {
					t.Fatalf("%s: %v %v", tool.name, read, err)
				}
				encoded, _ := json.Marshal(read.StructuredContent)
				var data struct {
					Items      []map[string]any
					Pagination struct {
						HasMore    bool
						NextCursor *string
					}
				}
				if err := json.Unmarshal(encoded, &data); err != nil {
					t.Fatal(err)
				}
				if len(data.Items) == 0 {
					t.Fatalf("%s returned no fixture records", tool.name)
				}
				if tool.name == "list_location_assets" {
					if len(data.Items) != 1 || data.Items[0]["parentAssetId"] != parentID || !data.Pagination.HasMore || data.Pagination.NextCursor == nil {
						t.Fatalf("parent pagination: %s", encoded)
					}
					tool.args["cursor"] = *data.Pagination.NextCursor
					next, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tool.name, Arguments: tool.args})
					if err != nil || next.IsError {
						t.Fatalf("second child page: %v %v", next, err)
					}
					reused, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_root_assets", Arguments: map[string]any{"tenantId": "home", "inventoryId": "main", "cursor": *data.Pagination.NextCursor}})
					if err == nil && !reused.IsError {
						t.Fatal("parent cursor was reusable against roots")
					}
				}
			}
			for _, denied := range []struct {
				name string
				args map[string]any
			}{
				{"create_asset", map[string]any{"tenantId": "home", "inventoryId": "main", "title": "Unauthorized write"}},
				{"get_asset", map[string]any{"tenantId": "other", "inventoryId": "secret", "assetId": assetID}},
				{"get_asset", map[string]any{"tenantId": "home", "inventoryId": "secret", "assetId": assetID}},
				{"list_location_assets", map[string]any{"tenantId": "home", "inventoryId": "main", "assetId": assetID}},
				{"get_asset", map[string]any{"tenantId": "home", "inventoryId": "main", "assetId": assetID, "execute": "delete"}},
				{"list_root_assets", map[string]any{"tenantId": "home", "inventoryId": "main", "limit": 0}},
				{"get_asset", map[string]any{"tenantId": "home", "inventoryId": "main"}},
			} {
				result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: denied.name, Arguments: denied.args})
				if err == nil && !result.IsError {
					t.Fatalf("accepted invalid or unauthorized call %s", denied.name)
				}
			}
			if user == "viewer" {
				if err := access.RevokeInventoryViewer(ctx, principal("viewer"), "home", "main"); err != nil {
					t.Fatal(err)
				}
				denied, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "get_asset", Arguments: map[string]any{"tenantId": "home", "inventoryId": "main", "assetId": assetID}})
				if err == nil && !denied.IsError {
					t.Fatal("revoked permission retained access")
				}
			}
			token.Store("Bearer dev:stranger") // The same logical client cannot retain the old principal.
			result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "get_asset", Arguments: map[string]any{"tenantId": "home", "inventoryId": "main", "assetId": assetID}})
			if err == nil && !result.IsError {
				t.Fatal("changed principal retained access")
			}
			encoded, _ = json.Marshal(result)
			if bytes.Contains(encoded, []byte("MCP camping tent")) {
				t.Fatal("denied result leaked inventory content")
			}
		})
	}
	records := performRequest(api, http.MethodGet, "/tenants/home/inventories/main/assets/"+assetID+"/audit-records", "Bearer dev:owner", nil)
	if records.Code != 200 || !strings.Contains(records.Body.String(), `"source":"mcp"`) {
		t.Fatalf("MCP read audit missing: %s", records.Body.String())
	}

}

func TestMCPRejectsUnsafeTransportAndUnapprovedTools(t *testing.T) {
	application, _ := mcpApplication(t)
	handler, err := mcpserver.New(application, mcpserver.Options{PublicURL: "https://stash.example.test/mcp", AuthMode: "local-dev", MaxBodyBytes: 1024})
	if err != nil {
		t.Fatal(err)
	}
	server := NewServerWithOptions(":0", application, Options{MCPHandler: handler})
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`
	for _, tc := range []struct {
		name, token, host, origin, body string
		status                          int
	}{
		{"missing bearer", "", "stash.example.test", "", body, 401},
		{"malformed bearer", "Bearer invalid", "stash.example.test", "", body, 401},
		{"untrusted origin", "Bearer dev:owner", "stash.example.test", "https://evil.invalid", body, 403},
		{"null origin", "Bearer dev:owner", "stash.example.test", "null", body, 403},
		{"rebinding host", "Bearer dev:owner", "evil.invalid", "", body, 403},
		{"body cap", "Bearer dev:owner", "stash.example.test", "", strings.Repeat("x", 1025), 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "https://"+tc.host+"/mcp", strings.NewReader(tc.body))
			request.Header.Set("Authorization", tc.token)
			if tc.origin != "" {
				request.Header.Set("Origin", tc.origin)
			}
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Accept", "application/json, text/event-stream")
			request.Header.Set("MCP-Protocol-Version", "2025-11-25")
			response := httptest.NewRecorder()
			server.Handler.ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", response.Code, tc.status, response.Body.String())
			}
			if tc.status == 401 && !strings.Contains(response.Header().Get("WWW-Authenticate"), "oauth-protected-resource/mcp") {
				t.Fatal("missing resource discovery challenge")
			}
		})
	}
}

func mcpApplication(t *testing.T) (app.App, *memory.Authorizer) {
	t.Helper()
	access := memory.NewAuthorizer()
	application := newSeededTestAppWithAuthorizer(t, seededState{tenants: []seedTenant{{id: "home", name: "Home", owner: "owner"}, {id: "other", name: "Hidden", owner: "other"}}, inventories: []seedInventory{{id: "main", tenantID: "home", name: "Inventory", owner: "owner"}, {id: "secret", tenantID: "other", name: "Hidden", owner: "other"}}}, access)
	if err := access.GrantInventoryViewer(context.Background(), principal("viewer"), "home", "main"); err != nil {
		t.Fatal(err)
	}
	if err := access.GrantInventoryEditor(context.Background(), principal("editor"), "home", "main"); err != nil {
		t.Fatal(err)
	}
	return application, access
}

type mcpBearerTransport struct{ token *atomic.Value }

func (t mcpBearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", t.token.Load().(string))
	return http.DefaultTransport.RoundTrip(r)
}

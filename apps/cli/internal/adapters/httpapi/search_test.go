package httpapi

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchPreservesExplicitScopeFiltersAndResult(t *testing.T) {
	expectedInventory := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tenants/home/search/assets" || r.Header.Get("Authorization") != "Bearer owner" || r.URL.Query().Get("inventoryId") == "denied" {
			w.WriteHeader(403)
			return
		}
		if r.URL.Query().Get("inventoryId") != expectedInventory {
			t.Error("inventory scope changed")
		}
		for key, want := range map[string]string{"q": "tea & coffee", "mode": "exact", "tagIds": "a,b", "customAssetTypeId": "type", "lifecycleState": "all", "checkoutState": "available", "limit": "2", "cursor": "next"} {
			if r.URL.Query().Get(key) != want {
				t.Errorf("filter %s lost: %s", key, r.URL)
			}
		}
		io.WriteString(w, `{"$schema":"schema","data":[{"type":"asset","tenantId":"home","inventory":{"id":"tools","name":"Tools"},"asset":{"id":"asset","inventoryId":"tools","title":"Coffee","customFields":{"serial":9007199254740993},"tags":[],"expiration":{"date":"2026-02","precision":"month"}},"matches":[{"field":"title","value":"Coffee"}],"ancestorPath":[{"id":"kitchen","title":"Kitchen"}]}],"meta":{"requestId":"trace","pagination":{"limit":2,"hasMore":true,"nextCursor":"more"}}}`)
	}))
	defer server.Close()
	for _, a := range []struct {
		token, tenant, inventory string
		allowed                  bool
	}{{"owner", "home", "tools", true}, {"owner", "home", "", true}, {"", "home", "tools", false}, {"denied", "home", "tools", false}, {"owner", "other", "tools", false}, {"owner", "home", "denied", false}} {
		expectedInventory = a.inventory
		api, _ := New(server.URL, a.token, server.Client())
		result, err := api.SearchAssets(context.Background(), ports.Scope{Tenant: a.tenant, Inventory: a.inventory}, ports.SearchQuery{Query: "tea & coffee", Mode: "exact", TagIDs: []string{"a", "b"}, TypeID: "type", Lifecycle: "all", CheckoutState: "available", Page: ports.Page{Limit: 2, Cursor: "next"}})
		if (err == nil) != a.allowed {
			t.Fatalf("scope: %+v %v", a, err)
		}
		if a.allowed {
			b, _ := json.Marshal(result)
			for _, want := range []string{`9007199254740993`, `"name":"Tools"`, `"field":"title"`, `"ancestorPath":[{"id":"kitchen","title":"Kitchen"}]`, `"nextCursor":"more"`, `"requestId":"trace"`} {
				if !strings.Contains(string(b), want) {
					t.Fatalf("lost %s in %s", want, b)
				}
			}
		}
	}
}

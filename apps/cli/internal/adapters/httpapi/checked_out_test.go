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

func TestCheckedOutAssetsPreserveDataAndScope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tenants/home/inventories/tools/checked-out-assets" || r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		if r.URL.Query().Get("limit") != "2" || r.URL.Query().Get("cursor") != "next" {
			t.Error("pagination lost")
		}
		io.WriteString(w, `{"$schema":"schema","data":[{"asset":{"id":"asset","title":"Drill","tags":[],"expiration":null,"customFields":{"serial":9007199254740993}},"checkout":{"id":"checkout","state":"checked_out","checkedOutAt":"today","checkedOutByPrincipalId":"owner","checkedOutByPrincipal":{"id":"owner","email":"owner@example.test"}}}],"meta":{"requestId":"trace","pagination":{"limit":2,"hasMore":true,"nextCursor":"more"}}}`)
	}))
	defer server.Close()
	for _, a := range []struct {
		token, tenant, inventory string
		allowed                  bool
	}{{"owner", "home", "tools", true}, {"", "home", "tools", false}, {"denied", "home", "tools", false}, {"owner", "elsewhere", "tools", false}, {"owner", "home", "elsewhere", false}} {
		api, _ := New(server.URL, a.token, server.Client())
		result, err := api.CheckedOutAssets(context.Background(), ports.Scope{Tenant: a.tenant, Inventory: a.inventory}, ports.Page{Limit: 2, Cursor: "next"})
		if (err == nil) != a.allowed {
			t.Fatalf("scope: %+v %v", a, err)
		}
		if a.allowed {
			b, _ := json.Marshal(result)
			for _, field := range []string{`9007199254740993`, `"email":"owner@example.test"`, `"nextCursor":"more"`, `"requestId":"trace"`, `"$schema":"schema"`} {
				if !strings.Contains(string(b), field) {
					t.Fatalf("lost %s in %s", field, b)
				}
			}
		}
	}
}

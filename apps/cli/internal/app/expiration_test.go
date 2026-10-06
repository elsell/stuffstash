package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExpirationFilterValidation(t *testing.T) {
	o, err := Parse([]string{"assets", "expiration", "--mode", "expired", "--tag-id", "a", "--tag-id", "b", "--from-date", "2026-01-01", "--through-date", "2026-12-31"}, func(string) string { return "" })
	if err != nil || len(o.Expiration.TagIDs) != 2 {
		t.Fatalf("filters lost: %+v %v", o, err)
	}
	for _, args := range [][]string{{"assets", "expiration", "--mode", "bad"}, {"assets", "expiration", "--from-date", "2026-02-30"}, {"assets", "expiration", "--from-date", "2026-12-31", "--through-date", "2026-01-01"}, {"assets", "expiration", "--limit", "101"}, {"assets", "list", "--mode", "expired"}} {
		if _, err := Parse(args, func(string) string { return "" }); err == nil {
			t.Fatalf("invalid filter accepted: %v", args)
		}
	}
}

func TestExpirationCommandPreservesFiltersDataAndScope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tenants/home/inventories/tools/expiration-assets" || r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		for key, want := range map[string]string{"mode": "expired", "kind": "item", "checkoutState": "available", "q": "tea & coffee", "customAssetTypeId": "type", "tagIds": "a,b", "locationId": "kitchen", "fromDate": "2026-01-01", "throughDate": "2026-12-31", "limit": "2", "cursor": "next"} {
			if r.URL.Query().Get(key) != want {
				t.Errorf("filter %s lost: %s", key, r.URL)
			}
		}
		io.WriteString(w, `{"$schema":"schema","data":{"counts":{"all":3,"expired":2,"soon":1},"timezone":"America/New_York","items":[{"id":"asset","title":"Coffee","tags":[],"expiration":{"date":"2026-02","precision":"month"},"customFields":{"serial":9007199254740993},"ancestorPath":[{"id":"kitchen","title":"Kitchen"}]}]},"meta":{"requestId":"trace","pagination":{"limit":2,"hasMore":true,"nextCursor":"more"}}}`)
	}))
	defer server.Close()
	o, err := Parse([]string{"assets", "expiration", "--mode", "expired", "--kind", "item", "--checkout-state", "available", "--query", "tea & coffee", "--type-id", "type", "--tag-id", "a", "--tag-id", "b", "--location-id", "kitchen", "--from-date", "2026-01-01", "--through-date", "2026-12-31", "--limit", "2", "--cursor", "next", "--tenant", "home", "--inventory", "tools"}, func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range []struct {
		token, tenant, inventory string
		allowed                  bool
	}{{"owner", "home", "tools", true}, {"", "home", "tools", false}, {"denied", "home", "tools", false}, {"owner", "other", "tools", false}, {"owner", "home", "other", false}} {
		o.Scope.Tenant = a.tenant
		o.Scope.Inventory = a.inventory
		if err := validateCommand(o); err != nil {
			t.Fatal(err)
		}
		api, err := httpapi.New(server.URL, a.token, server.Client())
		if err != nil {
			t.Fatal(err)
		}
		result, err := execute(context.Background(), api, o)
		if (err == nil) != a.allowed {
			t.Fatalf("scope: %+v %v", a, err)
		}
		if a.allowed {
			b, _ := json.Marshal(result)
			for _, want := range []string{`9007199254740993`, `"expired":2`, `"timezone":"America/New_York"`, `"ancestorPath":[{"id":"kitchen","title":"Kitchen"}]`, `"nextCursor":"more"`, `"requestId":"trace"`, `"$schema":"schema"`} {
				if !strings.Contains(string(b), want) {
					t.Fatalf("lost %s in %s", want, b)
				}
			}
		}
	}
}

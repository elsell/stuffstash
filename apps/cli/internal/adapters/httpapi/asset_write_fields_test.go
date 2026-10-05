package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAssetFullWriteBodiesAndScope(t *testing.T) {
	for _, action := range []string{"create", "update"} {
		t.Run(action, func(t *testing.T) {
			body := `{"title":"Drill","kind":"item","description":"Cordless","tagIds":[],"customAssetTypeId":"type","customFields":{"serial":9007199254740993},"expiration":{"date":"2027-01","precision":"month"},"parentAssetId":"shelf"}`
			method, path := http.MethodPost, "/tenants/home/inventories/tools/assets"
			if action == "update" {
				body = `{"description":"","tagIds":[],"customFields":{"serial":9007199254740993},"expiration":null,"parentAssetId":null}`
				method = http.MethodPatch
				path += "/asset"
			}
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer owner" || r.URL.Path != path {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				writes++
				got, _ := io.ReadAll(r.Body)
				if r.Method != method || string(got) != body || r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("request changed: %s %s", r.Method, got)
				}
				wantKey := ""
				if action == "create" {
					wantKey = "retry"
				}
				if r.Header.Get("Idempotency-Key") != wantKey {
					t.Error("wrong retry header")
				}
				w.Header().Set("Content-Type", "application/json")
				io.WriteString(w, `{"data":{"id":"asset","title":"Drill","tags":[],"expiration":null},"meta":{}}`)
			}))
			defer server.Close()
			for _, attempt := range []struct {
				token, tenant, inventory string
				allowed                  bool
			}{{"owner", "home", "tools", true}, {"", "home", "tools", false}, {"other", "home", "tools", false}, {"owner", "elsewhere", "tools", false}, {"owner", "home", "elsewhere", false}} {
				api, _ := New(server.URL, attempt.token, server.Client())
				scope := ports.Scope{Tenant: attempt.tenant, Inventory: attempt.inventory}
				var err error
				if action == "create" {
					_, err = api.CreateAsset(context.Background(), scope, ports.AssetInput{RequestBody: []byte(body)}, "retry")
				} else {
					_, err = api.UpdateAsset(context.Background(), scope, "asset", ports.AssetChange{RequestBody: []byte(body)}, "")
				}
				if (err == nil) != attempt.allowed {
					t.Fatalf("authorization result: %+v %v", attempt, err)
				}
			}
			if writes != 1 {
				t.Fatalf("unexpected writes: %d", writes)
			}
		})
	}
}

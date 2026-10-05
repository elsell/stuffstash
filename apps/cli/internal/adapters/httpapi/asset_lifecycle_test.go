package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAssetLifecycleScopeAndAuthorization(t *testing.T) {
	for _, action := range []string{"archive", "restore", "delete"} {
		t.Run(action, func(t *testing.T) {
			writes := 0
			path := "/tenants/home/inventories/tools/assets/asset"
			method := "DELETE"
			if action != "delete" {
				path += "/" + action
				method = "PATCH"
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer owner" || r.URL.Path != path {
					w.WriteHeader(403)
					return
				}
				if r.Method != method || r.Header.Get("Idempotency-Key") != "" {
					t.Errorf("wrong lifecycle request: %s", r.Method)
				}
				writes++
				if action == "delete" {
					w.WriteHeader(204)
					return
				}
				io.WriteString(w, `{"data":{"id":"asset","tags":[],"expiration":null,"undoableOperationId":"undo"},"meta":{"requestId":"trace"}}`)
			}))
			defer server.Close()
			for _, attempt := range []struct {
				token, tenant, inventory string
				allowed                  bool
			}{{"owner", "home", "tools", true}, {"", "home", "tools", false}, {"viewer", "home", "tools", false}, {"owner", "other", "tools", false}, {"owner", "home", "other", false}} {
				api, _ := New(server.URL, attempt.token, server.Client())
				scope := ports.Scope{Tenant: attempt.tenant, Inventory: attempt.inventory}
				var err error
				if action == "delete" {
					err = api.DeleteAsset(context.Background(), scope, "asset")
				} else {
					var result ports.Result[ports.Asset]
					result, err = api.SetArchived(context.Background(), scope, "asset", action == "archive", "")
					if attempt.allowed && (result.Data.UndoableOperationID == nil || result.Meta == nil) {
						t.Fatal("lifecycle response lost")
					}
				}
				if (err == nil) != attempt.allowed {
					t.Fatalf("authorization: %+v %v", attempt, err)
				}
			}
			if writes != 1 {
				t.Fatalf("writes=%d", writes)
			}
		})
	}
}

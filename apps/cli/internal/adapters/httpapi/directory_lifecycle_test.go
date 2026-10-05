package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDirectoryLifecycleRoutesAndNoContent(t *testing.T) {
	for _, resource := range []ports.DirectoryResource{ports.HouseholdResource, ports.InventoryResource} {
		for _, action := range []ports.LifecycleAction{ports.Archive, ports.Restore, ports.Delete} {
			t.Run(string(resource)+string(action), func(t *testing.T) {
				path := "/tenants/home"
				if resource == ports.InventoryResource {
					path += "/inventories/garage"
				}
				method := "DELETE"
				if action != ports.Delete {
					method = "PATCH"
					path += "/" + string(action)
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("Authorization") != "Bearer owner" {
						w.WriteHeader(403)
						io.WriteString(w, `{"error":{"message":"private detail"}}`)
						return
					}
					if r.Method != method || r.URL.Path != path {
						t.Errorf("wrong request: %s %s", r.Method, r.URL.Path)
					}
					if action == ports.Delete {
						w.WriteHeader(204)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					io.WriteString(w, `{"data":{"id":"garage","tenantId":"home","name":"Garage","lifecycleState":"active","access":{}},"meta":{"requestId":"trace"}}`)
				}))
				defer server.Close()
				for _, token := range []string{"owner", "viewer", ""} {
					api, _ := New(server.URL, token, server.Client())
					_, err := api.ChangeDirectoryLifecycle(context.Background(), resource, action, ports.Scope{Tenant: "home", Inventory: "garage"})
					if token == "owner" && err != nil {
						t.Fatal(err)
					}
					if token != "owner" && (err == nil || strings.Contains(err.Error(), "private")) {
						t.Fatalf("unsafe denial: %v", err)
					}
				}
			})
		}
	}
}

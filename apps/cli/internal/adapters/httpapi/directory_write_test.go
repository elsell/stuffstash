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

func TestDirectoryWritesPreserveScopeBodyAndDenyUnauthorized(t *testing.T) {
	body := `{"name":"Garage"}`
	for _, test := range []struct {
		kind         ports.DirectoryWrite
		method, path string
	}{
		{ports.CreateTenant, "POST", "/tenants"},
		{ports.UpdateTenant, "PATCH", "/tenants/home"},
		{ports.CreateInventory, "POST", "/tenants/home/inventories"},
		{ports.UpdateInventory, "PATCH", "/tenants/home/inventories/garage"},
	} {
		t.Run(test.path+test.method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					io.WriteString(w, `{"error":{"message":"private reason"}}`)
					return
				}
				if r.URL.Path != test.path || r.Method != test.method {
					t.Errorf("wrong target: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(403)
					return
				}
				got, _ := io.ReadAll(r.Body)
				if string(got) != body || r.Header.Get("Idempotency-Key") != "" || r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("request lost data: %s %v", got, r.Header)
				}
				io.WriteString(w, `{"data":{"id":"garage","tenantId":"home","name":"Garage","lifecycleState":"active","access":{"relationship":"owner","permissions":["write"]}},"meta":{}}`)
			}))
			defer server.Close()
			for _, token := range []string{"owner", "viewer", ""} {
				api, _ := New(server.URL, token, server.Client())
				_, err := api.WriteDirectory(context.Background(), test.kind, ports.Scope{Tenant: "home", Inventory: "garage"}, []byte(body))
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

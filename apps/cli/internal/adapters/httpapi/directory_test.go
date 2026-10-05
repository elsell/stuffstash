package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestDirectoryReadsPreserveFieldsAndRejectUnauthorizedScope(t *testing.T) {
	bodies := map[string]string{
		"/me":                              `{"data":{"id":"alice","displayName":"Alice","email":"alice@example.test"},"meta":{}}`,
		"/tenants/home":                    `{"data":{"id":"home","name":"Home","lifecycleState":"active","access":{"relationship":"owner","permissions":["read","write"]}},"meta":{}}`,
		"/tenants/home/inventories/garage": `{"data":{"id":"garage","tenantId":"home","name":"Garage","lifecycleState":"active","access":{"relationship":"editor","permissions":["read"]}},"meta":{}}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, ok := bodies[r.URL.Path]
		if r.Header.Get("Authorization") != "Bearer alice" || !ok {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, `{"error":{"message":"private detail"}}`)
			return
		}
		io.WriteString(w, body)
	}))
	defer server.Close()
	for _, c := range []struct {
		name  string
		read  func(*Client, ports.Scope) (any, error)
		field string
	}{
		{"account", func(c *Client, _ ports.Scope) (any, error) { return c.Principal(context.Background()) }, `"email":"alice@example.test"`},
		{"tenant", func(c *Client, s ports.Scope) (any, error) { return c.Tenant(context.Background(), s) }, `"permissions":["read","write"]`},
		{"inventory", func(c *Client, s ports.Scope) (any, error) { return c.Inventory(context.Background(), s) }, `"tenantId":"home"`},
	} {
		t.Run(c.name, func(t *testing.T) {
			client, _ := New(server.URL, "alice", server.Client())
			result, err := c.read(client, ports.Scope{Tenant: "home", Inventory: "garage"})
			data, _ := json.Marshal(result)
			if err != nil || !strings.Contains(string(data), c.field) {
				t.Fatalf("lost fields: %s %v", data, err)
			}
			for _, token := range []string{"", "bob"} {
				denied, _ := New(server.URL, token, server.Client())
				if _, err := c.read(denied, ports.Scope{Tenant: "home", Inventory: "garage"}); err == nil || strings.Contains(err.Error(), "private") {
					t.Fatalf("unsafe denied read: %v", err)
				}
			}
			if c.name != "account" {
				if _, err := c.read(client, ports.Scope{Tenant: "other", Inventory: "garage"}); err == nil {
					t.Fatal("cross-household request accepted")
				}
			}
		})
	}
}

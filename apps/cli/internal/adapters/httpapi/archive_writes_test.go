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

func TestArchiveWritesKeepInputsAndScope(t *testing.T) {
	for _, action := range []string{"create", "approve"} {
		t.Run(action, func(t *testing.T) {
			body := `{"inventoryId":"garage","photos":false,"otherFiles":true}`
			if action == "approve" {
				body = `{"name":"Restored inventory"}`
			}
			calls := 0
			fail := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				path := "/tenants/home/archive-jobs"
				if action == "approve" {
					path += "/job/approve"
				}
				if r.URL.Path != path || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if r.Method != "POST" {
					t.Error("wrong method")
				}
				raw, err := io.ReadAll(r.Body)
				if err != nil || string(raw) != body {
					t.Error("changed body")
				}
				if action == "create" && r.Header.Get("Idempotency-Key") != "stable-key" {
					t.Error("lost retry key")
				}
				if action == "approve" && r.URL.Query().Get("inventoryId") != "garage" {
					w.WriteHeader(403)
					return
				}
				if fail {
					w.WriteHeader(503)
					io.WriteString(w, "private-error")
					return
				}
				io.WriteString(w, `{"data":{"id":"job","createdAt":"2026-10-05T12:00:00Z","expiresAt":"2026-10-06T12:00:00Z","kind":"restore","state":"queued","phase":"pending","photos":false,"otherFiles":true,"destinationInventoryId":"new"},"meta":{"requestId":"trace"}}`)
			}))
			defer server.Close()
			client, err := New(server.URL, "owner", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			call := func(s ports.Scope) (ports.Result[ports.ArchiveJob], error) {
				if action == "create" {
					return client.CreateArchiveJob(context.Background(), s.Tenant, "stable-key", []byte(body))
				}
				return client.ApproveArchiveJob(context.Background(), s, "job", []byte(body))
			}
			result, err := call(ports.Scope{Tenant: "home", Inventory: "garage"})
			if err != nil || result.Data.ID != "job" || result.Data.Photos || !result.Data.OtherFiles || result.Data.DestinationInventoryID == nil || result.Meta == nil {
				t.Fatalf("archive: %+v %v", result, err)
			}
			if _, err := call(ports.Scope{Tenant: "other", Inventory: "garage"}); err == nil {
				t.Fatal("household bypass")
			}
			if action == "approve" {
				if _, err := call(ports.Scope{Tenant: "home", Inventory: "other"}); err == nil {
					t.Fatal("inventory bypass")
				}
			}
			fail = true
			before := calls
			if _, err := call(ports.Scope{Tenant: "home", Inventory: "garage"}); err == nil || strings.Contains(err.Error(), "private-error") || calls != before+1 {
				t.Fatal("unsafe error or retry")
			}
		})
	}
}

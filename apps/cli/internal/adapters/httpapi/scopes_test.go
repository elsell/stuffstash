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

func TestTenantChoicesUseAuthenticatedPaginatedSDKRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "Bearer allowed" {
			w.WriteHeader(403)
			io.WriteString(w, `{"error":{"message":"private backend detail"}}`)
			return
		}
		if r.URL.Path != "/me/tenants" || r.URL.Query().Get("cursor") != "next" || r.URL.Query().Get("limit") != "100" {
			t.Errorf("wrong request: %s", r.URL.String())
		}
		io.WriteString(w, `{"data":[{"id":"home","name":"Home","lifecycleState":"active","access":{"relationship":"owner","permissions":["read","write"]}}],"meta":{"pagination":{"limit":100,"hasMore":false}}}`)
	}))
	defer server.Close()
	client, _ := New(server.URL, "allowed", server.Client())
	got, err := client.Tenants(context.Background(), ports.Page{Limit: 100, Cursor: "next"})
	if err != nil || len(got.Data) != 1 || got.Data[0].ID != "home" || got.Data[0].Name != "Home" || got.Data[0].Access.Relationship != "owner" || len(got.Data[0].Access.Permissions) != 2 {
		t.Fatalf("choices: %+v %v", got, err)
	}
	denied, _ := New(server.URL, "denied", server.Client())
	if _, err := denied.Tenants(context.Background(), ports.Page{}); err == nil || strings.Contains(err.Error(), "private") {
		t.Fatalf("unsafe denial: %v", err)
	}
}

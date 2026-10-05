package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestTenantListCommandAuthenticatedDiscoveryAndDenial(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/tenants" || r.URL.Query().Get("limit") != "2" || r.URL.Query().Get("cursor") != "page-two" {
			t.Errorf("unexpected request: %s", r.URL)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "Bearer allowed" {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, `{"error":{"message":"private backend detail"}}`)
			return
		}
		io.WriteString(w, `{"data":[{"id":"home","name":"Home","lifecycleState":"active","access":{"relationship":"owner","permissions":["read"]}}],"meta":{"pagination":{"limit":2,"hasMore":true,"nextCursor":"page-three"}}}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	store := credentials.File{Path: filepath.Join(dir, "session.json")}
	getenv := func(key string) string {
		switch key {
		case "STUFF_STASH_CLI_SERVER":
			return server.URL
		case "STUFF_STASH_CLI_CREDENTIAL_FILE":
			return store.Path
		case "STUFF_STASH_CLI_CONFIG_FILE":
			return filepath.Join(dir, "config", "contexts.json")
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		}
		return ""
	}
	manager := contexts.Manager{Store: contextfile.Store{Path: filepath.Join(dir, "config", "contexts.json")}}
	principal := contexts.Principal(ports.Session{Issuer: "https://identity.example", Subject: "allowed"})
	for _, entry := range []contexts.Entry{
		{Name: "first", Server: server.URL, Principal: principal, Tenant: "first"},
		{Name: "second", Server: server.URL, Principal: principal, Tenant: "second"},
		{Name: "other-server", Server: "https://other.example", Principal: principal, Tenant: "other"},
	} {
		if err := manager.Remember(context.Background(), entry); err != nil {
			t.Fatal(err)
		}
	}
	for _, token := range []string{"allowed", "denied"} {
		if err := store.Save(context.Background(), ports.Session{Server: server.URL, IDToken: token, Issuer: "https://identity.example", Subject: token, ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
		var out, diagnostic bytes.Buffer
		code := Run(context.Background(), []string{"tenants", "list", "--json", "--no-input", "--limit", "2", "--cursor", "page-two"}, getenv, &out, &diagnostic)
		if token == "denied" {
			if code != 1 || out.Len() != 0 || !strings.Contains(diagnostic.String(), "forbidden") || strings.Contains(diagnostic.String(), "private backend") {
				t.Fatalf("unsafe denial: %d %s %s", code, &out, &diagnostic)
			}
			continue
		}
		var result ports.Result[[]ports.Tenant]
		if code != 0 || diagnostic.Len() != 0 || json.Unmarshal(out.Bytes(), &result) != nil || len(result.Data) != 1 || result.Data[0].Access.Relationship != "owner" || result.Pagination == nil || result.Pagination.NextCursor == nil || *result.Pagination.NextCursor != "page-three" {
			t.Fatalf("discovery failed: %d %s %s", code, &out, &diagnostic)
		}
	}
}

package bootstrap

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAuditCommandsRespectScopeAndPagination(t *testing.T) {
	for _, resource := range []string{"tenants", "inventories", "assets"} {
		t.Run(resource, func(t *testing.T) {
			calls := 0
			path := "/tenants/home"
			if resource != "tenants" {
				path += "/inventories/garage"
			}
			if resource == "assets" {
				path += "/assets/asset"
			}
			path += "/audit-records"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != path || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if r.Method != "GET" || r.URL.Query().Get("limit") != "2" {
					t.Error("wrong query")
				}
				if resource != "assets" && r.URL.Query().Get("cursor") != "next" {
					t.Error("cursor lost")
				}
				io.WriteString(w, `{"$schema":"schema","data":[{"id":"record","tenantId":"home","inventoryId":"garage","principalId":"owner","principal":{"id":"owner","email":"owner@example.test"},"requestId":"original","action":"asset.updated","source":"cli","targetType":"asset","targetId":"asset","occurredAt":"2026-10-05T12:00:00Z","metadata":{"changed":"title"}}],"meta":{"requestId":"trace","pagination":{"limit":2,"hasMore":true,"nextCursor":"more"}}}`)
			}))
			defer server.Close()
			dir := t.TempDir()
			os.Chmod(dir, 0700)
			store := credentials.File{Path: filepath.Join(dir, "session.json")}
			if err := store.Save(context.Background(), ports.Session{Server: server.URL, Issuer: "https://id.example", Subject: "owner", IDToken: "owner", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			getenv := func(k string) string {
				switch k {
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

			command := []string{resource, "audit"}
			if resource == "assets" {
				command = append(command, "asset")
			} else {
				command = append(command, "--cursor", "next")
			}
			command = append(command, "--tenant", "home", "--limit", "2", "--json", "--no-input")
			if resource != "tenants" {
				command = append(command, "--inventory", "garage")
			}
			var out, diagnostic bytes.Buffer
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 {
				t.Fatalf("audit: %d %d %s", code, calls, &diagnostic)
			}
			for _, field := range []string{`"email":"owner@example.test"`, `"requestId":"original"`, `"changed":"title"`, `"nextCursor":"more"`} {
				if !strings.Contains(out.String(), field) {
					t.Fatalf("audit field lost: %s", &out)
				}
			}
			out.Reset()
			diagnostic.Reset()
			if code := Run(context.Background(), append(command, "--tenant", "other"), getenv, &out, &diagnostic); code == 0 {
				t.Fatal("wrong household succeeded")
			}
			if resource != "tenants" {
				if code := Run(context.Background(), append(command, "--inventory", "other"), getenv, &out, &diagnostic); code == 0 {
					t.Fatal("wrong inventory succeeded")
				}
			}
			if resource == "assets" {
				before := calls
				if code := Run(context.Background(), append(command, "--cursor", "ignored"), getenv, &out, &diagnostic); code != 2 || calls != before {
					t.Fatal("asset cursor ignored")
				}
			}
		})
	}
}

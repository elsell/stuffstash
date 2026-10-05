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

func TestInvitationMutationsRespectScopeAndConfirmation(t *testing.T) {
	for _, action := range []string{"cancel", "delete"} {
		t.Run(action, func(t *testing.T) {
			calls := 0
			reading := false
			listing := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				path := "/tenants/home/inventories/garage/access-invitations/invite"
				method := "DELETE"
				if action == "cancel" && !reading {
					path += "/cancel"
					method = "PATCH"
				}
				if reading {
					method = "GET"
					if listing {
						path = "/tenants/home/inventories/garage/access-invitations"
					}
				}
				if r.URL.Path != path || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if r.Method != method {
					t.Error("wrong method")
				}
				if reading {
					data := `{"id":"invite","tenantId":"home","inventoryId":"garage","email":"guest@example.test","relationship":"viewer","status":"accepted","isExpired":false,"expiresAt":"2030-01-01T00:00:00Z","inviterPrincipalId":"owner","acceptedPrincipalId":"guest"}`
					if listing {
						if r.URL.Query().Get("status") != "all" || r.URL.Query().Get("cursor") != "next" || r.URL.Query().Get("limit") != "2" {
							t.Error("lost query")
						}
						data = "[" + data + "]"
					}
					io.WriteString(w, `{"data":`+data+`,"meta":{"requestId":"trace","pagination":{"limit":2,"hasMore":true,"nextCursor":"more"}}}`)
					return
				}
				w.WriteHeader(204)
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

			scope := []string{"--tenant", "home", "--inventory", "garage", "--json", "--no-input"}
			command := append([]string{"invitations", action, "invite"}, scope...)
			var out, diagnostic bytes.Buffer
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || calls != 0 {
				t.Fatal("unconfirmed operation reached API")
			}
			command = append(command, "--yes")
			out.Reset()
			diagnostic.Reset()
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 {
				t.Fatalf("operation: %d %d %s %s", code, calls, &out, &diagnostic)
			}
			for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
				if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
					t.Fatal("scope boundary bypassed")
				}
			}
			reading = true
			for _, mode := range []string{"show", "list"} {
				listing = mode == "list"
				readCommand := append([]string{"invitations", "show", "invite"}, scope...)
				if listing {
					readCommand = append([]string{"invitations", "list", "--status", "all", "--cursor", "next", "--limit", "2"}, scope...)
				}
				out.Reset()
				diagnostic.Reset()
				if code := Run(context.Background(), readCommand, getenv, &out, &diagnostic); code != 0 {
					t.Fatalf("read: %d %s", code, &diagnostic)
				}
				for _, field := range []string{`"acceptedPrincipalId":"guest"`, `"email":"guest@example.test"`, `"isExpired":false`, `"requestId":"trace"`} {
					if !strings.Contains(out.String(), field) {
						t.Fatalf("lost invitation field: %s", &out)
					}
				}
				if listing && !strings.Contains(out.String(), `"nextCursor":"more"`) {
					t.Fatal("lost cursor")
				}
				for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
					if code := Run(context.Background(), append(readCommand, override...), getenv, &out, &diagnostic); code == 0 {
						t.Fatal("read scope bypass")
					}
				}
			}
			before := calls
			if code := Run(context.Background(), append([]string{"invitations", "list", "--status", "bad"}, scope...), getenv, &out, &diagnostic); code != 2 || calls != before {
				t.Fatal("invalid status reached API")
			}

		})
	}
}

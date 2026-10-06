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

func TestPreferenceMutationRevisionsAndConflicts(t *testing.T) {
	for _, action := range []string{"update", "override", "remove-override"} {
		t.Run(action, func(t *testing.T) {
			calls := 0
			conflict := false
			body := `{"revision":9007199254740993,"defaults":{"enabled":false,"upcoming":false,"expired":true,"advanceDays":0},"timezone":"UTC","pushEnabled":false}`
			if action == "override" {
				body = `{"revision":9007199254740993,"settings":{"enabled":false,"upcoming":true,"expired":false,"advanceDays":0}}`
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				path := "/tenants/home/inventories/garage/notification-preferences"
				method := "PUT"
				if action != "update" {
					path += "/types/food"
				}
				if action == "remove-override" {
					method = "DELETE"
				}
				if r.URL.Path != path || r.Method != method || r.Header.Get("Authorization") != "Bearer owner" {
					t.Error("wrong mutation scope")
				}
				if action == "remove-override" {
					if r.URL.Query().Get("revision") != "9007199254740993" {
						t.Error("revision lost")
					}
				} else {
					b, _ := io.ReadAll(r.Body)
					if string(b) != body {
						t.Errorf("body changed: %s", b)
					}
				}
				if conflict {
					w.WriteHeader(409)
					return
				}
				io.WriteString(w, `{"data":{"revision":9007199254740994,"timezone":"UTC","pushEnabled":false,"defaults":{"enabled":false,"upcoming":true,"expired":true,"advanceDays":0},"overrides":null},"meta":{"requestId":"trace"}}`)
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

			input := filepath.Join(dir, "request.json")
			os.WriteFile(input, []byte(body), 0600)
			command := []string{"notification-preferences", action}
			if action != "update" {
				command = append(command, "food")
			}
			scope := []string{"--tenant", "home", "--inventory", "garage", "--json", "--no-input"}
			var out, diagnostic bytes.Buffer
			if code := Run(context.Background(), append(command, scope...), getenv, &out, &diagnostic); code != 2 || calls != 0 {
				t.Fatal("missing inputs reached API")
			}
			if action == "remove-override" {
				command = append(command, "--revision", "9007199254740993", "--yes")
			} else {
				command = append(command, "--input", input)
			}
			command = append(command, scope...)
			out.Reset()
			diagnostic.Reset()
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 || !strings.Contains(out.String(), `"revision":9007199254740994`) {
				t.Fatalf("mutation failed: %d %d %s %s", code, calls, &out, &diagnostic)
			}
			conflict = true
			out.Reset()
			diagnostic.Reset()
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || calls != 2 || !strings.Contains(diagnostic.String(), "conflict") {
				t.Fatalf("conflict retried: %d %d %s", code, calls, &diagnostic)
			}
		})
	}
}

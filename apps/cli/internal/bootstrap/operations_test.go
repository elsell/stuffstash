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

func TestUndoRedoConfirmationAndScope(t *testing.T) {
	for _, action := range []string{"undo", "redo"} {
		t.Run(action, func(t *testing.T) {
			calls := 0
			broken := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/tenants/home/inventories/garage/undoable-operations/op/"+action || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if r.Method != "POST" {
					t.Error("wrong method")
				}
				if broken {
					io.WriteString(w, "invalid")
					return
				}
				io.WriteString(w, `{"data":{"id":"asset","tenantId":"home","inventoryId":"garage","title":"Drill","kind":"item","description":"","lifecycleState":"active","createdAt":"today","updatedAt":"today","customFields":{"serial":9007199254740993},"tags":[],"expiration":null,"undoableOperationId":"next"},"meta":{"requestId":"trace"}}`)
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
			command := append([]string{"operations", action, "op"}, scope...)
			var out, diagnostic bytes.Buffer
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || calls != 0 {
				t.Fatal("unconfirmed operation reached API")
			}
			command = append(command, "--yes")
			out.Reset()
			diagnostic.Reset()
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 || !strings.Contains(out.String(), `"undoableOperationId":"next"`) || !strings.Contains(out.String(), `"serial":9007199254740993`) {
				t.Fatalf("operation: %d %d %s %s", code, calls, &out, &diagnostic)
			}
			for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
				if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
					t.Fatal("scope boundary bypassed")
				}
			}
			broken = true
			out.Reset()
			diagnostic.Reset()
			before := calls
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || calls != before+1 || !strings.Contains(diagnostic.String(), "before you try again") {
				t.Fatalf("unsafe recovery: %d %d %s", code, calls, &diagnostic)
			}
		})
	}
}

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

func TestDirectoryWriteCommandsRouteExactScopeAndBody(t *testing.T) {
	for _, test := range []struct {
		command      []string
		method, path string
	}{
		{[]string{"tenants", "create"}, "POST", "/tenants"},
		{[]string{"tenants", "update", "--tenant", "home"}, "PATCH", "/tenants/home"},
		{[]string{"inventories", "create", "--tenant", "home"}, "POST", "/tenants/home/inventories"},
		{[]string{"inventories", "update", "--tenant", "home", "--inventory", "garage"}, "PATCH", "/tenants/home/inventories/garage"},
	} {
		t.Run(test.path+test.method, func(t *testing.T) {
			count := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				body, _ := io.ReadAll(r.Body)
				if r.Method != test.method || r.URL.Path != test.path || string(body) != `{"name":"Garage"}` || r.Header.Get("Authorization") != "Bearer owner" {
					t.Errorf("wrong request %s %s %s", r.Method, r.URL.Path, body)
				}
				w.Header().Set("Content-Type", "application/json")
				io.WriteString(w, `{"data":{"id":"garage","tenantId":"home","name":"Garage","lifecycleState":"active","access":{"relationship":"owner","permissions":["write"]}},"meta":{}}`)
			}))
			defer server.Close()
			dir := t.TempDir()
			os.Chmod(dir, 0700)
			store := credentials.File{Path: filepath.Join(dir, "session.json")}
			if err := store.Save(context.Background(), ports.Session{Server: server.URL, IDToken: "owner", Issuer: "https://id.example", Subject: "owner", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
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
			var out, diagnostic bytes.Buffer
			args := append(append([]string{}, test.command...), "--name", "Garage", "--json", "--no-input")
			code := Run(context.Background(), args, getenv, &out, &diagnostic)
			if code != 0 || count != 1 || !strings.Contains(out.String(), `"name":"Garage"`) || !strings.Contains(diagnostic.String(), "Server:") {
				t.Fatalf("write failed %d count=%d %s %s", code, count, &out, &diagnostic)
			}
		})
	}
}

func TestUnsupportedInputRejectedBeforeConnectorDispatch(t *testing.T) {
	var out, diagnostic bytes.Buffer
	code := Run(context.Background(), []string{"connectors", "print", "register", "--name", "Home", "--input", "missing.json"}, func(string) string { return "" }, &out, &diagnostic)
	if code != 2 || !strings.Contains(diagnostic.String(), "does not accept --input") || out.Len() != 0 {
		t.Fatalf("dispatch before validation: %d %s %s", code, &out, &diagnostic)
	}
}

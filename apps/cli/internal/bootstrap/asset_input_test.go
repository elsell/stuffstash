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

func TestAssetJSONCommandsRouteCompleteInput(t *testing.T) {
	for _, test := range []struct {
		command      []string
		method, path string
		print        bool
	}{
		{[]string{"assets", "create"}, "POST", "/tenants/home/inventories/garage/assets", false},
		{[]string{"assets", "create"}, "POST", "/tenants/home/inventories/garage/assets", true},
		{[]string{"assets", "update", "asset"}, "PATCH", "/tenants/home/inventories/garage/assets/asset", false},
	} {
		t.Run(test.path+test.method, func(t *testing.T) {
			count := 0
			requestBody := `{"title":"Garage","kind":"container","customFields":{"serial":9007199254740993}}`
			if test.print {
				requestBody = `{"title":"Garage","kind":"container","printLabel":{"printerId":"printer"}}`
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				body, _ := io.ReadAll(r.Body)
				if r.Method != test.method || r.URL.Path != test.path || string(body) != requestBody || r.Header.Get("Authorization") != "Bearer owner" {
					t.Errorf("wrong request %s %s %s", r.Method, r.URL.Path, body)
				}
				if (r.Header.Get("Idempotency-Key") != "") != test.print {
					t.Error("wrong retry key behavior")
				}
				if count == 2 {
					io.WriteString(w, "invalid JSON")
					return
				}
				w.Header().Set("Content-Type", "application/json")
				io.WriteString(w, `{"data":{"id":"asset","title":"Garage","tags":[],"expiration":null},"meta":{}}`)
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
			inputPath := filepath.Join(dir, "request.json")
			if err := os.WriteFile(inputPath, []byte(requestBody), 0600); err != nil {
				t.Fatal(err)
			}
			var out, diagnostic bytes.Buffer
			args := append(append([]string{}, test.command...), "--input", inputPath, "--tenant", "home", "--inventory", "garage", "--json", "--no-input")
			code := Run(context.Background(), args, getenv, &out, &diagnostic)
			if code != 0 || count != 1 || !strings.Contains(out.String(), `"title":"Garage"`) || !strings.Contains(diagnostic.String(), "Server:") {
				t.Fatalf("write failed %d count=%d %s %s", code, count, &out, &diagnostic)
			}
			if test.method == "POST" {
				out.Reset()
				diagnostic.Reset()
				code = Run(context.Background(), args, getenv, &out, &diagnostic)
				want := "Run assets list"
				if test.print {
					want = "same request key"
				}
				if code == 0 || count != 2 || !strings.Contains(diagnostic.String(), want) {
					t.Fatalf("unsafe recovery: %d count=%d %s", code, count, &diagnostic)
				}
				if test.print && !strings.Contains(diagnostic.String(), "Print request key:") {
					t.Fatal("retry key not visible")
				}
			}
		})
	}
}

func TestMoveRetryKeyRejectedBeforeLogin(t *testing.T) {
	var out, diagnostic bytes.Buffer
	code := Run(context.Background(), []string{"assets", "move", "asset", "--parent", "shelf", "--idempotency-key", "retry", "--server", "https://stash.example"}, func(string) string { return "" }, &out, &diagnostic)
	if code != 2 || !strings.Contains(diagnostic.String(), "Remove --idempotency-key") || out.Len() != 0 {
		t.Fatalf("unsupported retry: %d %s %s", code, &out, &diagnostic)
	}
}

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

func TestCheckoutCommandsPreserveNotesAndDoNotRetry(t *testing.T) {
	for _, test := range []struct {
		command      []string
		method, path string
	}{
		{[]string{"assets", "checkout", "asset"}, "POST", "/tenants/home/inventories/garage/assets/asset/checkout"},
		{[]string{"assets", "return", "asset"}, "POST", "/tenants/home/inventories/garage/assets/asset/return"},
		{[]string{"assets", "return-details", "asset", "checkout"}, "PATCH", "/tenants/home/inventories/garage/assets/asset/checkouts/checkout/return-details"},
	} {
		t.Run(test.path+test.method, func(t *testing.T) {
			count := 0
			requestBody := `{"details":""}`
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				body, _ := io.ReadAll(r.Body)
				if r.Method != test.method || r.URL.Path != test.path || string(body) != requestBody || r.Header.Get("Authorization") != "Bearer owner" {
					t.Errorf("wrong request %s %s %s", r.Method, r.URL.Path, body)
				}
				if r.Header.Get("Idempotency-Key") != "" {
					t.Error("wrong retry key behavior")
				}
				if count == 2 {
					io.WriteString(w, "invalid JSON")
					return
				}
				w.Header().Set("Content-Type", "application/json")
				io.WriteString(w, `{"data":{"id":"checkout","assetId":"asset","state":"returned","returnDetails":""},"meta":{}}`)
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
			invalidArgs := append(append([]string{}, test.command...), "--name", "Alice", "--tenant", "home", "--inventory", "garage", "--json", "--no-input")
			if code := Run(context.Background(), invalidArgs, getenv, &out, &diagnostic); code != 2 || count != 0 {
				t.Fatalf("ignored field sent a request: %d %d %s", code, count, &diagnostic)
			}
			out.Reset()
			diagnostic.Reset()
			code := Run(context.Background(), args, getenv, &out, &diagnostic)
			if code != 0 || count != 1 || !strings.Contains(out.String(), `"returnDetails":""`) || !strings.Contains(diagnostic.String(), "Server:") {
				t.Fatalf("write failed %d count=%d %s %s", code, count, &out, &diagnostic)
			}
			{
				out.Reset()
				diagnostic.Reset()
				code = Run(context.Background(), args, getenv, &out, &diagnostic)
				want := "Run assets checkouts"
				if code == 0 || count != 2 || !strings.Contains(diagnostic.String(), want) {
					t.Fatalf("unsafe recovery: %d count=%d %s", code, count, &diagnostic)
				}
			}
		})
	}
}

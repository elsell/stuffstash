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

func TestAttachmentUploadCompletionAndUncertainRecovery(t *testing.T) {
	calls := 0
	broken := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/tenants/home/inventories/garage/assets/asset/attachments/direct-uploads/secret-upload-token/complete" || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer owner" {
			t.Errorf("wrong completion request: %s %s", r.Method, r.URL)
		}
		w.WriteHeader(201)
		if broken {
			io.WriteString(w, "broken")
			return
		}
		io.WriteString(w, `{"data":{"id":"photo","tenantId":"home","inventoryId":"garage","assetId":"asset","fileName":"photo.jpg","contentType":"image/jpeg","sizeBytes":15,"sha256":"digest","createdAt":"today","lifecycleState":"active"},"meta":{"requestId":"trace"}}`)
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
	var out, diagnostic bytes.Buffer

	command := append([]string{"attachments", "complete-upload", "asset", "secret-upload-token"}, scope...)
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 || !strings.Contains(out.String(), `"sha256":"digest"`) {
		t.Fatalf("completion failed: %d %d %s %s", code, calls, &out, &diagnostic)
	}
	if strings.Contains(diagnostic.String(), "secret-upload-token") {
		t.Fatal("upload token leaked")
	}
	broken = true
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || calls != 2 || !strings.Contains(diagnostic.String(), "attachments list") || strings.Contains(diagnostic.String(), "secret-upload-token") {
		t.Fatalf("unsafe recovery: %d %d %s", code, calls, &diagnostic)
	}
}

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

func TestAttachmentCommandsAndConfirmation(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/tenants/home/inventories/garage/assets/asset/attachments/photo" || r.Header.Get("Authorization") != "Bearer owner" {
			t.Errorf("wrong request: %s", r.URL)
		}
		if r.Method == "DELETE" {
			w.WriteHeader(204)
			return
		}
		if r.Method != "GET" {
			t.Error("wrong method")
		}
		io.WriteString(w, `{"data":{"id":"photo","tenantId":"home","inventoryId":"garage","assetId":"asset","fileName":"photo.jpg","contentType":"image/jpeg","sizeBytes":15,"sha256":"digest","createdAt":"today","lifecycleState":"active"},"meta":{}}`)
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

	for _, action := range []string{"archive", "delete"} {
		if code := Run(context.Background(), append([]string{"attachments", action, "asset", "photo"}, scope...), getenv, &out, &diagnostic); code != 2 || calls != 0 {
			t.Fatalf("unapproved mutation: %d %d %s", code, calls, &diagnostic)
		}
	}
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), append([]string{"attachments", "show", "asset", "photo"}, scope...), getenv, &out, &diagnostic); code != 0 || calls != 1 || !strings.Contains(out.String(), `"sha256":"digest"`) {
		t.Fatalf("show: %d %d %s %s", code, calls, &out, &diagnostic)
	}
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), append([]string{"attachments", "delete", "asset", "photo", "--yes"}, scope...), getenv, &out, &diagnostic); code != 0 || calls != 2 || !strings.Contains(out.String(), `"attachmentId":"photo"`) {
		t.Fatalf("delete: %d %d %s %s", code, calls, &out, &diagnostic)
	}
}

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

func TestTagCommandUpdateAndUnconfirmedDeletion(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		body, _ := io.ReadAll(r.Body)
		if r.Method != "PATCH" || r.URL.Path != "/tenants/home/inventories/garage/tags/tag" || r.Header.Get("Authorization") != "Bearer owner" || string(body) != `{"color":""}` {
			t.Errorf("wrong tag request: %s %s %s", r.Method, r.URL.Path, body)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"data":{"id":"tag","tenantId":"home","inventoryId":"garage","key":"tools","displayName":"Tools","color":"","lifecycleState":"active","createdAt":"created","updatedAt":"updated"},"meta":{}}`)
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
	if code := Run(context.Background(), append([]string{"tags", "delete", "tag"}, scope...), getenv, &out, &diagnostic); code != 2 || calls != 0 {
		t.Fatalf("unapproved delete: %d %d %s", code, calls, &diagnostic)
	}
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), append([]string{"tags", "update", "tag", "--tag-color", ""}, scope...), getenv, &out, &diagnostic); code != 0 || calls != 1 || !strings.Contains(out.String(), `"color":""`) {
		t.Fatalf("color update: %d %d %s %s", code, calls, &out, &diagnostic)
	}
}

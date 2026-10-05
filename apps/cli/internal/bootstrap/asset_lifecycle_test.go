package bootstrap

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAssetDeleteRequiresApprovalAndPreservesInventory(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "DELETE" || r.URL.Path != "/tenants/home/inventories/garage/assets/asset" || r.Header.Get("Authorization") != "Bearer owner" {
			t.Errorf("wrong delete: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	store := credentials.File{Path: filepath.Join(dir, "session.json")}
	session := ports.Session{Server: server.URL, Issuer: "https://id.example", Subject: "owner", IDToken: "owner", ExpiresAt: time.Now().Add(time.Hour)}
	if err := store.Save(context.Background(), session); err != nil {
		t.Fatal(err)
	}
	config := contextfile.Store{Path: filepath.Join(dir, "config", "contexts.json")}
	manager := contexts.Manager{Store: config}
	if err := manager.Remember(context.Background(), contexts.Entry{Name: "home", Server: server.URL, Principal: contexts.Principal(session), Tenant: "home", Inventory: "garage"}); err != nil {
		t.Fatal(err)
	}
	getenv := func(k string) string {
		switch k {
		case "STUFF_STASH_CLI_CONFIG_FILE":
			return config.Path
		case "STUFF_STASH_CLI_CREDENTIAL_FILE":
			return store.Path
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		}
		return ""
	}
	args := []string{"assets", "delete", "asset", "--json", "--no-input"}
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), args, getenv, &out, &diagnostic); code != 2 || calls != 0 {
		t.Fatalf("unapproved delete: code=%d calls=%d %s", code, calls, &diagnostic)
	}
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), append(args, "--yes"), getenv, &out, &diagnostic); code != 0 || calls != 1 || !strings.Contains(out.String(), `"status":"deleted"`) {
		t.Fatalf("confirmed delete: code=%d calls=%d %s %s", code, calls, &out, &diagnostic)
	}
	current, err := manager.Current(context.Background())
	if err != nil || current.Tenant != "home" || current.Inventory != "garage" {
		t.Fatalf("stale scope: %+v %v", current, err)
	}
}

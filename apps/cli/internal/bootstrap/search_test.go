package bootstrap

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestSearchUsesSavedScopeAndPreservesDefaultsAcrossOverrides(t *testing.T) {
	ctx := context.Background()
	calls := 0
	inventory := "garage"
	denied := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		q := r.URL.Query()
		if r.Method != "GET" || r.URL.Path != "/tenants/home/search/assets" || r.Header.Get("Authorization") != "Bearer owner" || q.Get("inventoryId") != inventory || q.Get("q") != "Cordless drill" || q.Get("mode") != "exact" || q.Get("customAssetTypeId") != "tools" || q.Get("lifecycleState") != "all" || q.Get("checkoutState") != "available" || q.Get("limit") != "3" || q.Get("cursor") != "next" || q.Get("tagIds") != "one,two" {
			t.Errorf("wrong search request: %s %s", r.Method, r.URL)
		}
		w.Header().Set("Content-Type", "application/json")
		if denied {
			w.WriteHeader(403)
			io.WriteString(w, `{"message":"private server internals"}`)
			return
		}
		io.WriteString(w, `{"data":[],"meta":{"pagination":{"limit":3,"hasMore":true,"nextCursor":"after"}}}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	credential := credentials.File{Path: filepath.Join(dir, "session.json")}
	session := ports.Session{Server: server.URL, Issuer: "https://id.example", Subject: "owner", IDToken: "owner", ExpiresAt: time.Now().Add(time.Hour)}
	if err := credential.Save(ctx, session); err != nil {
		t.Fatal(err)
	}
	config := contextfile.Store{Path: filepath.Join(dir, "config", "contexts.json")}
	if err := (contexts.Manager{Store: config}).Remember(ctx, contexts.Entry{Name: "home", Server: server.URL, Principal: contexts.Principal(session), Tenant: "home", Inventory: "garage"}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(config.Path)
	if err != nil {
		t.Fatal(err)
	}
	env := func(key string) string {
		switch key {
		case "STUFF_STASH_CLI_SERVER":
			return server.URL
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		case "STUFF_STASH_CLI_CREDENTIAL_FILE":
			return credential.Path
		case "STUFF_STASH_CLI_CONFIG_FILE":
			return config.Path
		}
		return ""
	}
	args := []string{"assets", "search", "--query", "Cordless drill", "--mode", "exact", "--type-id", "tools", "--tag-id", "one", "--tag-id", "two", "--lifecycle", "all", "--checkout-state", "available", "--limit", "3", "--cursor", "next", "--json", "--no-input"}
	for _, test := range []struct {
		inventory string
		flags     []string
	}{{"garage", nil}, {"kitchen", []string{"--inventory", "kitchen"}}, {"", []string{"--all-inventories"}}} {
		inventory = test.inventory
		var out, diagnostic bytes.Buffer
		if code := Run(ctx, append(append([]string{}, args...), test.flags...), env, &out, &diagnostic); code != 0 || !strings.Contains(out.String(), `"hasMore":true`) {
			t.Fatalf("search: %d %s %s", code, &out, &diagnostic)
		}
	}
	inventory = "garage"
	denied = true
	var out, diagnostic bytes.Buffer
	if code := Run(ctx, args, env, &out, &diagnostic); code != 1 || out.Len() != 0 || strings.Contains(diagnostic.String(), "private server internals") {
		t.Fatalf("denial: %d %s %s", code, &out, &diagnostic)
	}
	after, err := os.ReadFile(config.Path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("search changed saved defaults: %v", err)
	}
	previous := calls
	out.Reset()
	diagnostic.Reset()
	if code := Run(ctx, append(args, "--kind", "item"), env, &out, &diagnostic); code != 2 || calls != previous {
		t.Fatalf("unsupported filter reached server: %d %s", code, &diagnostic)
	}
}

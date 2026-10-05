package bootstrap

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalContextListWithoutServerOrLogin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "contexts.json")
	getenv := func(key string) string {
		if key == "STUFF_STASH_CLI_CONFIG_FILE" {
			return path
		}
		return ""
	}
	var out, diagnostics bytes.Buffer
	code := Run(context.Background(), []string{"context", "list", "--json", "--no-input"}, getenv, &out, &diagnostics)
	if code != 0 || out.String() != "[]\n" || diagnostics.Len() != 0 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
	}
}

func TestCanceledContextCommandExits130(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out, diagnostics bytes.Buffer
	getenv := func(key string) string {
		if key == "STUFF_STASH_CLI_CONFIG_FILE" {
			return filepath.Join(t.TempDir(), "contexts.json")
		}
		return ""
	}
	if code := Run(ctx, []string{"context", "list", "--json"}, getenv, &out, &diagnostics); code != 130 {
		t.Fatalf("cancellation exit=%d: %s", code, diagnostics.String())
	}
}

func TestRememberedScopeIsBoundToAuthenticatedAccount(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Error("missing authentication")
		}
		if r.URL.Path != "/tenants/home/inventories" {
			t.Errorf("unexpected scope: %s", r.URL.Path)
			w.WriteHeader(403)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"data":[],"meta":{"pagination":{"hasMore":false}}}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	session := ports.Session{Server: server.URL, Issuer: "https://identity.example", Subject: "alice", IDToken: "user-token", ExpiresAt: time.Now().Add(time.Hour)}
	credentialsStore := credentials.File{Path: filepath.Join(dir, "session.json")}
	config := contextfile.Store{Path: filepath.Join(dir, "config", "contexts.json")}
	manager := contexts.Manager{Store: config}
	if err := manager.Remember(ctx, contexts.Entry{Name: "home", Server: server.URL, Principal: contexts.Principal(session), Tenant: "home", Inventory: "garage"}); err != nil {
		t.Fatal(err)
	}
	getenv := func(key string) string {
		switch key {
		case "STUFF_STASH_CLI_CONFIG_FILE":
			return config.Path
		case "STUFF_STASH_CLI_CREDENTIAL_FILE":
			return credentialsStore.Path
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		}
		return ""
	}
	for _, test := range []struct {
		subject string
		want    int
	}{{"alice", 0}, {"bob", 2}} {
		session.Subject = test.subject
		if err := credentialsStore.Save(ctx, session); err != nil {
			t.Fatal(err)
		}
		var out, diagnostics bytes.Buffer
		code := Run(ctx, []string{"inventories", "list", "--json", "--no-input"}, getenv, &out, &diagnostics)
		if code != test.want {
			t.Fatalf("account=%s exit=%d stdout=%s stderr=%s", test.subject, code, out.String(), diagnostics.String())
		}
	}
	var out, diagnostics bytes.Buffer
	if code := Run(ctx, []string{"logout", "--json"}, getenv, &out, &diagnostics); code != 0 {
		t.Fatalf("logout=%d %s", code, diagnostics.String())
	}
	saved, err := manager.Current(ctx)
	if err != nil || saved.Principal != "" || saved.Tenant != "" || saved.Inventory != "" {
		t.Fatalf("logout retained saved account scope: %+v %v", saved, err)
	}
}

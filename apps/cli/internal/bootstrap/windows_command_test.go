//go:build windows

package bootstrap

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestWindowsKeyringCommandBoundary(t *testing.T) {
	var identity [16]byte
	if _, err := rand.Read(identity[:]); err != nil {
		t.Fatal("Cannot create test identity")
	}
	prefix := "/cli-test-" + hex.EncodeToString(identity[:])
	const secret = "synthetic-windows-command-secret"
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "Bearer "+secret {
			t.Error("Stored credential not used")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch strings.TrimPrefix(r.URL.Path, prefix) {
		case "/tenants/house/inventories/inventory/assets":
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}, "meta": map[string]any{}})
		case "/tenants/house/inventories/forbidden/assets":
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{"detail": "Synthetic denied scope", "status": 403})
		default:
			t.Error("Unexpected request outside selected scope")
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	baseURL := server.URL + prefix
	ctx := context.Background()
	store := credentials.Keyring{}
	t.Cleanup(func() {
		if err := store.Delete(ctx, baseURL); err != nil {
			t.Error("Cannot clean up synthetic credential")
		}
	})
	session := ports.Session{Server: baseURL, Issuer: "https://issuer.invalid", Subject: "synthetic", ClientID: "test", IDToken: secret, ExpiresAt: time.Now().Add(time.Hour)}
	if err := store.Save(ctx, session); err != nil {
		t.Fatal("Cannot prepare Windows credential store")
	}
	contextPath := filepath.Join(t.TempDir(), "missing-context-dir", "contexts.json")
	getenv := func(k string) string {
		if k == "STUFF_STASH_CLI_CONFIG_FILE" {
			return contextPath
		}
		if k == "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP" {
			return "true"
		}
		return ""
	}
	run := func(inventory string) (int, string, string) {
		var out, diagnostics bytes.Buffer
		code := Run(ctx, []string{"assets", "list", "--server", baseURL, "--tenant", "house", "--inventory", inventory, "--json", "--no-input"}, getenv, &out, &diagnostics)
		if strings.Contains(out.String()+diagnostics.String(), secret) {
			t.Fatal("Credential exposed in output")
		}
		return code, out.String(), diagnostics.String()
	}
	code, out, _ := run("inventory")
	if code != 0 || !json.Valid([]byte(out)) || calls.Load() != 1 {
		t.Fatal("Authenticated Windows command did not return JSON once")
	}
	code, _, _ = run("forbidden")
	if code == 0 || calls.Load() != 2 {
		t.Fatal("Forbidden scope succeeded or retried")
	}
	if err := store.Delete(ctx, baseURL); err != nil {
		t.Fatal("Cannot remove synthetic credential")
	}
	code, _, _ = run("inventory")
	if code == 0 || calls.Load() != 2 {
		t.Fatal("Missing credential did not fail before request")
	}
}

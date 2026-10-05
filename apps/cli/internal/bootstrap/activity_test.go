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

func TestActivityCommandRespectsScopeAndPreservesTimeline(t *testing.T) {
	calls := 0
	path := "/tenants/home/inventories/garage/assets/asset/activity"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != path || r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		if r.Method != "GET" || r.URL.Query().Get("limit") != "2" || r.URL.Query().Get("view") != "all" || r.URL.Query().Get("cursor") != "next" {
			t.Error("wrong query")
		}
		io.WriteString(w, `{"data":[{"id":"event","action":"asset.updated","category":"change","changes":[{"field":"title","previousValue":"Old","currentValue":"New"}],"occurredAt":"2026-10-05T12:00:00Z","principalId":"owner","principal":{"id":"owner","email":"owner@example.test"},"requestId":"original","source":"api","technicalMetadata":{"changed":"title"},"undo":{"operationId":"undo-id","status":"available"}}],"meta":{"pagination":{"limit":2,"hasMore":true,"nextCursor":"more"}}}`)
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

	command := []string{"assets", "activity", "asset", "--view", "all", "--cursor", "next", "--tenant", "home", "--inventory", "garage", "--limit", "2", "--json", "--no-input"}
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 {
		t.Fatalf("activity: %d %d %s", code, calls, &diagnostic)
	}
	for _, field := range []string{`"email":"owner@example.test"`, `"requestId":"original"`, `"changed":"title"`, `"previousValue":"Old"`, `"operationId":"undo-id"`, `"nextCursor":"more"`} {
		if !strings.Contains(out.String(), field) {
			t.Fatalf("activity field lost: %s", &out)
		}
	}
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), append(command, "--tenant", "other"), getenv, &out, &diagnostic); code == 0 {
		t.Fatal("wrong household succeeded")
	}
	if code := Run(context.Background(), append(command, "--inventory", "other"), getenv, &out, &diagnostic); code == 0 {
		t.Fatal("wrong inventory succeeded")
	}
	before := calls
	if code := Run(context.Background(), append(command, "--view", "invalid"), getenv, &out, &diagnostic); code != 2 || calls != before {
		t.Fatal("invalid view accepted")
	}
}

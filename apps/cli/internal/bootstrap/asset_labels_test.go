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

func TestAssetLabelScopeAndAssignment(t *testing.T) {
	method := "GET"
	calls := 0
	failure := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer owner" || r.URL.Path != "/tenants/home/inventories/garage/assets/asset/label" {
			w.WriteHeader(403)
			return
		}
		if r.Method != method || r.URL.RawQuery != "" {
			t.Error("wrong request")
		}
		if failure {
			w.WriteHeader(503)
			io.WriteString(w, "private-detail")
			return
		}
		io.WriteString(w, `{"$schema":"label-schema","data":{"instanceId":"instance","labelId":"label","url":"https://example.test/label","tenantId":"home","inventoryId":"garage","assetId":"asset","lifecycleState":"active"},"meta":{"requestId":"trace"}}`)
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

	for _, action := range []string{"show", "assign"} {
		command := []string{"labels", action, "asset", "--tenant", "home", "--inventory", "garage", "--json", "--no-input"}
		var out, diagnostic bytes.Buffer
		if action == "assign" {
			method = "POST"
			before := calls
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || calls != before {
				t.Fatal("unconfirmed assignment")
			}
			command = append(command, "--yes")
		}
		beforeOptions := calls
		if code := Run(context.Background(), append(command, "--output", "label.png"), getenv, &out, &diagnostic); code != 2 || calls != beforeOptions {
			t.Fatal("render option reached identity API")
		}
		out.Reset()
		diagnostic.Reset()
		if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 {
			t.Fatalf("label: %d %s", code, &diagnostic)
		}
		for _, field := range []string{`"instanceId":"instance"`, `"labelId":"label"`, `"url":"https://example.test/label"`, `"tenantId":"home"`, `"inventoryId":"garage"`, `"assetId":"asset"`, `"lifecycleState":"active"`, `"requestId":"trace"`, `"$schema":"label-schema"`} {
			if !strings.Contains(out.String(), field) {
				t.Fatalf("missing %s: %s", field, &out)
			}
		}
		for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
			if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
				t.Fatal("scope bypass")
			}
		}
		before := calls
		failure = true
		out.Reset()
		diagnostic.Reset()
		if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || calls != before+1 || strings.Contains(diagnostic.String(), "private-detail") {
			t.Fatal("unsafe failure or retry")
		}
		failure = false
	}
}

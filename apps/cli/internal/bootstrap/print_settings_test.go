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

func TestPrintSettingsCompleteRead(t *testing.T) {
	printer := "null"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer owner" || r.URL.Path != "/tenants/home/inventories/garage/print-settings" {
			w.WriteHeader(403)
			return
		}
		if r.Method != "GET" || r.URL.RawQuery != "" {
			t.Error("wrong request")
		}
		io.WriteString(w, `{"$schema":"response-schema","data":{"$schema":"settings-schema","defaultPrinterId":`+printer+`,"printOnCreateDefault":false,"revision":9007199254740993,"template":{"id":"template","version":2,"options":{"showReference":true}}},"meta":{"requestId":"trace"}}`)
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

	command := []string{"print-settings", "show", "--tenant", "home", "--inventory", "garage", "--json", "--no-input"}
	for _, value := range []string{"null", `"printer"`} {
		printer = value
		var out, diagnostic bytes.Buffer
		if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 {
			t.Fatalf("read: %d %s", code, &diagnostic)
		}
		for _, field := range []string{`"defaultPrinterId":` + value, `"printOnCreateDefault":false`, `"revision":9007199254740993`, `"version":2`, `"showReference":true`, `"requestId":"trace"`, `"$schema":"settings-schema"`, `"$schema":"response-schema"`} {
			if !strings.Contains(out.String(), field) {
				t.Fatalf("missing %s: %s", field, &out)
			}
		}
		for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
			if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
				t.Fatal("scope bypass")
			}
		}
	}
}

package bootstrap

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicServerCommandsNeedNoCredentials(t *testing.T) {
	calls := 0
	fail := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "" || r.Method != "GET" {
			t.Error("public discovery sent credentials or wrong method")
		}
		if fail {
			w.WriteHeader(503)
			io.WriteString(w, "private-server-error")
			return
		}
		switch r.URL.Path {
		case "/instance":
			io.WriteString(w, `{"$schema":"instance-schema","data":{"instanceId":"home-server","protocolVersion":9007199254740993},"meta":{"requestId":"trace"}}`)
		case "/auth/cli/config":
			io.WriteString(w, `{"$schema":"auth-schema","data":{"issuer":"https://id.example","clientId":"cli","scopes":["openid","profile"],"loginMethods":["loopback","device_code"],"loopbackRedirect":{"host":"127.0.0.1","pathPrefix":"/callback/","ephemeralPort":true}},"meta":{"requestId":"trace"}}`)
		default:
			t.Error("wrong path")
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	dir := t.TempDir()
	getenv := func(k string) string {
		switch k {
		case "STUFF_STASH_CLI_SERVER":
			return server.URL
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		case "STUFF_STASH_CLI_CREDENTIAL_FILE":
			return filepath.Join(dir, "absent.json")
		case "STUFF_STASH_CLI_CONFIG_FILE":
			return filepath.Join(dir, "config", "contexts.json")
		}
		return ""
	}
	for _, action := range []string{"show", "auth-config"} {
		var out, diagnostic bytes.Buffer
		if code := Run(context.Background(), []string{"server", action, "--json", "--no-input"}, getenv, &out, &diagnostic); code != 0 {
			t.Fatalf("server: %d %s", code, &diagnostic)
		}
		expected := []string{`"protocolVersion":9007199254740993`, `"instanceId":"home-server"`}
		if action == "auth-config" {
			expected = []string{`"issuer":"https://id.example"`, `"scopes":["openid","profile"]`, `"loginMethods":["loopback","device_code"]`, `"ephemeralPort":true`, `"pathPrefix":"/callback/"`}
		}
		for _, field := range append(expected, `"requestId":"trace"`, `"$schema":`) {
			if !strings.Contains(out.String(), field) {
				t.Fatalf("lost field: %s", &out)
			}
		}
	}
	if calls != 2 {
		t.Fatal("extra discovery requests")
	}
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), []string{"server", "show", "--print-label"}, getenv, &out, &diagnostic); code != 2 || calls != 2 {
		t.Fatal("invalid option reached server")
	}
	fail = true
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), []string{"server", "show", "--json", "--no-input"}, getenv, &out, &diagnostic); code == 0 || strings.Contains(out.String()+diagnostic.String(), "private-server-error") {
		t.Fatal("unsafe server error")
	}

}

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

func TestNotificationDeviceRemovalRequiresRevisionAndConfirmation(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		if r.Method == "GET" {
			if r.URL.Path != "/tenants/home/inventories/garage/notification-devices/by-installation/install" {
				t.Error("wrong lookup path")
			}
		} else {
			if r.Method != "DELETE" || r.URL.Path != "/tenants/home/inventories/garage/notification-devices/device" || r.URL.Query().Get("revision") != "9007199254740993" {
				t.Error("wrong removal request")
			}
		}
		io.WriteString(w, `{"data":{"id":"device","installationId":"install","transport":"apns","revision":9007199254740993,"active":false},"meta":{"requestId":"trace"}}`)
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
	for _, args := range [][]string{{"notification-devices", "remove", "device", "--yes"}, {"notification-devices", "remove", "device", "--revision", "9007199254740993"}} {
		if code := Run(context.Background(), append(args, scope...), getenv, &out, &diagnostic); code != 2 || calls != 0 {
			t.Fatalf("unguarded removal: %d %d %s", code, calls, &diagnostic)
		}
	}
	for _, args := range [][]string{{"notification-devices", "show", "install"}, {"notification-devices", "remove", "device", "--revision", "9007199254740993", "--yes"}} {
		out.Reset()
		diagnostic.Reset()
		if code := Run(context.Background(), append(args, scope...), getenv, &out, &diagnostic); code != 0 || !strings.Contains(out.String(), `"revision":9007199254740993`) || !strings.Contains(out.String(), `"active":false`) {
			t.Fatalf("device command: %d %s %s", code, &out, &diagnostic)
		}
	}
	if calls != 2 {
		t.Fatalf("unexpected retries: %d", calls)
	}
}

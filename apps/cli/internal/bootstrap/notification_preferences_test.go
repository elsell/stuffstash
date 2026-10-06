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

func TestNotificationPreferenceShowAndInitialize(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		path := "/tenants/home/inventories/garage/notification-preferences"
		if r.Method == "POST" {
			path += "/initialize"
			b, _ := io.ReadAll(r.Body)
			if string(b) != `{"timezone":"America/New_York"}` {
				t.Errorf("wrong timezone: %s", b)
			}
		} else if r.Method != "GET" {
			t.Error("wrong method")
		}
		if r.URL.Path != path {
			t.Error("wrong scope")
		}
		io.WriteString(w, `{"data":{"revision":9007199254740993,"timezone":"America/New_York","pushEnabled":false,"defaults":{"enabled":false,"upcoming":true,"expired":true,"advanceDays":3},"overrides":[{"customAssetTypeId":"food","settings":{"enabled":true,"upcoming":true,"expired":false,"advanceDays":1}}]},"meta":{"requestId":"trace"}}`)
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
	if code := Run(context.Background(), append([]string{"notification-preferences", "initialize"}, scope...), getenv, &out, &diagnostic); code != 2 || calls != 0 {
		t.Fatal("missing timezone sent a request")
	}
	for _, args := range [][]string{{"notification-preferences", "show"}, {"notification-preferences", "initialize", "--timezone", "America/New_York"}} {
		out.Reset()
		diagnostic.Reset()
		if code := Run(context.Background(), append(args, scope...), getenv, &out, &diagnostic); code != 0 || !strings.Contains(out.String(), `"revision":9007199254740993`) || !strings.Contains(out.String(), `"customAssetTypeId":"food"`) || !strings.Contains(out.String(), `"pushEnabled":false`) {
			t.Fatalf("preference command: %d %s %s", code, &out, &diagnostic)
		}
	}
	if calls != 2 {
		t.Fatalf("unexpected requests: %d", calls)
	}
}

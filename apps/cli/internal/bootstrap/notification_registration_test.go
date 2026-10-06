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

func TestNotificationRegistrationKeepsTokenPrivate(t *testing.T) {
	calls := 0
	broken := false
	body := `{"installationId":"install","transport":"apns","token":"secret-device-token","revision":0}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/tenants/home/inventories/garage/notification-devices" || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer owner" {
			t.Error("wrong registration request")
		}
		b, _ := io.ReadAll(r.Body)
		if string(b) != body {
			t.Error("registration body changed")
		}
		if broken {
			io.WriteString(w, "broken secret-device-token")
			return
		}
		io.WriteString(w, `{"data":{"id":"device","installationId":"install","transport":"apns","revision":1,"active":true},"meta":{"requestId":"trace"}}`)
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

	input := filepath.Join(dir, "request.json")
	os.WriteFile(input, []byte(body), 0600)
	scope := []string{"--tenant", "home", "--inventory", "garage", "--json", "--no-input"}
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), append([]string{"notification-devices", "register"}, scope...), getenv, &out, &diagnostic); code != 2 || calls != 0 {
		t.Fatal("missing registration input reached API")
	}
	command := append([]string{"notification-devices", "register", "--input", input}, scope...)
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 || !strings.Contains(out.String(), `"active":true`) {
		t.Fatalf("register: %d %d %s %s", code, calls, &out, &diagnostic)
	}
	if strings.Contains(out.String()+diagnostic.String(), "secret-device-token") {
		t.Fatal("token leaked")
	}
	broken = true
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || calls != 2 || !strings.Contains(diagnostic.String(), "notification-devices show") || strings.Contains(out.String()+diagnostic.String(), "secret-device-token") {
		t.Fatalf("unsafe registration recovery: %d %d %s", code, calls, &diagnostic)
	}
}

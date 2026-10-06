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

func TestProviderLifecycleConfirmationAndScope(t *testing.T) {
	action := ""
	calls := 0
	status := 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/tenants/home/provider-profiles/profile/"+action || r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		if r.Method != "POST" || r.URL.RawQuery != "" {
			t.Error("wrong request")
		}
		if status != 200 {
			w.WriteHeader(status)
			io.WriteString(w, "private-provider-secret")
			return
		}
		data := `{"id":"profile","tenantId":"home","lifecycleState":"disabled","credentialStatus":"configured","runtimeOptions":{"limit":9007199254740993}}`
		if action == "test" {
			data = `{"providerProfileId":"profile","status":"failed","message":"Check provider settings","providerKind":"compatible","capability":"inference","testedAt":"today"}`
		}
		io.WriteString(w, `{"data":`+data+`,"meta":{"requestId":"trace"}}`)
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

	for _, actionName := range []string{"enable", "disable", "archive", "test"} {
		action = actionName
		command := []string{"provider-profiles", action, "profile", "--tenant", "home", "--json", "--no-input"}
		var out, diagnostic bytes.Buffer
		before := calls
		if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || calls != before {
			t.Fatal("unconfirmed request")
		}
		command = append(command, "--yes")
		out.Reset()
		diagnostic.Reset()
		if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != before+1 {
			t.Fatalf("action: %d %s", code, &diagnostic)
		}
		expected := []string{`"limit":9007199254740993`, `"credentialStatus":"configured"`}
		if action == "test" {
			expected = []string{`"status":"failed"`, `"message":"Check provider settings"`, `"testedAt":"today"`, `"providerProfileId":"profile"`, `"capability":"inference"`, `"providerKind":"compatible"`}
		}
		for _, field := range append(expected, `"requestId":"trace"`) {
			if !strings.Contains(out.String(), field) {
				t.Fatalf("missing %s: %s", field, &out)
			}
		}
		for _, failure := range []int{401, 403, 409, 503} {
			status = failure
			before = calls
			out.Reset()
			diagnostic.Reset()
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || calls != before+1 || strings.Contains(out.String()+diagnostic.String(), "private-provider-secret") {
				t.Fatal("unsafe failure or automatic retry")
			}
		}
		status = 200
		before = calls
		if code := Run(context.Background(), append(command, "--tenant", "other"), getenv, &out, &diagnostic); code == 0 || calls != before+1 {
			t.Fatal("scope bypass")
		}
	}
}

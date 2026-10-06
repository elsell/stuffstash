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

func TestInvitationCreateRequiresConfirmedScopeAndShowsOneTimeLink(t *testing.T) {
	calls := 0
	denied := false
	missingLink := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		body, _ := io.ReadAll(r.Body)
		if r.Method != "POST" || r.URL.Path != "/tenants/home/inventories/garage/access-invitations" || r.Header.Get("Authorization") != "Bearer owner" || string(body) != `{"email":"friend@example.test","relationship":"editor"}` {
			t.Errorf("request %s %s %s", r.Method, r.URL, body)
		}
		w.Header().Set("Content-Type", "application/json")
		if denied {
			w.WriteHeader(403)
			io.WriteString(w, `{"message":"secret error"}`)
			return
		}
		if missingLink {
			w.WriteHeader(201)
			io.WriteString(w, `{"data":{"id":"invite"}}`)
			return
		}
		w.WriteHeader(201)
		io.WriteString(w, `{"data":{"id":"invite","tenantId":"home","inventoryId":"garage","email":"friend@example.test","relationship":"editor","status":"pending","expiresAt":"later","isExpired":false,"inviterPrincipalId":"owner","inviteUrl":"https://stash.example/invite/once"},"meta":{"requestId":"trace"}}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	store := credentials.File{Path: filepath.Join(dir, "session.json")}
	if err := store.Save(context.Background(), ports.Session{Server: server.URL, Issuer: "https://id.example", Subject: "owner", IDToken: "owner", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	env := func(k string) string {
		switch k {
		case "STUFF_STASH_CLI_SERVER":
			return server.URL
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		case "STUFF_STASH_CLI_CREDENTIAL_FILE":
			return store.Path
		case "STUFF_STASH_CLI_CONFIG_FILE":
			return filepath.Join(dir, "contexts.json")
		}
		return ""
	}
	args := []string{"invitations", "create", "--tenant", "home", "--inventory", "garage", "--email", "friend@example.test", "--role", "editor", "--no-input", "--json"}
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), args, env, &out, &diagnostic); code != 2 || calls != 0 {
		t.Fatalf("unconfirmed %d %d %s", code, calls, &diagnostic)
	}
	out.Reset()
	diagnostic.Reset()
	args = append(args, "--yes")
	if code := Run(context.Background(), args, env, &out, &diagnostic); code != 0 || calls != 1 || !strings.Contains(out.String(), `"inviteUrl":"https://stash.example/invite/once"`) || strings.Contains(diagnostic.String(), "/invite/once") {
		t.Fatalf("create %d %d %s %s", code, calls, &out, &diagnostic)
	}
	missingLink = true
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), args, env, &out, &diagnostic); code != 1 || out.Len() != 0 || !strings.Contains(diagnostic.String(), "result is unknown") {
		t.Fatalf("missing link %d %s %s", code, &out, &diagnostic)
	}
	missingLink = false
	denied = true
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), args, env, &out, &diagnostic); code != 1 || calls != 3 || out.Len() != 0 || !strings.Contains(diagnostic.String(), `"category":"forbidden"`) || strings.Contains(diagnostic.String(), "secret error") {
		t.Fatalf("denied %d %s %s", code, &out, &diagnostic)
	}
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), []string{"invitations", "create", "--no-input", "--yes"}, env, &out, &diagnostic); code != 2 || calls != 3 {
		t.Fatalf("missing input %d %s", code, &diagnostic)
	}
}

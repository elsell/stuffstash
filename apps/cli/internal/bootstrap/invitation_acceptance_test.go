package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestInvitationAcceptanceProtectsTokenAndScope(t *testing.T) {
	calls := 0
	path := "/tenants/home/inventories/garage/access-invitations/invite/"
	accepted := 0
	broken := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if !strings.HasPrefix(r.URL.Path, path) || r.Header.Get("Authorization") != "Bearer owner" {
			w.WriteHeader(403)
			return
		}
		var body map[string]string
		if r.Method != "POST" || json.NewDecoder(r.Body).Decode(&body) != nil || body["acceptanceToken"] != "private-token" {
			t.Error("wrong token request")
		}
		switch strings.TrimPrefix(r.URL.Path, path) {
		case "preview":
			io.WriteString(w, `{"data":{"inventoryId":"garage","inventoryName":"Garage","relationship":"viewer","status":"pending","expiresAt":"2030-01-01T00:00:00Z","isExpired":false},"meta":{"requestId":"preview"}}`)
		case "accept":
			accepted++
			if broken {
				io.WriteString(w, "private-token invalid response")
				return
			}
			io.WriteString(w, `{"data":{"invitation":{"id":"invite","tenantId":"home","inventoryId":"garage","email":"owner@example.test","relationship":"viewer","status":"accepted","expiresAt":"2030-01-01T00:00:00Z","isExpired":false,"inviterPrincipalId":"sender","acceptedPrincipalId":"owner"},"grant":{"principalId":"owner","tenantId":"home","inventoryId":"garage","relationship":"viewer"}},"meta":{"requestId":"accepted"}}`)
		default:
			t.Error("wrong route")
		}
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

	input := filepath.Join(dir, "invite.json")
	os.WriteFile(input, []byte(`{"acceptanceToken":"private-token"}`), 0600)
	command := []string{"invitations", "accept", "invite", "--input", input, "--tenant", "home", "--inventory", "garage", "--json", "--no-input"}
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || accepted != 0 {
		t.Fatal("unconfirmed acceptance")
	}
	command = append(command, "--yes")
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || accepted != 1 {
		t.Fatalf("accept: %d %s", code, &diagnostic)
	}
	for _, field := range []string{`"acceptedPrincipalId":"owner"`, `"grant":`, `"requestId":"accepted"`} {
		if !strings.Contains(out.String(), field) {
			t.Fatalf("lost field %s", &out)
		}
	}
	if strings.Contains(out.String()+diagnostic.String(), "private-token") {
		t.Fatal("token exposed")
	}
	for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
		if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
			t.Fatal("scope bypass")
		}
	}
	if accepted != 1 {
		t.Fatal("denied preview accepted")
	}
	previewCommand := append([]string(nil), command...)
	previewCommand[1] = "preview"
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), previewCommand, getenv, &out, &diagnostic); code != 0 || !strings.Contains(out.String(), `"inventoryName":"Garage"`) || accepted != 1 {
		t.Fatalf("preview failed %s", &diagnostic)
	}
	broken = true
	out.Reset()
	diagnostic.Reset()
	beforeBroken := calls
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || calls != beforeBroken+2 || accepted != 2 || !strings.Contains(diagnostic.String(), "before you retry") {
		t.Fatal("uncertain acceptance retried or not explained")
	}
	if strings.Contains(out.String()+diagnostic.String(), "private-token") {
		t.Fatal("error exposed token")
	}

	os.WriteFile(input, []byte(`{"acceptanceToken":""}`), 0600)
	before := calls
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || calls != before {
		t.Fatal("empty token reached API")
	}
}

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

func TestPrintCancellationRequiresConfirmationAndCurrentRevision(t *testing.T) {
	reads, writes := 0, 0
	fail := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := "/tenants/home/inventories/garage/print-jobs/job"
		if r.Header.Get("Authorization") != "Bearer owner" || (r.URL.Path != path && r.URL.Path != path+"/cancellation") {
			w.WriteHeader(403)
			return
		}
		if r.Method == "GET" && r.URL.Path == path {
			reads++
		} else if r.Method == "POST" && r.URL.Path == path+"/cancellation" {
			writes++
			var body struct {
				Revision int64 `json:"revision"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil || body.Revision != 7 {
				t.Error("lost current revision")
			}
			if fail {
				w.WriteHeader(409)
				io.WriteString(w, "private-detail")
				return
			}
		} else {
			t.Error("wrong request")
		}
		io.WriteString(w, `{"data":{"id":"job","printerId":"printer","status":"cancelled","revision":7,"copies":1,"createdAt":"2026-10-05T10:00:00Z","updatedAt":"2026-10-05T11:00:00Z","attempts":[]},"meta":{"requestId":"trace"}}`)
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

	command := []string{"print-jobs", "cancel", "job", "--tenant", "home", "--inventory", "garage", "--json", "--no-input"}
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || reads != 0 || writes != 0 {
		t.Fatal("unconfirmed cancellation reached server")
	}
	command = append(command, "--yes")
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || reads != 1 || writes != 1 || !strings.Contains(out.String(), `"requestId":"trace"`) || !strings.Contains(diagnostic.String(), "A label might already have printed.") {
		t.Fatalf("cancel: %d %s", code, &diagnostic)
	}
	for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
		if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 || writes != 1 {
			t.Fatal("scope bypass")
		}
	}
	fail = true
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || writes != 2 || reads != 2 || strings.Contains(diagnostic.String(), "private-detail") {
		t.Fatal("unsafe conflict or retry")
	}
}

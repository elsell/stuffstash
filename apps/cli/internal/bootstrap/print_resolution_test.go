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

func TestPrintResolutionExplicitEvidence(t *testing.T) {
	calls := 0
	fail := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer owner" || r.URL.Path != "/tenants/home/inventories/garage/print-jobs/job/resolution" {
			w.WriteHeader(403)
			return
		}
		if r.Method != "POST" {
			t.Error("wrong method")
		}
		var input struct {
			Revision int64  `json:"revision"`
			Outcome  string `json:"reportedOutcome"`
			Ack      bool   `json:"acknowledgeUncertainty"`
		}
		if json.NewDecoder(r.Body).Decode(&input) != nil || input.Revision != 7 || input.Outcome != "not_printed" || !input.Ack {
			t.Error("lost evidence or revision")
		}
		if fail {
			w.WriteHeader(409)
			return
		}
		io.WriteString(w, `{"data":{"id":"job","printerId":"printer","status":"resolved","revision":8,"copies":1,"createdAt":"2026-10-05T10:00:00Z","updatedAt":"2026-10-05T11:00:00Z","resolution":{"reportedOutcome":"not_printed","resolvedBy":"owner","resolvedAt":"2026-10-05T11:00:00Z"},"attempts":[]},"meta":{"requestId":"trace"}}`)
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

	input := filepath.Join(dir, "input.json")
	if err := os.WriteFile(input, []byte(`{"reportedOutcome":"not_printed","revision":7,"acknowledgeUncertainty":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	command := []string{"print-jobs", "resolve", "job", "--tenant", "home", "--inventory", "garage", "--input", input, "--json", "--no-input"}
	var out, diagnostic bytes.Buffer
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || calls != 0 {
		t.Fatal("unconfirmed resolution")
	}
	command = append(command, "--yes")
	out.Reset()
	diagnostic.Reset()
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 || !strings.Contains(out.String(), `"reportedOutcome":"not_printed"`) {
		t.Fatalf("resolve: %d %s %s", code, &out, &diagnostic)
	}
	for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
		if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
			t.Fatal("scope bypass")
		}
	}
	fail = true
	before := calls
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || calls != before+1 {
		t.Fatal("conflict retried")
	}
	if err := os.WriteFile(input, []byte(`{"reportedOutcome":"printed","revision":7,"acknowledgeUncertainty":false}`), 0600); err != nil {
		t.Fatal(err)
	}
	before = calls
	if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || calls != before {
		t.Fatal("unacknowledged evidence sent")
	}
}

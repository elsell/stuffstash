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

func TestTelemetryExplicitSubmission(t *testing.T) {
	body := `{"$schema":"input","measurements":[{"platform":"ios","operation":"request","surface":"home","variant":"none","outcome":"success","durationMs":0.125}]}`
	calls, status := 0, 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.URL.Path != "/client-telemetry" || r.Header.Get("Authorization") != "Bearer owner" {
			t.Error("Incorrect authenticated request")
		}
		raw, _ := io.ReadAll(r.Body)
		if string(raw) != body {
			t.Errorf("Changed input: %s", raw)
		}
		w.WriteHeader(status)
		if status != 200 {
			io.WriteString(w, "private-server-data")
			return
		}
		io.WriteString(w, `{"$schema":"result","data":{"accepted":1,"secret":"private-server-data"},"meta":{"requestId":"trace"}}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	store := credentials.File{Path: filepath.Join(dir, "session.json")}
	if err := store.Save(context.Background(), ports.Session{Server: server.URL, Subject: "owner", IDToken: "owner", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(dir, "input.json")
	getenv := func(k string) string {
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
	run := func(yes bool) (int, string, string) {
		var out, diag bytes.Buffer
		args := []string{"telemetry", "submit", "--input", input, "--json", "--no-input"}
		if yes {
			args = append(args, "--yes")
		}
		code := Run(context.Background(), args, getenv, &out, &diag)
		return code, out.String(), diag.String()
	}
	if err := os.WriteFile(input, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := run(false); code == 0 || calls != 0 {
		t.Fatal("Submitted without confirmation")
	}
	code, out, diag := run(true)
	var response struct {
		Schema string `json:"$schema"`
		Data   struct {
			Accepted int `json:"accepted"`
		} `json:"data"`
		Meta struct {
			RequestID string `json:"requestId"`
		} `json:"meta"`
	}
	if err := json.Unmarshal([]byte(out), &response); err != nil {
		t.Fatal(err)
	}
	if code != 0 || calls != 1 || response.Schema != "result" || response.Data.Accepted != 1 || response.Meta.RequestID != "trace" {
		t.Fatalf("Submission failed: %d %s %s", code, out, diag)
	}
	if strings.Contains(out+diag, "private-server-data") {
		t.Fatal("Server data leaked")
	}
	for _, denial := range []int{401, 403, 500} {
		status = denial
		before := calls
		code, out, diag = run(true)
		if code == 0 || calls != before+1 || strings.Contains(out+diag, "private-server-data") {
			t.Fatal("Denial retried or leaked data")
		}
	}
	if err := store.Delete(context.Background(), server.URL); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{`{"measurements":[]}`, strings.Replace(body, `"durationMs":0.125`, `"durationMs":null`, 1), strings.Replace(body, `"ios"`, `"cli"`, 1), strings.Replace(body, `0.125`, `60001`, 1), `{"measurements":null}`, strings.Replace(body, `"durationMs":0.125`, `"other":0`, 1)} {
		if err := os.WriteFile(input, []byte(invalid), 0600); err != nil {
			t.Fatal(err)
		}
		before := calls
		code, _, diag = run(true)
		if code != 2 || calls != before || strings.Contains(diag, "sign in") {
			t.Fatalf("Invalid input reached credentials: %d %s", code, diag)
		}
	}
}

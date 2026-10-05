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

func TestPrintSubmissionExactRequestAndRetryBoundary(t *testing.T) {
	body := `{ "$schema":"print-schema", "printerId":"printer", "expectedMediaFingerprint":"reviewed-media", "templateId":"asset", "templateVersion":4294967295, "templateOptions":{"showReference":false}, "copies":9007199254740993, "previewFingerprint":"reviewed-preview" }`
	calls, status := 0, 201
	var path, key string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" {
			t.Error("explicit request performed a lookup")
		}
		if r.Header.Get("Authorization") != "Bearer owner" || r.URL.Path != path {
			w.WriteHeader(403)
			io.WriteString(w, "private-secret")
			return
		}
		if r.Header.Get("Idempotency-Key") == "" || key != "" && r.Header.Get("Idempotency-Key") != key {
			t.Error("retry key changed")
		}
		key = r.Header.Get("Idempotency-Key")
		if r.Header.Get("X-Request-ID") != "trace-client" {
			t.Error("request correlation missing")
		}
		data, err := io.ReadAll(r.Body)
		if err != nil || string(data) != body {
			t.Errorf("explicit body changed: %s %v", data, err)
		}
		w.WriteHeader(status)
		if status != 200 && status != 201 {
			io.WriteString(w, "private-secret")
			return
		}
		io.WriteString(w, `{"$schema":"job-schema","data":{"id":"job","printerId":"printer","status":"queued","revision":1,"copies":1,"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z","attempts":[]},"meta":{"requestId":"trace"}}`)
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
	if err := os.WriteFile(input, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ command, path string }{{"labels print asset", "assets/asset/print-jobs"}, {"printers test printer", "printers/printer/test-jobs"}, {"print-jobs reprint prior", "print-jobs/prior/reprints"}} {
		t.Run(tc.command, func(t *testing.T) {
			path = "/tenants/home/inventories/garage/" + tc.path
			status = 201
			key = ""
			before := calls
			args := append(strings.Fields(tc.command), "--tenant", "home", "--inventory", "garage", "--input", input, "--json", "--no-input", "--request-id", "trace-client")
			var out, diag bytes.Buffer
			if code := Run(context.Background(), args, getenv, &out, &diag); code != 0 || calls != before+1 {
				t.Fatalf("submission: %d %s", code, &diag)
			}
			if key == "" || !strings.Contains(diag.String(), key) || !strings.Contains(out.String(), `"$schema":"job-schema"`) || !strings.Contains(out.String(), `"requestId":"trace"`) {
				t.Fatal("retry key or response missing")
			}
			args = append(args, "--idempotency-key", key)
			status = 200
			before = calls
			out.Reset()
			diag.Reset()
			if code := Run(context.Background(), args, getenv, &out, &diag); code != 0 || calls != before+1 {
				t.Fatalf("same-key request: %d %s", code, &diag)
			}
			for _, code := range []int{401, 403, 409} {
				status = code
				before = calls
				out.Reset()
				diag.Reset()
				if exit := Run(context.Background(), args, getenv, &out, &diag); exit == 0 || calls != before+1 || strings.Contains(out.String()+diag.String(), "private-secret") {
					t.Fatalf("unsafe failure %d: %s", code, &diag)
				}
			}
			for _, scope := range []string{"--tenant", "--inventory"} {
				before = calls
				if code := Run(context.Background(), append(args, scope, "other"), getenv, &out, &diag); code == 0 || calls != before+1 {
					t.Fatal("cross-scope request allowed")
				}
			}
		})
	}
}

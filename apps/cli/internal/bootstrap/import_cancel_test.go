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

func TestImportCancellationModeConfirmationAndScope(t *testing.T) {
	for _, action := range []string{"keep_partial_progress", "discard_partial_progress"} {
		t.Run(action, func(t *testing.T) {
			calls := 0
			broken := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/tenants/home/inventories/garage/imports/jobs/job/cancel" || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if r.Method != "POST" {
					t.Error("wrong method")
				}
				var body map[string]string
				if json.NewDecoder(r.Body).Decode(&body) != nil || body["mode"] != action {
					t.Error("wrong cancellation mode")
				}
				if broken {
					io.WriteString(w, "invalid")
					return
				}
				io.WriteString(w, `{"data":{"id":"job","status":"cancelling","cancellationMode":"`+action+`","counts":{"recordsDiscarded":9007199254740993},"progress":{"phase":"cancellation","done":0,"total":1}},"meta":{"requestId":"trace"}}`)
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
			input := filepath.Join(dir, "cancel.json")
			os.WriteFile(input, []byte(`{"mode":"`+action+`"}`), 0600)
			command := append([]string{"import-jobs", "cancel", "job", "--input", input}, scope...)
			var out, diagnostic bytes.Buffer
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || calls != 0 {
				t.Fatal("unconfirmed operation reached API")
			}
			command = append(command, "--yes")
			out.Reset()
			diagnostic.Reset()
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 0 || calls != 1 || !strings.Contains(out.String(), `"status":"cancelling"`) || !strings.Contains(out.String(), `"recordsDiscarded":9007199254740993`) {
				t.Fatalf("operation: %d %d %s %s", code, calls, &out, &diagnostic)
			}
			for _, override := range [][]string{{"--tenant", "other"}, {"--inventory", "other"}} {
				if code := Run(context.Background(), append(command, override...), getenv, &out, &diagnostic); code == 0 {
					t.Fatal("scope boundary bypassed")
				}
			}
			broken = true
			out.Reset()
			diagnostic.Reset()
			before := calls
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code == 0 || calls != before+1 || !strings.Contains(diagnostic.String(), "before you try again") {
				t.Fatalf("unsafe recovery: %d %d %s", code, calls, &diagnostic)
			}
			os.WriteFile(input, []byte(`{"mode":"guess"}`), 0600)
			before = calls
			if code := Run(context.Background(), command, getenv, &out, &diagnostic); code != 2 || calls != before {
				t.Fatal("invalid mode reached API")
			}

		})
	}
}

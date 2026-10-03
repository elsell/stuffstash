package bootstrap

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLIProcessListsInventoryAndReportsSafeFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Error("missing user authentication")
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/tenants/tenant/inventories" {
			w.WriteHeader(403)
			io.WriteString(w, `{"error":{"message":"secret backend details"}}`)
			return
		}
		io.WriteString(w, `{"data":[{"id":"inventory","name":"Home","tenantId":"tenant","lifecycleState":"active","access":{}}],"meta":{"pagination":{"limit":1,"hasMore":true,"nextCursor":"next-page"}}}`)
	}))
	defer server.Close()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "stuffstash")
	build := exec.Command("go", "build", "-buildvcs=false", "-o", binary, "./cmd/stuffstash")
	build.Dir = "../.."
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %s %v", out, err)
	}
	path := filepath.Join(dir, "session.json")
	if err := (credentials.File{Path: path}).Save(context.Background(), ports.Session{Server: server.URL, IDToken: "user-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		tenant   string
		wantCode int
	}{{"tenant", 0}, {"other", 1}} {
		command := exec.Command(binary, "inventories", "list", "--tenant", test.tenant, "--server", server.URL, "--limit", "1", "--json")
		command.Env = append(os.Environ(), "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP=true", "STUFF_STASH_CLI_CREDENTIAL_FILE="+path)
		out, err := command.CombinedOutput()
		code := 0
		if err != nil {
			exit, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatal(err)
			}
			code = exit.ExitCode()
		}
		if code != test.wantCode {
			t.Fatalf("exit %d: %s", code, out)
		}
		if !json.Valid(out) || strings.Contains(string(out), "user-token") || strings.Contains(string(out), "secret backend") {
			t.Fatalf("unsafe output %s", out)
		}
		if code == 0 && !strings.Contains(string(out), "next-page") {
			t.Fatalf("pagination lost: %s", out)
		}
	}
}

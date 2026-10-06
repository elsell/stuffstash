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

func TestPrinterAdministrationContractAndDenials(t *testing.T) {
	printer := `{"id":"printer","name":"Label","adapterId":"adapter","revision":18446744073709551615,"media":{"name":"Roll","presetId":"preset","version":4294967295},"mediaFingerprint":"media","readiness":"ready","retired":false}`
	connector := `{"id":"connector","name":"USB","generation":18446744073709551615,"authorizationPending":false,"availability":"online","state":"active","printerIds":[],"report":null}`
	settings := `{"$schema":"nested-schema","defaultPrinterId":null,"printOnCreateDefault":false,"revision":18446744073709551615,"template":{"id":"template","version":4294967295,"options":{"showReference":false}}}`
	for _, tc := range []struct {
		command                  []string
		method, path, body, data string
	}{
		{[]string{"printers", "create"}, "POST", "/printers", `{"$schema":"schema","name":"Label","adapterId":"adapter","presetId":"preset","presetVersion":4294967295}`, printer},
		{[]string{"printers", "update", "printer"}, "PATCH", "/printers/printer", `{"$schema":"schema","revision":18446744073709551615,"retired":false,"name":null}`, printer},
		{[]string{"connectors", "print", "update", "connector"}, "PATCH", "/print-connectors/connector", `{"$schema":"schema","generation":18446744073709551615,"revoked":false,"printerIds":[]}`, connector},
		{[]string{"print-settings", "update"}, "PUT", "/print-settings", settings, settings},
	} {
		t.Run(strings.Join(tc.command, "_"), func(t *testing.T) {
			calls, status := 0, 200
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if tc.method == "POST" && r.Header.Get("Idempotency-Key") != "stable-printer-key" {
					t.Error("retry key lost")
				}
				if r.Method != tc.method {
					t.Error("write fetched or retried")
				}
				if r.URL.Path != "/tenants/home/inventories/garage"+tc.path || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				body, _ := io.ReadAll(r.Body)
				if string(body) != tc.body {
					t.Errorf("changed exact input: %s", body)
				}
				w.WriteHeader(status)
				if status != 200 {
					io.WriteString(w, "private-denial")
					return
				}
				io.WriteString(w, `{"$schema":"result-schema","data":`+tc.data+`,"meta":{"requestId":"trace","tenantId":"home"}}`)
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
			if err := os.WriteFile(input, []byte(tc.body), 0600); err != nil {
				t.Fatal(err)
			}
			args := append([]string{}, tc.command...)
			args = append(args, "--tenant", "home", "--inventory", "garage", "--input", input, "--json", "--no-input")
			var out, diag bytes.Buffer
			if code := Run(context.Background(), args, getenv, &out, &diag); code != 2 || calls != 0 {
				t.Fatalf("unconfirmed write: %d %s", code, &diag)
			}
			if tc.command[0] == "printers" && tc.command[1] == "create" {
				args = append(args, "--idempotency-key", "stable-printer-key")
			}
			args = append(args, "--yes")
			out.Reset()
			diag.Reset()
			if code := Run(context.Background(), args, getenv, &out, &diag); code != 0 || calls != 1 {
				t.Fatalf("write: %d %s", code, &diag)
			}
			for _, field := range []string{`"$schema":"result-schema"`, `"requestId":"trace"`, `"tenantId":"home"`, "18446744073709551615", "false"} {
				if !strings.Contains(out.String(), field) {
					t.Fatalf("missing %s: %s", field, &out)
				}
			}

			if tc.command[0] != "connectors" && !strings.Contains(out.String(), "4294967295") {
				t.Fatalf("unsigned version lost: %s", &out)
			}
			for _, denial := range []int{401, 403, 409, 500} {
				status = denial
				before := calls
				out.Reset()
				diag.Reset()
				if code := Run(context.Background(), args, getenv, &out, &diag); code == 0 || calls != before+1 || strings.Contains(out.String()+diag.String(), "private-denial") {
					t.Fatalf("unsafe denial/retry: %d %s", code, &diag)
				}
			}
			status = 200
			before := calls
			if code := Run(context.Background(), append(args, "--inventory", "other"), getenv, &out, &diag); code == 0 || calls != before+1 {
				t.Fatal("wrong scope succeeded")
			}
		})
	}
}

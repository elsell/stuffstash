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

func TestPairingApprovalHumanBoundary(t *testing.T) {
	for _, action := range []string{"review", "approve", "rotate"} {
		t.Run(action, func(t *testing.T) {
			calls, mutations, status := 0, 0, 200
			candidate := "candidate"
			rotation := action == "rotate"
			body := `{"$schema":"input","userCode":"private-code","tenantId":"home","inventoryId":"garage"}`
			command := []string{"connectors", "print", "pairings", action, "pairing"}
			if action == "approve" {
				body = `{"$schema":"input","userCode":"private-code","tenantId":"home","inventoryId":"garage","bindings":[{"candidateId":"candidate","printerId":"printer"}]}`
			}
			if action == "rotate" {
				body = `{"$schema":"input","generation":18446744073709551615,"pairingId":"pairing","userCode":"private-code"}`
				command = []string{"connectors", "print", "rotations", "approve", "connector"}
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer human" || r.Header.Get("X-Request-ID") != "trace" {
					t.Error("wrong principal/method/correlation")
					w.WriteHeader(403)
					return
				}
				raw, _ := io.ReadAll(r.Body)
				if r.URL.Path == "/print-connector-pairings/pairing/review" {
					if action == "review" && string(raw) != body {
						t.Error("review changed bytes")
					}
					for _, f := range []string{`"userCode":"private-code"`, `"tenantId":"home"`, `"inventoryId":"garage"`} {
						if !strings.Contains(string(raw), f) {
							t.Errorf("review missing %s", f)
						}
					}
					if status == 401 || status == 403 {
						w.WriteHeader(status)
						io.WriteString(w, "private-code")
						return
					}
					io.WriteString(w, `{"$schema":"public-review","data":{"id":"pairing","name":"Kitchen","rotation":`+map[bool]string{true: "true", false: "false"}[rotation]+`,"publicKeyFingerprint":"fingerprint","candidates":[{"id":"`+candidate+`","name":"USB\u001b printer","adapterId":"brother","deviceId":"private-device"}],"pollToken":"private-token"},"meta":{"requestId":"trace"}}`)
					return
				}
				mutations++
				if string(raw) != body {
					t.Errorf("approval changed bytes %s", raw)
				}
				expected := "/print-connector-pairings/pairing/approval"
				if action == "rotate" {
					expected = "/tenants/home/inventories/garage/print-connectors/connector/credential-rotation"
				}
				if r.URL.Path != expected {
					t.Errorf("wrong path %s", r.URL.Path)
				}
				w.WriteHeader(status)
				if status != 200 {
					io.WriteString(w, "private-code")
					return
				}
				if action == "rotate" {
					io.WriteString(w, `{"$schema":"pairing-status","data":{"id":"pairing","state":"approved","expiresAt":"2026-10-05T12:00:00Z","credential":"private-credential"},"meta":{"requestId":"trace"}}`)
					return
				}
				io.WriteString(w, `{"$schema":"connector","data":{"id":"connector","name":"Kitchen","generation":18446744073709551615,"authorizationPending":true,"availability":"unknown","state":"active","printerIds":["printer"],"lastSeenAt":null,"report":{"architecture":"amd64","platform":"linux","commit":"commit","version":"version","adapters":[{"id":"adapter","completionEvidence":"ack","contractVersions":[4294967295],"formats":["png"],"wake":false,"media":[{"id":"preset","version":4294967295}]}]},"reportReceivedAt":null,"credential":"private-credential"},"meta":{"requestId":"trace","tenantId":"home"}}`)
			}))
			defer server.Close()
			dir := t.TempDir()
			os.Chmod(dir, 0700)
			store := credentials.File{Path: filepath.Join(dir, "session")}
			if err := store.Save(context.Background(), ports.Session{Server: server.URL, Issuer: "issuer", Subject: "human", IDToken: "human", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(dir, "input.json")
			os.WriteFile(file, []byte(body), 0600)
			getenv := func(k string) string {
				switch k {
				case "STUFF_STASH_CLI_SERVER":
					return server.URL
				case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
					return "true"
				case "STUFF_STASH_CLI_CREDENTIAL_FILE":
					return store.Path
				case "STUFF_STASH_CLI_CONFIG_FILE":
					return filepath.Join(dir, "config", "contexts.json")
				case "STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE":
					return "must-not-read"
				}
				return ""
			}
			args := append(command, "--tenant", "home", "--inventory", "garage", "--input", file, "--json", "--no-input", "--request-id", "trace")
			run := func(extra ...string) (int, string) {
				var out, diag bytes.Buffer
				code := Run(context.Background(), append(append([]string{}, args...), extra...), getenv, &out, &diag)
				all := out.String() + diag.String()
				for _, secret := range []string{"private-code", "private-device", "private-token", "private-credential"} {
					if strings.Contains(all, secret) {
						t.Fatalf("secret leaked %s", all)
					}
				}
				return code, all
			}
			if action != "review" {
				if code, _ := run(); code == 0 || mutations != 0 {
					t.Fatal("unconfirmed mutation")
				}
				args = append(args, "--yes")
			}
			code, out := run()
			if code != 0 {
				t.Fatalf("approval %d %s", code, out)
			}
			want := `"publicKeyFingerprint":"fingerprint"`
			if action == "approve" {
				want = `"generation":18446744073709551615`
			}
			if action == "rotate" {
				want = `"expiresAt":"2026-10-05T12:00:00Z"`
				if strings.Contains(out, `"generation"`) {
					t.Fatal("rotation status invented connector fields")
				}
			}
			if !strings.Contains(out, want) || !strings.Contains(out, `"requestId":"trace"`) {
				t.Fatalf("lost output %s", out)
			}
			if action == "approve" {
				for _, field := range []string{`"lastSeenAt":null`, `"reportReceivedAt":null`, `"architecture":"amd64"`, `"contractVersions":[4294967295]`, `"wake":false`, `"version":4294967295`, `"authorizationPending":true`} {
					if !strings.Contains(out, field) {
						t.Fatalf("lost public field %s: %s", field, out)
					}
				}
			}
			if action != "rotate" {
				before := calls
				if code, _ := run("--tenant", "other"); code == 0 || calls != before {
					t.Fatal("mismatched input scope reached network")
				}
			}
			for _, denial := range []int{401, 403} {
				status = denial
				before := calls
				if code, _ := run(); code == 0 || calls != before+1 {
					t.Fatal("denial retried")
				}
			}
			if action != "review" {
				for _, failure := range []int{409, 500} {
					status = failure
					before := mutations
					if code, _ := run(); code == 0 || mutations != before+1 {
						t.Fatal("approval failure retried")
					}
				}
				status = 200
				if action == "approve" {
					candidate = "other"
					before := mutations
					if code, _ := run(); code == 0 || mutations != before {
						t.Fatal("unreviewed binding approved")
					}
					candidate = "candidate"
				}
				rotation = !rotation
				before := mutations
				if code, _ := run(); code == 0 || mutations != before {
					t.Fatal("wrong approval type accepted")
				}
			}
		})
	}
}

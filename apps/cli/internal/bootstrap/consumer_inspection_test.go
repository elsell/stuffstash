package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestConsumerInspectionBoundary(t *testing.T) {
	for _, tc := range []struct {
		command []string
		path    string
		data    string
		fields  []string
	}{
		{[]string{"printers"}, "/print-consumer/printers", `[{"bindingGeneration":18446744073709551615,"deviceId":"usb\u001bdevice","printer":{"id":"printer","name":"Kitchen","adapterId":"brother","revision":18446744073709551615,"retired":false,"readiness":"ready","readinessReason":null,"reportedAt":null,"mediaFingerprint":"fp","media":{"name":"roll","presetId":"preset","version":4294967295,"widthMicrometers":9007199254740993,"heightMicrometers":2,"resolutionDpi":3,"rasterWidth":4,"rasterHeight":5,"orientation":"portrait","colorMode":"mono","cutPolicy":"end","displayRotation":0,"marginsMicrometers":{"left":1,"right":2,"top":3,"bottom":4}}},"credential":"injected-secret"}]`, []string{`"bindingGeneration":18446744073709551615`, `"revision":18446744073709551615`, `"version":4294967295`, `"widthMicrometers":9007199254740993`, `"reportedAt":null`, `"readinessReason":null`, `"deviceId":"usb\u001bdevice"`}},
		{[]string{"attempts", "list", "--printer", "printer", "--status", "unsettled", "--limit", "1", "--cursor", "before"}, "/print-consumer/attempts", "[" + consumerAttemptFixture + "]", []string{`"revision":18446744073709551615`, `"version":4294967295`, `"copies":9007199254740993`, `"artifact":null`, `"startedAt":null`, `"nextCursor":"after"`}},
		{[]string{"attempts", "show", "attempt"}, "/print-consumer/attempts/attempt", consumerAttemptFixture, []string{`"revision":18446744073709551615`, `"version":4294967295`, `"copies":9007199254740993`, `"artifact":null`, `"startedAt":null`}},
	} {
		t.Run(tc.path, func(t *testing.T) {
			calls := 0
			status := 200
			redirected := 0
			responseData := tc.data
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirected++ }))
			defer target.Close()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "GET" || r.URL.Path != tc.path || r.Header.Get("Authorization") != "Bearer machine-secret" || r.Header.Get("X-Request-ID") != "trace" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(403)
					return
				}
				if tc.path == "/print-consumer/attempts" && r.URL.Query().Encode() != "cursor=before&limit=1&printerId=printer&status=unsettled" {
					t.Errorf("filters: %s", r.URL.RawQuery)
				}
				if status == 302 {
					http.Redirect(w, r, target.URL, status)
					return
				}
				w.WriteHeader(status)
				if status != 200 {
					io.WriteString(w, `{"credential":"injected-secret"}`)
					return
				}
				io.WriteString(w, `{"$schema":"consumer-schema","data":`+responseData+`,"meta":{"requestId":"trace","tenantId":"home","pagination":{"limit":1,"nextCursor":"after","hasMore":true}}}`)
			}))
			defer server.Close()
			dir := t.TempDir()
			os.Chmod(dir, 0700)
			store := credentials.ConnectorFile{Path: filepath.Join(dir, "connector.json")}
			registration := ports.ConnectorRegistration{Server: server.URL, ConnectorID: "connector", TenantID: "home", InventoryID: "garage", Credential: "machine-secret", ExpiresAt: time.Now().Add(time.Hour)}
			if err := store.Save(context.Background(), registration); err != nil {
				t.Fatal(err)
			}
			getenv := func(k string) string {
				switch k {
				case "STUFF_STASH_CLI_SERVER":
					return server.URL
				case "STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE":
					return store.Path
				case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
					return "true"
				case "STUFF_STASH_CLI_CONFIG_FILE", "STUFF_STASH_CLI_CREDENTIAL_FILE":
					return filepath.Join(dir, "must-not-read")
				}
				return ""
			}
			base := append([]string{"connectors", "print"}, tc.command...)
			base = append(base, "--connector", "connector", "--request-id", "trace", "--no-input")
			run := func(extra ...string) (int, string) {
				var out, err bytes.Buffer
				code := Run(context.Background(), append(append([]string{}, base...), extra...), getenv, &out, &err)
				s := out.String() + err.String()
				if strings.Contains(s, "machine-secret") || strings.Contains(s, "injected-secret") {
					t.Fatalf("secret output %s", s)
				}
				return code, s
			}
			code, out := run("--json")
			if code != 0 {
				t.Fatalf("read %d: %s", code, out)
			}
			decode := func(body string) map[string]any {
				t.Helper()
				var value map[string]any
				decoder := json.NewDecoder(strings.NewReader(body))
				decoder.UseNumber()
				if err := decoder.Decode(&value); err != nil {
					t.Fatal(err)
				}
				return value
			}
			expected := decode(`{"data":` + tc.data + `}`)
			switch data := expected["data"].(type) {
			case []any:
				for _, v := range data {
					delete(v.(map[string]any), "credential")
					delete(v.(map[string]any), "claimToken")
				}
			case map[string]any:
				delete(data, "claimToken")
			}
			if actual := decode(out)["data"]; !reflect.DeepEqual(actual, expected["data"]) {
				t.Fatalf("declared response fields changed: got %#v want %#v", actual, expected["data"])
			}

			for _, f := range append(tc.fields, `"$schema":"consumer-schema"`, `"requestId":"trace"`) {
				if !strings.Contains(out, f) {
					t.Fatalf("missing %s: %s", f, out)
				}
			}
			if code, out := run(); code != 0 || strings.ContainsRune(out, '\x1b') {
				t.Fatalf("human %d: %s", code, out)
			}
			for _, extra := range [][]string{{"--connector", "wrong"}, {"--server", target.URL}, {"--tenant", "other"}, {"--inventory", "other"}, {"--context", "human"}, {"--credential-file", "human"}, {"--input", "secret"}} {
				before := calls
				if code, _ := run(extra...); code == 0 || calls != before {
					t.Fatalf("identity/scope accepted %v", extra)
				}
			}
			if tc.path != "/print-consumer/printers" {
				responseData = strings.Replace(tc.data, `"artifact":null`, `"artifact":{"sha256":"sha","contentType":"image/png","byteLength":9223372036854775807,"widthPixels":9007199254740993,"heightPixels":9,"expiresAt":"2026-10-05T12:00:00Z","claimToken":"injected-secret"}`, 1)
				if code, out := run("--json"); code != 0 || !strings.Contains(out, `"byteLength":9223372036854775807`) || !strings.Contains(out, `"widthPixels":9007199254740993`) {
					t.Fatalf("artifact %d %s", code, out)
				}
				responseData = strings.Replace(tc.data, `,"artifact":null`, "", 1)
				responseData = strings.Replace(responseData, `,"startedAt":null`, "", 1)
				if code, out := run("--json"); code != 0 || strings.Contains(out, `"artifact"`) || strings.Contains(out, `"startedAt"`) {
					t.Fatalf("absent optional %d %s", code, out)
				}
			}
			for _, s := range []int{401, 403, 302} {
				status = s
				before := calls
				if code, _ := run("--json"); code == 0 || calls != before+1 {
					t.Fatalf("status %d retried or accepted", s)
				}
			}
			if redirected != 0 {
				t.Fatal("followed redirect")
			}
			registration.ExpiresAt = time.Now().Add(-time.Hour)
			if err := store.Save(context.Background(), registration); err != nil {
				t.Fatal(err)
			}
			before := calls
			if code, _ := run(); code == 0 || calls != before {
				t.Fatal("expired credential used")
			}
		})
	}
}

const consumerAttemptFixture = `{"protocolVersion":9007199254740993,"jobId":"job","printerId":"printer","attemptId":"attempt","sessionId":"session","status":"uncertain","revision":18446744073709551615,"copies":9007199254740993,"leaseExpiresAt":"2026-10-05T12:00:00Z","leaseValid":false,"startedAt":null,"settledAt":null,"resolvedAt":null,"outcome":{"kind":"uncertain","completedCopies":9007199254740993,"retryable":false,"reason":"unknown"},"mediaFingerprint":"fp","media":{"presetId":"preset","version":4294967295,"widthMicrometers":1,"heightMicrometers":2,"resolutionDpi":3,"rasterWidth":4,"rasterHeight":5,"orientation":"portrait","colorMode":"mono","cutPolicy":"end","displayRotation":0,"marginsMicrometers":{"left":1,"right":2,"top":3,"bottom":4}},"artifact":null,"claimToken":"injected-secret"}`

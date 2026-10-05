package httpapi_test

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestWorkerFullUnsignedTransport(t *testing.T) {
	const revision = "18446744073709551615"
	attempt := workerFullRangeClaimFixture
	calls := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls[r.URL.Path]++
		if r.Header.Get("Authorization") != "Bearer machine-secret" {
			t.Error("machine authorization lost")
			w.WriteHeader(401)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/content") {
			if r.Header.Get("X-Print-Revision") != revision || r.Header.Get("X-Print-Claim-Token") != "claim-token" || r.Header.Get("X-Print-Session-ID") != "session" {
				t.Errorf("artifact proof changed: %v", r.Header)
			}
			w.Header().Set("Content-Type", "image/png")
			io.WriteString(w, "png")
			return
		}
		if r.Method == "POST" {
			raw, _ := io.ReadAll(r.Body)
			var body map[string]json.RawMessage
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Fatal(err)
			}
			switch r.URL.Path {
			case "/print-consumer/claims":
				for _, name := range []string{"attemptId", "sessionId", "claimToken", "printerId"} {
					if len(body[name]) == 0 {
						t.Errorf("missing proof %s", name)
					}
				}
			case "/print-consumer/heartbeat":
				for _, part := range []string{`"contractVersions":[4294967295]`, `"version":4294967295`, `"sessionId":"session"`} {
					if !strings.Contains(string(raw), part) {
						t.Errorf("heartbeat narrowing %s", raw)
					}
				}
			default:
				if string(body["revision"]) != revision {
					t.Errorf("revision narrowed: %s", raw)
				}
				if strings.Contains(r.URL.Path, "/claims/") {
					if string(body["sessionId"]) != `"session"` || string(body["claimToken"]) != `"claim-token"` {
						t.Errorf("proof identity changed %s", raw)
					}
				}
			}
		}
		switch r.URL.Path {
		case "/print-consumer/printers":
			io.WriteString(w, `{"data":[{"bindingGeneration":`+revision+`,"deviceId":"usb","printer":{"id":"printer","adapterId":"adapter","revision":`+revision+`,"media":{"presetId":"media","version":4294967295}}}]}`)
		case "/print-consumer/heartbeat":
			io.WriteString(w, `{"data":{"id":"connector","generation":`+revision+`}}`)
		case "/print-consumer/attempts":
			if r.URL.Query().Get("printerId") != "printer" || r.URL.Query().Get("status") != "unsettled" {
				t.Error("recovery filter changed")
			}
			io.WriteString(w, `{"data":[`+attempt+`],"meta":{"pagination":{"hasMore":false,"nextCursor":null}}}`)
		default:
			io.WriteString(w, `{"data":`+attempt+`}`)
		}
	}))
	defer server.Close()
	client, err := httpapi.New(server.URL, "machine-secret", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	control := printing.AttemptControl{AttemptID: "attempt", SessionID: "session", ClaimToken: "claim-token", Revision: math.MaxUint64}
	claim, err := client.Claim(ctx, "printer", control)
	if err != nil || claim == nil {
		t.Fatalf("claim %v", err)
	}
	if claim.Control.Revision != math.MaxUint64 || claim.Media.Version != math.MaxUint32 {
		t.Fatalf("claim narrowed %+v", claim)
	}
	for _, call := range []func(context.Context, printing.AttemptControl) (printing.AttemptStatus, error){client.Start, client.Renew} {
		value, err := call(ctx, control)
		if err != nil || value.Revision != math.MaxUint64 {
			t.Fatalf("proof/status %v %+v", err, value)
		}
	}
	value, err := client.Attempt(ctx, "attempt")
	if err != nil || value.Revision != math.MaxUint64 {
		t.Fatalf("attempt %v %+v", err, value)
	}
	values, err := client.Unsettled(ctx, "printer")
	if err != nil || len(values) != 1 || values[0].Revision != math.MaxUint64 {
		t.Fatalf("unsettled %v %+v", err, values)
	}
	evidence := printing.Evidence{Outcome: printing.NoOutput, Reason: printing.Disconnected}
	if err := client.Outcome(ctx, control, evidence); err != nil {
		t.Fatal(err)
	}
	if err := client.Reconcile(ctx, "attempt", math.MaxUint64, evidence); err != nil {
		t.Fatal(err)
	}
	if err := client.ConfirmIdle(ctx, "attempt", math.MaxUint64); err != nil {
		t.Fatal(err)
	}
	if data, _, err := client.Artifact(ctx, control, 10); err != nil || string(data) != "png" {
		t.Fatalf("artifact %v", err)
	}
	printers, err := client.Printers(ctx)
	if err != nil || len(printers) != 1 || printers[0].BindingGeneration != math.MaxUint64 || printers[0].Media.Version != math.MaxUint32 {
		t.Fatalf("printers %v %+v", err, printers)
	}
	maxVersion, _ := strconv.ParseInt("4294967295", 10, 64)
	report := &printing.ConnectorReport{Version: "1", Commit: "commit", Platform: "linux", Architecture: "amd64", Adapters: []printing.Descriptor{{ID: "adapter", ContractVersions: []int{int(maxVersion)}, Media: []printing.Media{{PresetID: "media", Version: math.MaxUint32}}}}}
	if err := client.Heartbeat(ctx, "session", report); err != nil {
		t.Fatal(err)
	}
	for path, count := range calls {
		if count != 1 {
			t.Errorf("unexpected retry %s: %d", path, count)
		}
	}
}

func TestWorkerRejectsInvalidWireNumbersAndTimes(t *testing.T) {
	cases := []struct{ name, old, value string }{
		{"valid baseline", "", ""},
		{"revision overflow", `"revision":18446744073709551615`, `"revision":18446744073709551616`},
		{"negative revision", `"revision":18446744073709551615`, `"revision":-1`},
		{"media overflow", `"version":4294967295`, `"version":4294967296`},
		{"negative media", `"version":4294967295`, `"version":-1`},
		{"invalid lease time", `"leaseExpiresAt":"2026-10-05T12:00:00Z"`, `"leaseExpiresAt":"not-a-time"`},
		{"invalid artifact time", `"expiresAt":"2026-10-05T12:00:00Z"`, `"expiresAt":"not-a-time"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := workerFullRangeClaimFixture
			if tc.old != "" {
				if strings.Count(data, tc.old) != 1 {
					t.Fatal("fixture mutation must replace exactly one field")
				}
				data = strings.Replace(data, tc.old, tc.value, 1)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, `{"data":`+data+`}`) }))
			defer server.Close()
			client, err := httpapi.New(server.URL, "machine-secret", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			claim, err := client.Claim(context.Background(), "printer", printing.AttemptControl{AttemptID: "attempt", SessionID: "session", ClaimToken: "claim-token"})
			if tc.old == "" {
				if err != nil || claim == nil {
					t.Fatalf("valid baseline rejected: %v", err)
				}
			} else if err == nil {
				t.Fatal("invalid transport value accepted")
			}
		})
	}
}

const workerFullRangeClaimFixture = `{"attemptId":"attempt","sessionId":"session","jobId":"job","printerId":"printer","revision":18446744073709551615,"status":"claimed","leaseValid":true,"leaseExpiresAt":"2026-10-05T12:00:00Z","protocolVersion":1,"copies":1,"outcome":{"kind":"no_output","completedCopies":0,"retryable":true,"reason":""},"media":{"presetId":"media","version":4294967295},"artifact":{"sha256":"hash","contentType":"image/png","byteLength":3,"widthPixels":1,"heightPixels":1,"expiresAt":"2026-10-05T12:00:00Z"}}`

package httpapi

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type recordedProtocolReceipts struct{ values []ports.ProtocolReceipt }

func (r *recordedProtocolReceipts) Record(_ context.Context, v ports.ProtocolReceipt) {
	r.values = append(r.values, v)
}
func TestWorkerReceiptsPreserveSafeEnvelopeAndDoNotEmitFailedRequests(t *testing.T) {
	status := 200
	body := `{"$schema":"attempt-schema","meta":{"requestId":"request","tenantId":"household"},"data":{"attemptId":"attempt","revision":18446744073709551615,"leaseExpiresAt":"2026-01-01T00:00:00Z","startedAt":null,"settledAt":null,"resolvedAt":null,"media":null,"artifact":null,"claimToken":"secret-never-print"}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer credential" {
			t.Error("missing credential")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()
	sink := &recordedProtocolReceipts{}
	client, err := New(server.URL, "credential", server.Client(), Options{Receipts: sink})
	if err != nil {
		t.Fatal(err)
	}
	if err = client.Outcome(context.Background(), printing.AttemptControl{}, printing.Evidence{}); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(sink.values)
	for _, want := range []string{`18446744073709551615`, `"startedAt":null`, `"artifact":null`, `attempt-schema`, `household`} {
		if !strings.Contains(string(encoded), want) {
			t.Fatalf("missing %s: %s", want, encoded)
		}
	}
	if strings.Contains(string(encoded), "secret-never-print") {
		t.Fatal("secret leaked")
	}
	for _, code := range []int{401, 403, 409} {
		status = code
		if client.Outcome(context.Background(), printing.AttemptControl{}, printing.Evidence{}) == nil {
			t.Fatal("expected failure")
		}
	}
	if len(sink.values) != 1 {
		t.Fatal("failed requests emitted receipts")
	}
	status = 200
	body = `{"data":null,"meta":{"requestId":"idle-poll"}}`
	claim, err := client.Claim(context.Background(), "printer", printing.AttemptControl{})
	if err != nil || claim != nil {
		t.Fatalf("null claim: %v %v", claim, err)
	}
	encoded, _ = json.Marshal(sink.values[1])
	if !strings.Contains(string(encoded), `"data":null`) {
		t.Fatalf("null lost: %s", encoded)
	}
}

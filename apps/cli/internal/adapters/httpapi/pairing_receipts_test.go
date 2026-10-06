package httpapi

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/pairingkeys"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type pairingReceiptRecorder struct{ receipts []ports.ProtocolReceipt }

func (r *pairingReceiptRecorder) Record(_ context.Context, v ports.ProtocolReceipt) {
	r.receipts = append(r.receipts, v)
}
func TestPairingReceiptsPreserveSafeEnvelope(t *testing.T) {
	recorder := &pairingReceiptRecorder{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := `{"id":"pair","pollToken":"private-poll-token","userCode":"CODE","verificationUrl":"https://web.example/pair","expiresAt":"2030-01-02T03:04:05Z"}`
		switch {
		case strings.HasSuffix(r.URL.Path, "/credential"):
			data = `{"connectorId":"connector","tenantId":"home","inventoryId":"garage","credential":"private-machine-secret","expiresAt":"2030-01-03T03:04:05Z","activationDeadline":"2030-01-02T03:05:05Z"}`
		case r.Method == "GET":
			data = `{"id":"pair","state":"approved","expiresAt":"2030-01-02T03:04:05Z"}`
		}
		w.Write([]byte(`{"$schema":"receipt-schema","data":` + data + `,"meta":{"requestId":"request-trace"}}`))
	}))
	defer server.Close()
	api, err := NewPairing(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	api.Receipts = recorder
	key, _ := (pairingkeys.Keys{}).NewKey()
	ctx := context.Background()
	challenge, err := api.Start(ctx, ports.PairingRequest{Name: "test", PublicKey: key.PublicKey()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = api.Poll(ctx, challenge); err != nil {
		t.Fatal(err)
	}
	if _, err = api.Exchange(ctx, challenge, key.Sign(challenge.ID, challenge.PollToken)); err != nil {
		t.Fatal(err)
	}
	if len(recorder.receipts) != 3 {
		t.Fatalf("receipt count %d", len(recorder.receipts))
	}
	for _, v := range recorder.receipts {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		if strings.Contains(s, "private-") || strings.Contains(s, "signature") || strings.Contains(s, "pollToken") {
			t.Fatal("secret serialized")
		}
		for _, field := range []string{"receipt-schema", "request-trace", "expiresAt"} {
			if !strings.Contains(s, field) {
				t.Fatalf("lost %s: %s", field, s)
			}
		}
	}
	b, _ := json.Marshal(recorder.receipts[2])
	for _, field := range []string{"activationDeadline", "connectorId", "tenantId", "inventoryId"} {
		if !strings.Contains(string(b), field) {
			t.Fatal("lost " + field)
		}
	}
}

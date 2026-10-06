package presentation

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
	"testing"
	"time"
)

func TestPairingReceiptHumanAndJSON(t *testing.T) {
	v := ports.PairingCredentialReceipt{Data: ports.SafePairingCredential{ConnectorID: "connector", TenantID: "home", InventoryID: "garage", ExpiresAt: time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC), ActivationDeadline: time.Date(2030, 1, 1, 3, 4, 5, 0, time.UTC)}}
	for _, jsonMode := range []bool{false, true} {
		var out, diagnostic bytes.Buffer
		sink := NewProtocolReceipts(Output{Stdout: &out, Stderr: &diagnostic, JSON: jsonMode})
		sink.Record(context.Background(), ports.ProtocolReceipt{Operation: "pairing.credential.exchanged", Result: v})
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if !sink.Close(ctx) {
			t.Fatal("receipt drain timeout")
		}
		cancel()
		for _, want := range []string{"connector", "home", "garage", "2030-01-01T03:04:05Z"} {
			if !strings.Contains(out.String(), want) {
				t.Fatalf("lost %s: %s", want, &out)
			}
		}
		if jsonMode && !strings.Contains(out.String(), `"result":{"data":`) {
			t.Fatal("missing structured envelope")
		}
	}
}

package printworker_test

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"testing"
)

type receiptRecorder struct{ values []ports.ProtocolReceipt }

func (r *receiptRecorder) Record(_ context.Context, v ports.ProtocolReceipt) {
	r.values = append(r.values, v)
}
func TestArtifactReceiptRequiresVerifiedContent(t *testing.T) {
	for _, corrupt := range []bool{false, true} {
		w, api, journal, printer := fixture(t)
		sink := &receiptRecorder{}
		w.Receipts = sink
		if corrupt {
			api.bytes[0] ^= 1
		}
		err := w.Step(context.Background(), journal, printer)
		if corrupt {
			if err == nil || len(sink.values) != 0 {
				t.Fatalf("corrupt artifact acknowledged: %v %#v", err, sink.values)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(sink.values) != 1 {
			t.Fatalf("receipts: %#v", sink.values)
		}
		receipt, ok := sink.values[0].Result.(ports.VerifiedPrintArtifactReceipt)
		if !ok || receipt.SHA256 != api.claim.Artifact.SHA256 || receipt.ByteLength != int64(len(api.bytes)) || receipt.AttemptID == "" {
			t.Fatalf("bad receipt: %#v", receipt)
		}
	}
}

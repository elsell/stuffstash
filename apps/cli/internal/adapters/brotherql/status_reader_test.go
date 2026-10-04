package brotherql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/fakes"
)

func TestCancelledStatusWaitDoesNotConsumeQueuedEvidence(t *testing.T) {
	device := fakes.NewPrinterTransport()
	reader := newStatusReader(device)
	defer device.Close()
	defer reader.close()
	if _, err := device.Write(context.Background(), []byte{0x1b, 0x69, 0x53}); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(time.Second)
	for len(reader.frames) == 0 {
		select {
		case <-deadline:
			t.Fatal("printer reply was not queued")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Both cancellation and a real device reply are ready. No call may select
	// the reply in preference to an already-cancelled operation.
	for i := 0; i < 64; i++ {
		if _, err := reader.next(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled wait consumed status: %v", err)
		}
	}
	live, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if reply, err := reader.next(live); err != nil || reply.kind != statusReply {
		t.Fatal("queued evidence lost", reply, err)
	}
}

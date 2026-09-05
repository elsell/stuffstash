package scheduling

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDelayHonorsCancellationAndElapsedInterval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (Delay{}).Wait(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled delay waited")
	}
	if err := (Delay{}).Wait(context.Background(), time.Nanosecond); err != nil {
		t.Fatal(err)
	}
}

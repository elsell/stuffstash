package observability

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestLiveTelemetryCollectorAcceptance(t *testing.T) {
	if os.Getenv("STUFF_STASH_TEST_LIVE_TELEMETRY") != "true" {
		t.Skip("live collector probe requires explicit opt-in")
	}
	cfg, err := config.LoadTelemetry()
	if err != nil {
		t.Fatal(err)
	}
	profiles, err := config.LoadProfiling()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled || !profiles.Enabled || cfg.ServiceName != "stuffstash-observability-probe" || profiles.ServiceName != cfg.ServiceName {
		t.Fatal("probe requires both signals enabled with isolated service identity")
	}
	ctx := context.Background()
	telemetry, err := NewRuntime(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer telemetry.Shutdown(ctx)
	observer := &probeProfileFailures{}
	profiler, err := NewProfiler(ctx, profiles, observer)
	if err != nil {
		t.Fatal(err)
	}
	defer profiler.Stop(ctx)
	operation, finish := telemetry.Telemetry.Start(ctx, ports.OperationHTTP)
	telemetry.Observer.Record(operation, ports.Event{Name: ports.EventHealthChecked})
	finish(nil)
	shutdown, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := profiler.Stop(shutdown); err != nil {
		t.Fatal(err)
	}
	if err := telemetry.Shutdown(shutdown); err != nil {
		t.Fatal(err)
	}
	if observer.failures.Load() != 0 {
		t.Fatal("profile delivery failed")
	}
	for _, count := range telemetry.DroppedBatches() {
		if count != 0 {
			t.Fatal("OTLP delivery failed")
		}
	}
}

type probeProfileFailures struct{ failures atomic.Int64 }

func (p *probeProfileFailures) Record(_ context.Context, event ports.Event) {
	if event.Name == ports.EventProfilingDeliveryFailed {
		p.failures.Add(1)
	}
}

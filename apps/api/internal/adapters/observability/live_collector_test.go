package observability

import (
	"context"
	"crypto/sha256"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"go.opentelemetry.io/otel/trace"
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
	cfg.SampleRatio = 1
	ctx := context.Background()
	telemetry, err := NewRuntime(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_ = telemetry.Shutdown(cleanup)
	}()
	observer := &probeProfileFailures{}
	profiler, err := NewProfiler(ctx, profiles, observer)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_ = profiler.Stop(cleanup)
	}()
	// Keep the CPU sampler active long enough to observe real work. This is a
	// fixed-input collector check, not an image-processing benchmark.
	payload := make([]byte, 256*1024)
	until := time.Now().Add(time.Second)
	var digest [32]byte
	for time.Now().Before(until) {
		digest = sha256.Sum256(payload)
		payload[0] = digest[0]
	}
	if digest == [32]byte{} {
		t.Fatal("CPU probe did not execute")
	}
	operation, finish := telemetry.Telemetry.Start(ctx, ports.OperationHTTP)
	if !trace.SpanContextFromContext(operation).IsSampled() {
		finish(nil)
		t.Fatal("probe span was not sampled")
	}
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

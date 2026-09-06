package observability

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRuntimeExportsAllSignalsAndFlushesOnShutdown(t *testing.T) {
	var mu sync.Mutex
	paths := map[string]int{}
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Basic controlled-test" {
			t.Error("exporter did not apply credential")
		}
		data, err := io.ReadAll(r.Body)
		if err != nil || len(data) == 0 {
			t.Error("missing OTLP payload")
		}
		mu.Lock()
		paths[r.URL.Path]++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	defer collector.Close()
	runtime, err := NewRuntime(context.Background(), config.TelemetryConfig{
		Enabled: true, Endpoint: collector.URL + "/otlp", ServiceName: "test-service", SampleRatio: 1,
		Headers:       map[string]string{"Authorization": "Basic controlled-test"},
		ExportTimeout: time.Second, BatchInterval: time.Hour, MetricInterval: time.Hour, QueueSize: 32, BatchSize: 16,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, finish := runtime.Telemetry.Start(context.Background(), ports.OperationThumbnailGenerate)
	runtime.Observer.Record(ctx, ports.Event{Name: ports.EventAttachmentThumbnailGenerated})
	finish(nil)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := runtime.Shutdown(shutdownCtx); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	for _, path := range []string{"/otlp/v1/traces", "/otlp/v1/metrics", "/otlp/v1/logs"} {
		if paths[path] == 0 {
			t.Errorf("missing signal %s", path)
		}
	}
}

func TestDisabledRuntimeDoesNotRequireExporter(t *testing.T) {
	runtime, err := NewRuntime(context.Background(), config.TelemetryConfig{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, finish := runtime.Telemetry.Start(context.Background(), ports.OperationHTTP)
	runtime.Observer.Record(ctx, ports.Event{Name: ports.EventHealthChecked})
	finish(nil)
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

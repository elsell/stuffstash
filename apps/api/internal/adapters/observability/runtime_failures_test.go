package observability

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestCollectorFailuresArePrivateAndCounted(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "private-collector-secret", http.StatusBadRequest)
	}))
	defer collector.Close()
	runtime, err := NewRuntime(context.Background(), config.TelemetryConfig{Enabled: true, Endpoint: collector.URL, ServiceName: "test", SampleRatio: 1, ExportTimeout: time.Second, BatchInterval: time.Hour, MetricInterval: time.Hour, QueueSize: 16, BatchSize: 8})
	if err != nil {
		t.Fatal(err)
	}
	ctx, finish := runtime.Telemetry.Start(context.Background(), ports.OperationHTTP)
	runtime.Observer.Record(ctx, ports.Event{Name: ports.EventHealthChecked})
	finish(nil)
	shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := runtime.Shutdown(shutdown); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "private-collector-secret") {
		t.Fatal("collector error body leaked")
	}
	failures := runtime.DroppedBatches()
	for _, signal := range []string{"traces", "metrics", "logs"} {
		if failures[signal] != 1 {
			t.Errorf("missing %s drop count: %d", signal, failures[signal])
		}
	}
}

func TestRuntimeRejectsUnsafeAmbientSDKSettingsBeforeParsing(t *testing.T) {
	for _, name := range []string{"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "OTEL_BLRP_MAX_QUEUE_SIZE", "OTEL_EXPORTER_OTLP_HEADERS"} {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			previous := log.Writer()
			log.SetOutput(&output)
			defer log.SetOutput(previous)
			t.Setenv(name, "private-secret-invalid-value")
			_, err := NewRuntime(context.Background(), config.TelemetryConfig{Enabled: true, Endpoint: "http://localhost:4318", ServiceName: "test", SampleRatio: 1, ExportTimeout: time.Second, BatchInterval: time.Second, MetricInterval: time.Second, QueueSize: 16, BatchSize: 8})
			if err == nil {
				t.Fatal("unsafe SDK environment accepted")
			}
			if strings.Contains(err.Error()+output.String(), "private-secret") {
				t.Fatal("ambient configuration leaked")
			}
		})
	}
}

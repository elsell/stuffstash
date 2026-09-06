package observability

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
	"go.opentelemetry.io/otel/codes"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTelemetryCorrelatesSafeEventsAndOperationMetrics(t *testing.T) {
	traces := tracetest.NewInMemoryExporter()
	tracer := sdktrace.NewTracerProvider(sdktrace.WithSyncer(traces))
	t.Cleanup(func() { _ = tracer.Shutdown(context.Background()) })
	reader := sdkmetric.NewManualReader()
	meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = meter.Shutdown(context.Background()) })
	logs := &recordingLogExporter{}
	logger := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewSimpleProcessor(logs)))
	t.Cleanup(func() { _ = logger.Shutdown(context.Background()) })
	telemetry, err := NewTelemetry(tracer, meter, logger)
	if err != nil {
		t.Fatal(err)
	}
	ctx, finish := telemetry.Start(context.Background(), ports.OperationThumbnailGenerate)
	telemetry.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentThumbnailGenerated,
		Message: "private photo name and bearer credential",
		Fields:  map[string]string{"variant": "small", "source": "generated", "file_name": "private.jpg", "authorization": "Bearer secret", "tenant_id": "private-tenant"},
	})
	telemetry.Record(ctx, ports.Event{Name: ports.EventName("Bearer private-secret")})
	telemetry.Record(ctx, ports.Event{Name: ports.EventAttachmentThumbnailGenerated, Fields: map[string]string{"variant": "private-secret", "source": "private-secret"}})
	finish(errors.New("provider endpoint with secret credential"))
	spans := traces.GetSpans()
	if len(spans) != 1 || spans[0].Name != string(ports.OperationThumbnailGenerate) {
		t.Fatalf("unexpected spans: %d", len(spans))
	}
	if len(logs.records) != 2 || logs.records[0].TraceID() != spans[0].SpanContext.TraceID() {
		t.Fatal("event must correlate with its operation trace")
	}
	if logs.records[1].AttributesLen() != 0 {
		t.Fatal("unsafe values passed allowlist")
	}
	if logs.records[0].SpanID() != spans[0].SpanContext.SpanID() {
		t.Fatal("span correlation missing")
	}
	if spans[0].Status.Code != codes.Error {
		t.Fatal("error status missing")
	}
	record := logs.records[0]
	if strings.Contains(record.Body().AsString(), "private") {
		t.Fatal("event message leaked")
	}
	if record.AttributesLen() != 2 {
		t.Fatalf("only variant and source may be exported, got %d attributes", record.AttributesLen())
	}
	if spans[0].Status.Description != "" {
		t.Fatal("raw error description must not be exported")
	}
	var data metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &data); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, scope := range data.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name != "stuffstash.operation.duration" {
				continue
			}
			histogram, ok := metric.Data.(metricdata.Histogram[float64])
			if !ok {
				t.Fatal("operation duration must be a histogram")
			}
			if len(histogram.DataPoints) != 1 || histogram.DataPoints[0].Count != 1 {
				t.Fatal("operation completion must record one duration")
			}
			if len(histogram.DataPoints[0].Bounds) < 3 || histogram.DataPoints[0].Bounds[0] != 0.001 {
				t.Fatal("missing subsecond latency buckets")
			}
			found = true
		}
	}
	if !found {
		t.Fatal("missing duration metric")
	}
}

func TestTelemetryCompletionIsIdempotent(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tracer := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	reader := sdkmetric.NewManualReader()
	meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	logger := sdklog.NewLoggerProvider()
	t.Cleanup(func() {
		_ = tracer.Shutdown(context.Background())
		_ = meter.Shutdown(context.Background())
		_ = logger.Shutdown(context.Background())
	})
	telemetry, err := NewTelemetry(tracer, meter, logger)
	if err != nil {
		t.Fatal(err)
	}
	_, finish := telemetry.Start(context.Background(), ports.OperationThumbnailGenerate)
	finish(nil)
	finish(errors.New("late failure"))
	if len(exporter.GetSpans()) != 1 {
		t.Fatal("completion must finish exactly once")
	}
	var data metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &data); err != nil {
		t.Fatal(err)
	}
	for _, scope := range data.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if h, ok := metric.Data.(metricdata.Histogram[float64]); ok {
				for _, p := range h.DataPoints {
					if p.Count != 1 {
						t.Fatal("completion recorded twice")
					}
				}
			}
		}
	}
}

type recordingLogExporter struct{ records []sdklog.Record }

func (e *recordingLogExporter) Export(_ context.Context, records []sdklog.Record) error {
	for _, record := range records {
		e.records = append(e.records, record.Clone())
	}
	return nil
}
func (*recordingLogExporter) Shutdown(context.Context) error   { return nil }
func (*recordingLogExporter) ForceFlush(context.Context) error { return nil }

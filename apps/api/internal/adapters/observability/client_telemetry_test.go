package observability

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/ports"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"testing"
)

func TestClientMeasurementsExportDurationAndSafeLogs(t *testing.T) {
	ctx := context.Background()
	reader := sdkmetric.NewManualReader()
	meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	tracer := sdktrace.NewTracerProvider()
	logs := &recordingLogExporter{}
	logger := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewSimpleProcessor(logs)))
	defer meter.Shutdown(ctx)
	defer tracer.Shutdown(ctx)
	defer logger.Shutdown(ctx)
	telemetry, err := NewTelemetry(tracer, meter, logger)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]string{"platform": "ios", "operation": "image", "surface": "gallery", "variant": "medium", "outcome": "success", "duration_ms": "125.5", "url": "private-url", "tenant_id": "private-tenant"}
	telemetry.Record(ctx, ports.Event{Name: ports.EventClientPerformanceObserved, Message: "private-message", Fields: fields})
	fields["duration_ms"] = "NaN"
	telemetry.Record(ctx, ports.Event{Name: ports.EventClientPerformanceObserved, Fields: fields})
	if len(logs.records) != 1 || logs.records[0].AttributesLen() != 6 {
		t.Fatal("client log validation or privacy failed")
	}
	var result metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &result); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, scope := range result.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != "stuffstash.client.duration" {
				continue
			}
			found = true
			data, ok := m.Data.(metricdata.Histogram[float64])
			if !ok || len(data.DataPoints) != 1 {
				t.Fatal("client duration missing")
			}
			point := data.DataPoints[0]
			if point.Count != 1 || point.Sum != 0.1255 || point.Attributes.Len() != 5 || m.Unit != "s" {
				t.Fatal("client duration or dimensions changed")
			}
		}
	}
	if !found {
		t.Fatal("client metric missing")
	}
}

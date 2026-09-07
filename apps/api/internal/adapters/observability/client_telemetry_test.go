package observability

import (
	"context"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/ports"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"strings"
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
	for _, invalid := range []string{"NaN", "+Inf", "60001", "-1"} {
		fields["duration_ms"] = invalid
		telemetry.Record(ctx, ports.Event{Name: ports.EventClientPerformanceObserved, Fields: fields})
	}
	fields["duration_ms"] = "125.5"
	fields["platform"] = "private-platform"
	telemetry.Record(ctx, ports.Event{Name: ports.EventClientPerformanceObserved, Fields: fields})
	if len(logs.records) != 1 || logs.records[0].AttributesLen() != 6 {
		t.Fatal("client log validation or privacy failed")
	}
	expected := map[string]string{"platform": "ios", "operation": "image", "surface": "gallery", "variant": "medium", "outcome": "success"}
	logs.records[0].WalkAttributes(func(value otellog.KeyValue) bool {
		if value.Key == "duration_ms" {
			if value.Value.AsFloat64() != 125.5 {
				t.Fatal("log duration changed")
			}
			return true
		}
		if expected[value.Key] != value.Value.AsString() || strings.Contains(fmt.Sprint(value), "private") {
			t.Fatal("unsafe client log attributes")
		}
		return true
	})
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
			for _, attr := range point.Attributes.ToSlice() {
				if expected[string(attr.Key)] != attr.Value.AsString() || strings.Contains(fmt.Sprint(attr), "private") {
					t.Fatal("unsafe client metric attributes")
				}
			}
			if point.Count != 1 || point.Sum != 0.1255 || point.Attributes.Len() != 5 || m.Unit != "s" {
				t.Fatal("client duration or dimensions changed")
			}
		}
	}
	if !found {
		t.Fatal("client metric missing")
	}
}

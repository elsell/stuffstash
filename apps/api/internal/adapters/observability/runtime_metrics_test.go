package observability

import (
	"context"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestRuntimeMetricsCollectRealUnlabelledResources(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer provider.Shutdown(context.Background())
	if err := registerRuntimeMetrics(provider.Meter(instrumentationName)); err != nil {
		t.Fatal(err)
	}
	var result metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &result); err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{
		"stuffstash.go.heap.objects": "By", "stuffstash.go.goroutines": "{goroutine}",
		"stuffstash.go.heap.allocated": "By", "stuffstash.go.gc.cycles": "{cycle}",
	}
	for _, scope := range result.ScopeMetrics {
		for _, value := range scope.Metrics {
			unit, ok := expected[value.Name]
			if !ok || unit != value.Unit {
				t.Fatalf("unexpected runtime metric %s (%s)", value.Name, value.Unit)
			}
			var points []metricdata.DataPoint[int64]
			switch data := value.Data.(type) {
			case metricdata.Gauge[int64]:
				if value.Name != "stuffstash.go.heap.objects" && value.Name != "stuffstash.go.goroutines" {
					t.Fatal("counter emitted as gauge")
				}
				points = data.DataPoints
			case metricdata.Sum[int64]:
				if !data.IsMonotonic {
					t.Fatal("runtime counter not monotonic")
				}
				points = data.DataPoints
			default:
				t.Fatal("unexpected runtime aggregation")
			}
			if len(points) != 1 || points[0].Attributes.Len() != 0 || points[0].Value < 0 {
				t.Fatal("missing, labelled or invalid runtime value")
			}
			if value.Name != "stuffstash.go.gc.cycles" && points[0].Value == 0 {
				t.Fatal("runtime resource value is empty")
			}
			delete(expected, value.Name)
		}
	}
	if len(expected) != 0 {
		t.Fatalf("missing runtime metrics: %v", expected)
	}
}

package observability

import (
	"context"
	"errors"
	"math"
	runtimemetrics "runtime/metrics"

	"go.opentelemetry.io/otel/metric"
)

func registerRuntimeMetrics(meter metric.Meter) error {
	available := make(map[string]runtimemetrics.ValueKind)
	for _, description := range runtimemetrics.All() {
		available[description.Name] = description.Kind
	}
	for _, definition := range []struct {
		name, source, unit, description string
		cumulative                      bool
	}{
		{"stuffstash.go.heap.objects", "/memory/classes/heap/objects:bytes", "By", "Bytes occupied by live or not-yet-swept Go heap objects", false},
		{"stuffstash.go.goroutines", "/sched/goroutines:goroutines", "{goroutine}", "Live Go goroutines", false},
		{"stuffstash.go.heap.allocated", "/gc/heap/allocs:bytes", "By", "Cumulative bytes allocated on the Go heap", true},
		{"stuffstash.go.gc.cycles", "/gc/cycles/total:gc-cycles", "{cycle}", "Completed Go garbage collection cycles", true},
	} {
		if available[definition.source] != runtimemetrics.KindUint64 {
			return errors.New("required runtime metric unavailable")
		}
		callback := metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			// Each collection owns its sample buffer; concurrent readers cannot race.
			samples := []runtimemetrics.Sample{{Name: definition.source}}
			runtimemetrics.Read(samples)
			if samples[0].Value.Kind() != runtimemetrics.KindUint64 {
				return errors.New("runtime metric unavailable")
			}
			value := samples[0].Value.Uint64()
			if value > math.MaxInt64 {
				value = math.MaxInt64
			}
			observer.Observe(int64(value))
			return nil
		})
		var err error
		if definition.cumulative {
			_, err = meter.Int64ObservableCounter(definition.name, metric.WithUnit(definition.unit), metric.WithDescription(definition.description), callback)
		} else {
			_, err = meter.Int64ObservableGauge(definition.name, metric.WithUnit(definition.unit), metric.WithDescription(definition.description), callback)
		}
		if err != nil {
			return errors.New("runtime metric initialization failed")
		}
	}
	return nil
}

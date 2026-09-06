package clienttelemetry

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"math"
	"testing"
)

func TestBatchValidationIsAtomicAndBounded(t *testing.T) {
	good := ports.ClientMeasurement{Platform: ports.ClientPlatformWeb, Operation: ports.ClientOperationImage, Surface: ports.ClientSurfaceGallery, Variant: ports.ClientVariantMedium, Outcome: ports.ClientOutcomeSuccess, DurationMS: 125.5}
	for _, batch := range [][]ports.ClientMeasurement{nil, make([]ports.ClientMeasurement, 51), {good, {Platform: ports.ClientPlatform("private")}}, {good, withDuration(good, math.NaN())}, {good, withDuration(good, math.Inf(1))}, {good, withDuration(good, 60001)}} {
		observer := &events{}
		if err := Record(context.Background(), observer, batch); err == nil {
			t.Fatal("invalid batch accepted")
		}
		if len(observer.values) != 0 {
			t.Fatal("invalid batch partially emitted")
		}
	}
	observer := &events{}
	if err := Record(context.Background(), observer, []ports.ClientMeasurement{good}); err != nil {
		t.Fatal(err)
	}
	if len(observer.values) != 1 || observer.values[0].Fields["duration_ms"] != "125.5" {
		t.Fatal("valid measurement lost")
	}
}
func withDuration(value ports.ClientMeasurement, duration float64) ports.ClientMeasurement {
	value.DurationMS = duration
	return value
}

type events struct{ values []ports.Event }

func (e *events) Record(_ context.Context, value ports.Event) { e.values = append(e.values, value) }

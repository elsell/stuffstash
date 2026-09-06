package clienttelemetry

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strconv"
)

func Record(ctx context.Context, observer ports.Observer, values []ports.ClientMeasurement) error {
	if len(values) == 0 || len(values) > 50 {
		return apperrors.ErrInvalidInput
	}
	for _, value := range values {
		if !value.Valid() {
			return apperrors.ErrInvalidInput
		}
	}
	for _, value := range values {
		observer.Record(ctx, ports.Event{Name: ports.EventClientPerformanceObserved, Fields: map[string]string{
			"platform": string(value.Platform), "operation": string(value.Operation), "surface": string(value.Surface), "variant": string(value.Variant), "outcome": string(value.Outcome), "duration_ms": strconv.FormatFloat(value.DurationMS, 'f', -1, 64),
		}})
	}
	return nil
}

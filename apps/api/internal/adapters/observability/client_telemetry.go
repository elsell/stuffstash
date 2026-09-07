package observability

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"strconv"
	"time"
)

func (t *Telemetry) recordClientMeasurement(ctx context.Context, fields map[string]string) {
	duration, err := strconv.ParseFloat(fields["duration_ms"], 64)
	if err != nil {
		return
	}
	value := ports.ClientMeasurement{Platform: ports.ClientPlatform(fields["platform"]), Operation: ports.ClientOperation(fields["operation"]), Surface: ports.ClientSurface(fields["surface"]), Variant: ports.ClientVariant(fields["variant"]), Outcome: ports.ClientOutcome(fields["outcome"]), DurationMS: duration}
	if !value.Valid() {
		return
	}
	// Rebuild attributes from typed measurements; never forward an arbitrary map.
	attrs := []attribute.KeyValue{attribute.String("platform", string(value.Platform)), attribute.String("operation", string(value.Operation)), attribute.String("surface", string(value.Surface)), attribute.String("variant", string(value.Variant)), attribute.String("outcome", string(value.Outcome))}
	t.clientDuration.Record(ctx, duration/1000, metric.WithAttributes(attrs...))
	var record otellog.Record
	record.SetTimestamp(time.Now())
	record.SetBody(otellog.StringValue(string(ports.EventClientPerformanceObserved)))
	record.SetSeverity(otellog.SeverityInfo)
	for _, attr := range attrs {
		record.AddAttributes(otellog.String(string(attr.Key), attr.Value.AsString()))
	}
	record.AddAttributes(otellog.Float64("duration_ms", duration))
	t.logger.Emit(ctx, record)
}

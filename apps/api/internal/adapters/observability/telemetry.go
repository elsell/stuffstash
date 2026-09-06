package observability

import (
	"context"
	"sync"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/stuffstash/stuff-stash"

type Telemetry struct {
	tracer   trace.Tracer
	logger   otellog.Logger
	duration metric.Float64Histogram
}

func NewTelemetry(tracer trace.TracerProvider, meter metric.MeterProvider, logger otellog.LoggerProvider) (*Telemetry, error) {
	duration, err := meter.Meter(instrumentationName).Float64Histogram("stuffstash.operation.duration", metric.WithUnit("s"), metric.WithExplicitBucketBoundaries(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30), metric.WithDescription("Duration of completed inventory operations"))
	if err != nil {
		return nil, err
	}
	return &Telemetry{tracer: tracer.Tracer(instrumentationName), logger: logger.Logger(instrumentationName), duration: duration}, nil
}

func (t *Telemetry) Start(ctx context.Context, operation ports.Operation) (context.Context, func(error)) {
	operation = operation.Bounded()
	ctx, span := t.tracer.Start(ctx, string(operation))
	started := time.Now()
	var once sync.Once
	return ctx, func(err error) {
		once.Do(func() {
			outcome := "success"
			if err != nil {
				outcome = "failure"
				span.SetStatus(codes.Error, "")
			}
			attrs := []attribute.KeyValue{attribute.String("operation", string(operation)), attribute.String("outcome", outcome)}
			span.SetAttributes(attrs...)
			t.duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
			span.End()
		})
	}
}

func (t *Telemetry) Record(ctx context.Context, event ports.Event) {
	if !event.Name.Known() {
		return
	}
	var record otellog.Record
	record.SetTimestamp(time.Now())
	record.SetBody(otellog.StringValue(string(event.Name)))
	record.SetSeverity(otellog.SeverityInfo)
	// Field names AND values are allowlisted so a user-derived string cannot leak
	// through an otherwise safe field or create unbounded values.
	for key, value := range event.Fields {
		if safeTelemetryField(key, value) {
			record.AddAttributes(otellog.String(key, value))
		}
	}
	t.logger.Emit(ctx, record)
}

func safeTelemetryField(key, value string) bool {
	switch key {
	case "variant":
		return value == "small" || value == "medium" || value == "large"
	case "source":
		return value == "cache" || value == "generated"
	case "content_type":
		return value == "image/jpeg" || value == "image/png" || value == "image/webp" || value == "application/pdf"
	default:
		return false
	}
}

var _ ports.Telemetry = (*Telemetry)(nil)
var _ ports.Observer = (*Telemetry)(nil)

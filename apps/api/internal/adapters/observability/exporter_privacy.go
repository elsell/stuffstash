package observability

import (
	"context"
	"errors"
	"sync/atomic"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Terminal failures are dropped only after the transport's bounded retries. Do
// not return collector responses to SDK processors: they use a global logger.
type exportFailures struct{ traces, metrics, logs atomic.Int64 }

func (r *Runtime) DroppedBatches() map[string]int64 {
	return map[string]int64{"traces": r.failures.traces.Load(), "metrics": r.failures.metrics.Load(), "logs": r.failures.logs.Load()}
}
func safeExporterError(err error) error {
	if err != nil {
		return errors.New("telemetry exporter failed")
	}
	return nil
}

type privateTraceExporter struct {
	delegate sdktrace.SpanExporter
	dropped  *atomic.Int64
}

func (e privateTraceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if err := e.delegate.ExportSpans(ctx, spans); err != nil {
		e.dropped.Add(1)
	}
	return nil
}
func (e privateTraceExporter) Shutdown(ctx context.Context) error {
	return safeExporterError(e.delegate.Shutdown(ctx))
}

type privateLogExporter struct {
	delegate sdklog.Exporter
	dropped  *atomic.Int64
}

func (e privateLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	if err := e.delegate.Export(ctx, records); err != nil {
		e.dropped.Add(1)
	}
	return nil
}
func (e privateLogExporter) ForceFlush(ctx context.Context) error {
	return safeExporterError(e.delegate.ForceFlush(ctx))
}
func (e privateLogExporter) Shutdown(ctx context.Context) error {
	return safeExporterError(e.delegate.Shutdown(ctx))
}

type privateMetricExporter struct {
	delegate sdkmetric.Exporter
	dropped  *atomic.Int64
}

func (e privateMetricExporter) Temporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	return e.delegate.Temporality(kind)
}
func (e privateMetricExporter) Aggregation(kind sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return e.delegate.Aggregation(kind)
}
func (e privateMetricExporter) Export(ctx context.Context, data *metricdata.ResourceMetrics) error {
	if err := e.delegate.Export(ctx, data); err != nil {
		e.dropped.Add(1)
	}
	return nil
}
func (e privateMetricExporter) ForceFlush(ctx context.Context) error {
	return safeExporterError(e.delegate.ForceFlush(ctx))
}
func (e privateMetricExporter) Shutdown(ctx context.Context) error {
	return safeExporterError(e.delegate.Shutdown(ctx))
}

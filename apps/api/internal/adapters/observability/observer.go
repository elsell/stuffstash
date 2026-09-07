package observability

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"log/slog"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

const (
	EventHTTPServerStartFailed    ports.EventName = "http.server.start_failed"
	EventHTTPServerShutdownFailed ports.EventName = "http.server.shutdown_failed"
)

type Event = ports.Event

type FanOut struct {
	observers []ports.Observer
}

func NewFanOut(observers ...ports.Observer) FanOut {
	return FanOut{observers: observers}
}

func (f FanOut) Record(ctx context.Context, event ports.Event) {
	for _, observer := range f.observers {
		observer.Record(ctx, event)
	}
}

type SlogObserver struct {
	logger *slog.Logger
}

func NewSlogObserver(logger *slog.Logger) SlogObserver {
	return SlogObserver{logger: logger}
}

func (s SlogObserver) Record(ctx context.Context, event ports.Event) {
	attrs := []any{
		slog.String("event", string(event.Name)),
	}
	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		attrs = append(attrs, slog.String("trace_id", span.TraceID().String()), slog.String("span_id", span.SpanID().String()))
	}
	if event.Message != "" {
		attrs = append(attrs, slog.String("message", event.Message))
	}
	for name, value := range event.Fields {
		attrs = append(attrs, slog.String(name, value))
	}
	s.logger.Info("domain event", attrs...)
}

package observability

import (
	"context"
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

func (s SlogObserver) Record(_ context.Context, event ports.Event) {
	attrs := []any{
		slog.String("event", string(event.Name)),
	}
	if event.Message != "" {
		attrs = append(attrs, slog.String("message", event.Message))
	}
	for name, value := range event.Fields {
		attrs = append(attrs, slog.String(name, value))
	}
	s.logger.Info("domain event", attrs...)
}

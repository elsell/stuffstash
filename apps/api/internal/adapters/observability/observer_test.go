package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
	"go.opentelemetry.io/otel/trace"
)

func TestConsoleEventsIncludeTraceCorrelation(t *testing.T) {
	var output bytes.Buffer
	observer := NewSlogObserver(slog.New(slog.NewJSONHandler(&output, nil)))
	traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID}))
	observer.Record(ctx, ports.Event{Name: ports.EventHealthChecked})
	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry["trace_id"] != traceID.String() || entry["span_id"] != spanID.String() {
		t.Fatal("console log lost trace correlation")
	}
}

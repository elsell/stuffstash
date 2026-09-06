package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestHTTPTracePreservesParentAndHidesResourcePaths(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tracer := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	meter := sdkmetric.NewMeterProvider()
	logger := sdklog.NewLoggerProvider()
	t.Cleanup(func() {
		_ = tracer.Shutdown(context.Background())
		_ = meter.Shutdown(context.Background())
		_ = logger.Shutdown(context.Background())
	})
	telemetry, err := NewTelemetry(tracer, meter, logger)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /assets/{id}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) })
	request := httptest.NewRequest(http.MethodGet, "/assets/private-id?secret=private-query", nil)
	request.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	recorder := httptest.NewRecorder()
	telemetry.WrapHTTP(mux).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatal("status changed")
	}
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatal("missing request span")
	}
	if spans[0].SpanContext.TraceID().String() != "4bf92f3577b34da6a3ce929d0e0e4736" || spans[0].Parent.SpanID().String() != "00f067aa0ba902b7" {
		t.Fatal("parent context lost")
	}
	routeFound := false
	for _, attr := range spans[0].Attributes {
		if attr.Key == "http.route" && attr.Value.AsString() == "/assets/{id}" {
			routeFound = true
		}
		if attr.Value.AsString() == "private-id" || attr.Value.AsString() == "private-query" {
			t.Fatal("private path leaked")
		}
	}
	if !routeFound {
		t.Fatal("route template missing")
	}
}

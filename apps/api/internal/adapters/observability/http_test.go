package observability

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"nhooyr.io/websocket"
	"runtime"
	"strings"
	"testing"
	"time"

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
		if strings.Contains(attr.Value.AsString(), "private-id") || strings.Contains(attr.Value.AsString(), "private-query") {
			t.Fatal("private path leaked")
		}
	}
	if !routeFound {
		t.Fatal("route template missing")
	}
}

func TestHTTPTraceRecordsActualStatusWhenHandlerPanics(t *testing.T) {
	for _, committed := range []bool{false, true} {
		t.Run(fmt.Sprint(committed), func(t *testing.T) {
			telemetry, exporter := httpTestTelemetry(t)
			handler := telemetry.WrapHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if committed {
					w.WriteHeader(200)
				}
				panic("controlled interruption")
			}))
			func() {
				defer func() { _ = recover() }()
				handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
			}()
			spans := exporter.GetSpans()
			if len(spans) != 1 {
				t.Fatal("missing interrupted span")
			}
			status := 0
			outcome := ""
			for _, attr := range spans[0].Attributes {
				if attr.Key == "http.response.status_code" {
					status = int(attr.Value.AsInt64())
				}
				if attr.Key == "outcome" {
					outcome = attr.Value.AsString()
				}
			}
			expected := 0
			if committed {
				expected = 200
			}
			if status != expected || outcome != "interrupted" {
				t.Fatalf("status=%d outcome=%s", status, outcome)
			}
		})
	}
}

func TestHTTPTracePreservesWebSocketUpgrade(t *testing.T) {
	telemetry, exporter := httpTestTelemetry(t)
	finished := make(chan struct{})
	server := httptest.NewServer(telemetry.WrapHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(finished)
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		_ = conn.Close(websocket.StatusNormalClosure, "")
	})))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.Close(websocket.StatusNormalClosure, "")
	<-finished
	if response.StatusCode != 101 {
		t.Fatal("upgrade failed")
	}
	// Server handler completion precedes the instrumentation defer; shutdown flushes
	// only after the recorded span arrives through this bounded wait.
	deadline := time.After(time.Second)
	for len(exporter.GetSpans()) == 0 {
		select {
		case <-deadline:
			t.Fatal("missing upgrade span")
		default:
			runtime.Gosched()
		}
	}
	for _, attr := range exporter.GetSpans()[0].Attributes {
		if attr.Key == "http.response.status_code" && attr.Value.AsInt64() == 101 {
			return
		}
	}
	t.Fatal("upgrade recorded as a different status")
}

func httpTestTelemetry(t *testing.T) (*Telemetry, *tracetest.InMemoryExporter) {
	t.Helper()
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
	return telemetry, exporter
}

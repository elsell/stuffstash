package observability

import (
	"net/http"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func (t *Telemetry) WrapHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := propagation.TraceContext{}.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := t.tracer.Start(ctx, string(ports.OperationHTTP), trace.WithSpanKind(trace.SpanKindServer))
		request := r.WithContext(ctx)
		started := time.Now()
		captured := httpsnoop.Metrics{Code: http.StatusOK}
		completed := false
		defer func() {
			if !completed {
				captured.Code = http.StatusInternalServerError
			}
			method := boundedHTTPMethod(r.Method)
			route := request.Pattern
			if _, path, ok := strings.Cut(route, " "); ok {
				route = path
			}
			if route == "" {
				route = "/unmatched"
			}
			outcome := "success"
			if captured.Code >= 400 {
				outcome = "failure"
				span.SetStatus(codes.Error, "")
			}
			attrs := []attribute.KeyValue{attribute.String("operation", string(ports.OperationHTTP)), attribute.String("outcome", outcome), attribute.String("http.request.method", method), attribute.String("http.route", route), attribute.Int("http.response.status_code", captured.Code)}
			span.SetName(method + " " + route)
			span.SetAttributes(attrs...)
			t.duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
			span.End()
		}()
		captured.CaptureMetrics(w, func(writer http.ResponseWriter) { next.ServeHTTP(writer, request) })
		completed = true
	})
}

func (r *Runtime) WrapHTTP(next http.Handler) http.Handler {
	if telemetry, ok := r.Telemetry.(*Telemetry); ok {
		return telemetry.WrapHTTP(next)
	}
	return next
}

func boundedHTTPMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodConnect, http.MethodTrace:
		return method
	default:
		return "OTHER"
	}
}

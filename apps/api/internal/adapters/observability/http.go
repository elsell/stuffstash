package observability

import (
	"bufio"
	"io"
	"net"
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
		status := 0
		hijacked := false
		completed := false
		defer func() {
			if completed && status == 0 && !hijacked {
				status = http.StatusOK
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
			if status >= 400 {
				outcome = "failure"
				span.SetStatus(codes.Error, "")
			}
			if !completed {
				outcome = "interrupted"
				span.SetStatus(codes.Error, "")
			}
			attrs := []attribute.KeyValue{attribute.String("operation", string(ports.OperationHTTP)), attribute.String("outcome", outcome), attribute.String("http.request.method", method), attribute.String("http.route", route)}
			if status != 0 {
				attrs = append(attrs, attribute.Int("http.response.status_code", status))
			}
			span.SetName(method + " " + route)
			span.SetAttributes(attrs...)
			t.duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
			span.End()
		}()
		implicitStatus := func() {
			if status == 0 {
				status = http.StatusOK
			}
		}
		writer := httpsnoop.Wrap(w, httpsnoop.Hooks{
			WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					next(code)
					if status == 0 && (code >= 200 || code == 101) {
						status = code
					}
				}
			},
			Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return func(p []byte) (int, error) { implicitStatus(); return next(p) }
			},
			ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
				return func(r io.Reader) (int64, error) { implicitStatus(); return next(r) }
			},
			Flush: func(next httpsnoop.FlushFunc) httpsnoop.FlushFunc { return func() { implicitStatus(); next() } },
			Hijack: func(next httpsnoop.HijackFunc) httpsnoop.HijackFunc {
				return func() (net.Conn, *bufio.ReadWriter, error) {
					conn, buffer, err := next()
					if err == nil {
						hijacked = true
					}
					return conn, buffer, err
				}
			},
		})
		next.ServeHTTP(writer, request)
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

package observability

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Runtime struct {
	Telemetry   ports.Telemetry
	Observer    ports.Observer
	shutdown    []func(context.Context) error
	once        sync.Once
	shutdownErr error
}

func NewRuntime(ctx context.Context, cfg config.TelemetryConfig) (*Runtime, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	runtime := &Runtime{Telemetry: ports.NoopTelemetry{}, Observer: discardObserver{}}
	if !cfg.Enabled {
		return runtime, nil
	}
	res := resource.NewSchemaless(attribute.String("service.name", cfg.ServiceName), attribute.String("service.version", cfg.ServiceVersion), attribute.String("deployment.environment.name", cfg.Environment))
	base := strings.TrimRight(cfg.Endpoint, "/")
	traceExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(base+"/v1/traces"), otlptracehttp.WithHeaders(cfg.Headers), otlptracehttp.WithTimeout(cfg.ExportTimeout))
	if err != nil {
		return nil, errors.New("trace exporter initialization failed")
	}
	tracer := sdktrace.NewTracerProvider(sdktrace.WithResource(res), sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))), sdktrace.WithBatcher(traceExporter, sdktrace.WithMaxQueueSize(cfg.QueueSize), sdktrace.WithMaxExportBatchSize(cfg.BatchSize), sdktrace.WithBatchTimeout(cfg.BatchInterval), sdktrace.WithExportTimeout(cfg.ExportTimeout)))
	runtime.shutdown = append(runtime.shutdown, tracer.Shutdown)
	metricExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(base+"/v1/metrics"), otlpmetrichttp.WithHeaders(cfg.Headers), otlpmetrichttp.WithTimeout(cfg.ExportTimeout))
	if err != nil {
		_ = runtime.Shutdown(ctx)
		return nil, errors.New("metric exporter initialization failed")
	}
	meter := sdkmetric.NewMeterProvider(sdkmetric.WithResource(res), sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(cfg.MetricInterval), sdkmetric.WithTimeout(cfg.ExportTimeout))))
	runtime.shutdown = append(runtime.shutdown, meter.Shutdown)
	logExporter, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(base+"/v1/logs"), otlploghttp.WithHeaders(cfg.Headers), otlploghttp.WithTimeout(cfg.ExportTimeout))
	if err != nil {
		_ = runtime.Shutdown(ctx)
		return nil, errors.New("log exporter initialization failed")
	}
	logger := sdklog.NewLoggerProvider(sdklog.WithResource(res), sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter, sdklog.WithMaxQueueSize(cfg.QueueSize), sdklog.WithExportMaxBatchSize(cfg.BatchSize), sdklog.WithExportInterval(cfg.BatchInterval), sdklog.WithExportTimeout(cfg.ExportTimeout))))
	runtime.shutdown = append(runtime.shutdown, logger.Shutdown)
	telemetry, err := NewTelemetry(tracer, meter, logger)
	if err != nil {
		_ = runtime.Shutdown(ctx)
		return nil, errors.New("telemetry instrumentation initialization failed")
	}
	runtime.Telemetry = telemetry
	runtime.Observer = telemetry
	return runtime, nil
}

func (r *Runtime) Shutdown(ctx context.Context) error {
	r.once.Do(func() {
		for _, shutdown := range r.shutdown {
			if err := shutdown(ctx); err != nil {
				r.shutdownErr = errors.New("telemetry shutdown failed")
			}
		}
	})
	return r.shutdownErr
}

type discardObserver struct{}

func (discardObserver) Record(context.Context, ports.Event) {}

package httpserver

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/observability"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTelemetryPreservesInventorySecurityBoundary(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const otherTenant = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "owner"}, {id: otherTenant, name: "Other", owner: "other"}},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Home", owner: "owner"}},
	}))
	exporter := tracetest.NewInMemoryExporter()
	tracer := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	meter := sdkmetric.NewMeterProvider()
	logger := sdklog.NewLoggerProvider()
	t.Cleanup(func() {
		_ = tracer.Shutdown(context.Background())
		_ = meter.Shutdown(context.Background())
		_ = logger.Shutdown(context.Background())
	})
	telemetry, err := observability.NewTelemetry(tracer, meter, logger)
	if err != nil {
		t.Fatal(err)
	}
	server.Handler = telemetry.WrapHTTP(server.Handler)
	path := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets"
	for _, test := range []struct {
		name, token, path string
		status            int
	}{
		{"owner", "Bearer dev:owner", path, 200},
		{"anonymous", "", path, 401},
		{"malformed", "Bearer private-invalid-token", path, 401},
		{"other principal", "Bearer dev:other", path, 403},
		{"wrong tenant", "Bearer dev:owner", strings.Replace(path, tenantID, otherTenant, 1), 403},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := performRequest(server, http.MethodGet, test.path, test.token, nil)
			if response.Code != test.status {
				t.Fatalf("status %d, expected %d", response.Code, test.status)
			}
		})
	}
	for _, span := range exporter.GetSpans() {
		if strings.Contains(span.Name, tenantID) {
			t.Fatal("tenant ID exposed in span name")
		}
		for _, attr := range span.Attributes {
			if strings.Contains(attr.Value.AsString(), tenantID) || strings.Contains(attr.Value.AsString(), "private-invalid-token") {
				t.Fatal("private request information exposed")
			}
		}
	}
}

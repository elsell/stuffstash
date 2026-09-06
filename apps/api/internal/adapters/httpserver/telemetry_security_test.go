package httpserver

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/homebox"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
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
	server := NewServer(":0", newSeededTestAppWithBlobAuthorizerAndImportSource(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "owner"}, {id: otherTenant, name: "Other", owner: "other"}},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Home", owner: "owner"}},
	}, nil, observability.ObserveAuthorizer(memory.NewAuthorizer(), telemetry), homebox.NewLegacyImporter(nil), observability.ObserveAuthenticator(auth.NewLocalDevAuthenticator(), telemetry)))
	exporter.Reset()
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
		{"wrong tenant", "Bearer dev:owner", strings.Replace(path, tenantID, otherTenant, 1), 404},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := performRequest(server, http.MethodGet, test.path, test.token, nil)
			if response.Code != test.status {
				t.Fatalf("status %d, expected %d", response.Code, test.status)
			}
		})
	}
	spans := exporter.GetSpans()
	requestCount, authenticationCount, authorizationCount := 0, 0, 0
	requestParents := map[string]string{}
	for _, span := range spans {
		if span.Name == "GET /tenants/{tenantId}/inventories/{inventoryId}/assets" {
			requestParents[span.SpanContext.SpanID().String()] = span.SpanContext.TraceID().String()
		}
	}
	for _, span := range spans {
		switch span.Name {
		case "GET /tenants/{tenantId}/inventories/{inventoryId}/assets":
			requestCount++
		case "identity.authenticate":
			authenticationCount++
		case "identity.authorize", "identity.visibility":
			authorizationCount++
		default:
			t.Fatalf("unexpected span %q", span.Name)
		}
		if strings.HasPrefix(span.Name, "identity.") && requestParents[span.Parent.SpanID().String()] != span.SpanContext.TraceID().String() {
			t.Fatal("identity span lost HTTP request parent")
		}
		if strings.Contains(span.Name, tenantID) {
			t.Fatal("tenant ID exposed in span name")
		}
		for _, attr := range span.Attributes {
			if strings.Contains(attr.Value.AsString(), tenantID) || strings.Contains(attr.Value.AsString(), inventoryID) || strings.Contains(attr.Value.AsString(), otherTenant) || strings.Contains(attr.Value.AsString(), "private-invalid-token") {
				t.Fatal("private request information exposed")
			}
		}
	}
	if requestCount != 5 || authenticationCount != 5 || authorizationCount == 0 {
		t.Fatalf("missing boundary spans: requests %d authentication %d authorization %d", requestCount, authenticationCount, authorizationCount)
	}
}

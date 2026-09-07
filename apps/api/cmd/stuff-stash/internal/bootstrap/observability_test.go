package bootstrap

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/observability"
	"github.com/stuffstash/stuff-stash/internal/config"
)

func TestRunValidatesTelemetryBeforeStartingServices(t *testing.T) {
	t.Setenv("STUFF_STASH_TELEMETRY_ENABLED", "true")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "file:///private-secret")
	err := Run(context.Background(), config.Config{}, observability.NewFanOut())
	if err == nil || err.Error() != "invalid telemetry endpoint" {
		t.Fatalf("expected safe telemetry validation, got %v", err)
	}
}

func TestRunValidatesProfilingBeforeStartingServices(t *testing.T) {
	t.Setenv("STUFF_STASH_TELEMETRY_ENABLED", "false")
	t.Setenv("STUFF_STASH_PROFILING_ENABLED", "true")
	t.Setenv("STUFF_STASH_PROFILING_ENDPOINT", "file:///private-secret")
	err := Run(context.Background(), config.Config{}, observability.NewFanOut())
	if err == nil || err.Error() != "invalid profiling endpoint" {
		t.Fatalf("expected safe profiling validation, got %v", err)
	}
}

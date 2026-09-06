package config

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestTelemetryDisabledNeedsNoExporterCredentials(t *testing.T) {
	t.Setenv("STUFF_STASH_TELEMETRY_ENABLED", "false")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "")
	cfg, err := LoadTelemetry()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled {
		t.Fatal("telemetry should remain disabled")
	}
}

func TestTelemetryParsesCloudConnectionWithoutExposingSecrets(t *testing.T) {
	t.Setenv("STUFF_STASH_TELEMETRY_ENABLED", "true")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://collector.example.test/otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Basic%20test-secret")
	t.Setenv("STUFF_STASH_TELEMETRY_SAMPLE_RATIO", "0.25")
	cfg, err := LoadTelemetry()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Headers["Authorization"] != "Basic test-secret" || cfg.SampleRatio != 0.25 {
		t.Fatal("connection settings not parsed")
	}
	if cfg.ExportTimeout != 5*time.Second {
		t.Fatal("unexpected timeout default")
	}
}

func TestTelemetryRejectsInvalidRuntimeConfiguration(t *testing.T) {
	for _, test := range []struct{ name, value string }{
		{"STUFF_STASH_TELEMETRY_ENABLED", "invalid-secret"},
		{"OTEL_EXPORTER_OTLP_ENDPOINT", "https://user:secret@collector.example.test"},
		{"OTEL_EXPORTER_OTLP_ENDPOINT", "file:///secret"},
		{"OTEL_EXPORTER_OTLP_HEADERS", "Authorization=secret,authorization=duplicate"},
		{"OTEL_EXPORTER_OTLP_HEADERS", "Authorization=secret%0D%0AInjected"},
		{"STUFF_STASH_TELEMETRY_SAMPLE_RATIO", "NaN"},
		{"STUFF_STASH_TELEMETRY_SAMPLE_RATIO", "1.1"},
		{"STUFF_STASH_TELEMETRY_EXPORT_TIMEOUT", "0s"},
		{"STUFF_STASH_TELEMETRY_QUEUE_SIZE", "0"},
		{"STUFF_STASH_TELEMETRY_BATCH_SIZE", "4096"},
	} {
		t.Run(test.name+"/"+test.value, func(t *testing.T) {
			t.Setenv("STUFF_STASH_TELEMETRY_ENABLED", "true")
			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://collector.example.test/otlp")
			t.Setenv(test.name, test.value)
			_, err := LoadTelemetry()
			if err == nil {
				t.Fatal("invalid configuration accepted")
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("configuration error leaked credential")
			}
		})
	}
}

func TestTelemetryValidationRejectsNonFiniteSampling(t *testing.T) {
	cfg := TelemetryConfig{Enabled: true, Endpoint: "https://collector.example.test", ServiceName: "test", SampleRatio: math.Inf(1), ExportTimeout: time.Second, BatchInterval: time.Second, MetricInterval: time.Second, QueueSize: 16, BatchSize: 8}
	if cfg.Validate() == nil {
		t.Fatal("infinite sampling accepted")
	}
}

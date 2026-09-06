package config

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type TelemetryConfig struct {
	Enabled        bool
	Endpoint       string
	Headers        map[string]string
	ServiceName    string
	ServiceVersion string
	Environment    string
	SampleRatio    float64
	ExportTimeout  time.Duration
	BatchInterval  time.Duration
	MetricInterval time.Duration
	QueueSize      int
	BatchSize      int
}

func LoadTelemetry() (TelemetryConfig, error) {
	cfg := TelemetryConfig{ServiceName: "stuffstash-api", SampleRatio: 0.1, ExportTimeout: 5 * time.Second, BatchInterval: 5 * time.Second, MetricInterval: 30 * time.Second, QueueSize: 2048, BatchSize: 256}
	if raw := os.Getenv("STUFF_STASH_TELEMETRY_ENABLED"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return cfg, errors.New("invalid telemetry enabled setting")
		}
		cfg.Enabled = value
	}
	if !cfg.Enabled {
		return cfg, nil
	}
	cfg.Endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if value := os.Getenv("OTEL_SERVICE_NAME"); value != "" {
		cfg.ServiceName = value
	}
	cfg.ServiceVersion = os.Getenv("OTEL_SERVICE_VERSION")
	cfg.Environment = os.Getenv("STUFF_STASH_DEPLOYMENT_ENVIRONMENT")
	var err error
	cfg.Headers, err = parseTelemetryHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"))
	if err != nil {
		return cfg, err
	}
	if raw := os.Getenv("STUFF_STASH_TELEMETRY_SAMPLE_RATIO"); raw != "" {
		cfg.SampleRatio, err = strconv.ParseFloat(raw, 64)
		if err != nil {
			return cfg, errors.New("invalid telemetry sample ratio")
		}
	}
	for name, target := range map[string]*time.Duration{
		"STUFF_STASH_TELEMETRY_EXPORT_TIMEOUT":  &cfg.ExportTimeout,
		"STUFF_STASH_TELEMETRY_BATCH_INTERVAL":  &cfg.BatchInterval,
		"STUFF_STASH_TELEMETRY_METRIC_INTERVAL": &cfg.MetricInterval,
	} {
		if raw := os.Getenv(name); raw != "" {
			*target, err = time.ParseDuration(raw)
			if err != nil {
				return cfg, fmt.Errorf("invalid %s", name)
			}
		}
	}
	for name, target := range map[string]*int{"STUFF_STASH_TELEMETRY_QUEUE_SIZE": &cfg.QueueSize, "STUFF_STASH_TELEMETRY_BATCH_SIZE": &cfg.BatchSize} {
		if raw := os.Getenv(name); raw != "" {
			*target, err = strconv.Atoi(raw)
			if err != nil {
				return cfg, fmt.Errorf("invalid %s", name)
			}
		}
	}
	return cfg, cfg.Validate()
}

func (c TelemetryConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if !validTelemetryEndpoint(c.Endpoint) {
		return errors.New("invalid telemetry endpoint")
	}
	if strings.TrimSpace(c.ServiceName) == "" || len(c.ServiceName) > 128 || len(c.ServiceVersion) > 128 || len(c.Environment) > 128 {
		return errors.New("invalid telemetry resource identity")
	}
	if math.IsNaN(c.SampleRatio) || math.IsInf(c.SampleRatio, 0) || c.SampleRatio < 0 || c.SampleRatio > 1 {
		return errors.New("invalid telemetry sample ratio")
	}
	if c.ExportTimeout <= 0 || c.BatchInterval <= 0 || c.MetricInterval <= 0 {
		return errors.New("invalid telemetry interval")
	}
	if c.QueueSize <= 0 || c.QueueSize > 65536 || c.BatchSize <= 0 || c.BatchSize > c.QueueSize {
		return errors.New("invalid telemetry queue settings")
	}
	for key, value := range c.Headers {
		if !validTelemetryHeader(key, value) {
			return errors.New("invalid telemetry header")
		}
	}
	return nil
}

func parseTelemetryHeaders(raw string) (map[string]string, error) {
	headers := map[string]string{}
	if raw == "" {
		return headers, nil
	}
	for _, entry := range strings.Split(raw, ",") {
		key, value, ok := strings.Cut(entry, "=")
		key = strings.TrimSpace(key)
		decoded, err := url.PathUnescape(strings.TrimSpace(value))
		if !ok || err != nil || !validTelemetryHeader(key, decoded) {
			return nil, errors.New("invalid telemetry headers")
		}
		key = http.CanonicalHeaderKey(key)
		if _, exists := headers[key]; exists {
			return nil, errors.New("duplicate telemetry header")
		}
		headers[key] = decoded
	}
	return headers, nil
}

func validTelemetryHeader(key, value string) bool {
	if key == "" || len(key) > 128 || len(value) > 8192 || strings.ContainsAny(value, "\r\n\x00") {
		return false
	}
	for _, c := range key {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", c)) {
			return false
		}
	}
	return true
}

package config

import (
	"errors"
	"net/url"
	"os"
	"strings"
)

// Exporter constructors inspect ambient SDK variables before applying explicit
// options. Reject unsupported settings before those parsers can log their values.
func ValidateTelemetryEnvironment() error {
	for _, entry := range os.Environ() {
		name, value, _ := strings.Cut(entry, "=")
		if value == "" || !strings.HasPrefix(name, "OTEL_") {
			continue
		}
		switch name {
		case "OTEL_SERVICE_NAME", "OTEL_SERVICE_VERSION":
		case "OTEL_EXPORTER_OTLP_ENDPOINT":
			if !validTelemetryEndpoint(value) {
				return errors.New("invalid ambient telemetry endpoint")
			}
		case "OTEL_EXPORTER_OTLP_HEADERS":
			if _, err := parseTelemetryHeaders(value); err != nil {
				return errors.New("invalid ambient telemetry headers")
			}
		default:
			return errors.New("unsupported telemetry SDK environment setting")
		}
	}
	return nil
}

func validTelemetryEndpoint(value string) bool {
	endpoint, err := url.Parse(value)
	return err == nil && endpoint.Host != "" && (endpoint.Scheme == "http" || endpoint.Scheme == "https") && endpoint.User == nil && endpoint.RawQuery == "" && endpoint.Fragment == ""
}

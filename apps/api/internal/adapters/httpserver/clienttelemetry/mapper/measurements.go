package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/clienttelemetry/dto"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func Measurements(values []dto.Measurement) []ports.ClientMeasurement {
	result := make([]ports.ClientMeasurement, len(values))
	for i, value := range values {
		result[i] = ports.ClientMeasurement{
			Platform:   ports.ClientPlatform(value.Platform),
			Operation:  ports.ClientOperation(value.Operation),
			Surface:    ports.ClientSurface(value.Surface),
			Variant:    ports.ClientVariant(value.Variant),
			Outcome:    ports.ClientOutcome(value.Outcome),
			DurationMS: value.DurationMS,
		}
	}
	return result
}

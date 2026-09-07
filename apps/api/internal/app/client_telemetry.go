package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/clienttelemetry"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) RecordClientTelemetry(ctx context.Context, values []ports.ClientMeasurement) error {
	return clienttelemetry.Record(ctx, a.observer, values)
}

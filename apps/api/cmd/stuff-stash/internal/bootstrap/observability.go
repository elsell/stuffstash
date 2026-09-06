package bootstrap

import (
	"context"
	"strconv"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/observability"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func startObservability(ctx context.Context, local ports.Observer) (*observability.Runtime, ports.Observer, func(), bool, error) {
	cfg, err := config.LoadTelemetry()
	if err != nil {
		return nil, nil, nil, false, err
	}
	runtime, err := observability.NewRuntime(ctx, cfg)
	if err != nil {
		return nil, nil, nil, false, err
	}
	combined := observability.NewFanOut(local, runtime.Observer)
	stop := func() {
		timeout := cfg.ExportTimeout + time.Second
		shutdown, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := runtime.Shutdown(shutdown); err != nil {
			local.Record(context.Background(), ports.Event{Name: ports.EventApplicationShutdownFailed, Message: "telemetry shutdown failed"})
		}
		for signal, count := range runtime.DroppedBatches() {
			if count > 0 {
				local.Record(context.Background(), ports.Event{Name: ports.EventTelemetryBatchesDropped, Message: "telemetry delivery incomplete", Fields: map[string]string{"signal": signal, "batch_count": strconv.FormatInt(count, 10)}})
			}
		}
	}
	return runtime, combined, stop, cfg.Enabled, nil
}

func observeRepositories(value repositories, telemetry ports.Telemetry) repositories {
	value.audit = observability.ObserveAudit(value.audit, telemetry)
	value.blobs = observability.ObserveBlobs(value.blobs, telemetry)
	value.imageProcessor = observability.ObserveImages(value.imageProcessor, telemetry)
	value.directUploads = observability.ObserveUploads(value.directUploads, telemetry)
	return value
}

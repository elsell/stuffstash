package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type App struct {
	observer ports.Observer
}

func New(observer ports.Observer) App {
	return App{observer: observer}
}

func (a App) Health(ctx context.Context) HealthStatus {
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventHealthChecked,
		Message: "health check completed",
	})

	return HealthStatus{
		Service: ServiceNameStuffStash,
		Status:  HealthStatusHealthy,
	}
}

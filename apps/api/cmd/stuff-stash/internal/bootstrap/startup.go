package bootstrap

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func RecordStartupFailure(observer ports.Observer, err error) {
	observer.Record(context.Background(), ports.Event{
		Name:    ports.EventApplicationStartupFailed,
		Message: "application startup failed",
		Fields:  map[string]string{"error": err.Error()},
	})
}

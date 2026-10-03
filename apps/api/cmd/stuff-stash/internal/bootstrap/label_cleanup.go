package bootstrap

import (
	"context"
	labelapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sync"
	"time"
)

func startLabelCleanup(ctx context.Context, service *labelapp.LabelService, interval time.Duration, observer ports.Observer) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if err := service.PurgeExpired(workerCtx); err != nil && workerCtx.Err() == nil && observer != nil {
				observer.Record(workerCtx, ports.Event{Name: ports.EventLabelCleanupFailed, Message: "Expired label previews could not be removed"})
			}
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return func() { cancel(); wg.Wait() }
}

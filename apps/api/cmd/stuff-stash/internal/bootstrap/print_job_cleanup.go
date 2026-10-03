package bootstrap

import (
	"context"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sync"
	"time"
)

func startPrintJobCleanup(ctx context.Context, service *printingapp.JobService, observer ports.Observer) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(service.CleanupInterval())
		defer ticker.Stop()
		after := ""
		for {
			page, err := service.Maintain(workerCtx, after)
			if err != nil {
				if workerCtx.Err() == nil && observer != nil {
					observer.Record(workerCtx, ports.Event{Name: ports.EventPrintJobCleanupFailed, Message: "Print job maintenance could not complete"})
				}
			} else {
				after = page.After
				if !page.HasMore {
					after = ""
				} else if workerCtx.Err() == nil {
					continue
				}
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

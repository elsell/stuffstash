package bootstrap

import (
	"context"
	"sync"
	"time"

	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type thumbnailDrainer interface {
	Drain(context.Context) (bool, error)
}

func startThumbnailWorkers(ctx context.Context, worker thumbnailDrainer, observer ports.Observer, cfg config.ThumbnailConfig) func() {
	if !cfg.WorkerEnabled {
		return func() {}
	}
	lifetime, cancel := context.WithCancel(ctx)
	var workers sync.WaitGroup
	for range cfg.Concurrency {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for lifetime.Err() == nil {
				draining, end := context.WithTimeout(lifetime, cfg.LeaseDuration)
				worked, err := worker.Drain(draining)
				end()
				if lifetime.Err() != nil {
					return
				}
				if err != nil {
					observer.Record(lifetime, ports.Event{Name: ports.EventThumbnailWorkerFailed, Message: "thumbnail worker drain failed", Fields: map[string]string{"outcome": "drain_failed"}})
				}
				if worked && err == nil {
					continue
				}
				timer := time.NewTimer(cfg.PollInterval)
				select {
				case <-lifetime.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
			}
		}()
	}
	return func() { cancel(); workers.Wait() }
}

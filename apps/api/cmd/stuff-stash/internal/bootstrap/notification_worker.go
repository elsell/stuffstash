package bootstrap

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

type notificationSweeper interface {
	GenerateSweepPage(context.Context, notifications.GenerationCursor, int) (notifications.GenerationCursor, bool, error)
}

func startNotificationWorker(parent context.Context, sweeper notificationSweeper, observer ports.Observer, cfg config.NotificationConfig) func() {
	if !cfg.Enabled {
		return func() {}
	}
	ctx, cancel := context.WithCancel(parent)
	done := make(chan struct{})
	go func() {
		defer close(done)
		cursor := notifications.GenerationCursor{}
		for {
			if ctx.Err() != nil {
				return
			}
			pageCtx, pageCancel := context.WithTimeout(ctx, cfg.PageTimeout)
			next, _, err := sweeper.GenerateSweepPage(pageCtx, cursor, cfg.PageSize)
			pageCancel()
			if err == nil {
				cursor = next
			} else if ctx.Err() == nil && observer != nil {
				observer.Record(ctx, ports.Event{Name: ports.EventNotificationWorkerFailed, Message: "expiration notification sweep failed", Fields: map[string]string{"error": err.Error()}})
			}
			timer := time.NewTimer(cfg.PollInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}()
	return func() { cancel(); <-done }
}

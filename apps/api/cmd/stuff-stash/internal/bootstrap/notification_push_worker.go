package bootstrap

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type notificationDeliverer interface {
	DeliverPage(context.Context, int, time.Duration, notification.RetryPolicy) (int, error)
}

func startNotificationPushWorker(parent context.Context, deliverer notificationDeliverer, observer ports.Observer, cfg config.NotificationPushConfig) func() {
	if !cfg.Enabled() {
		return func() {}
	}
	ctx, cancel := context.WithCancel(parent)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if ctx.Err() != nil {
				return
			}
			pageCtx, pageCancel := context.WithTimeout(ctx, cfg.PageTimeout)
			_, err := deliverer.DeliverPage(pageCtx, cfg.PageSize, cfg.Lease, cfg.Retry)
			pageCancel()
			if err != nil && ctx.Err() == nil && observer != nil {
				observer.Record(ctx, ports.Event{Name: ports.EventNotificationWorkerFailed, Message: "notification delivery page failed", Fields: map[string]string{"phase": "delivery"}})
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

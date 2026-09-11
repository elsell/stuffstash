package bootstrap

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"testing"
	"time"
)

type deliveryWorkerFake struct{ started, stopped chan struct{} }

func (f deliveryWorkerFake) DeliverPage(ctx context.Context, limit int, lease time.Duration, policy notification.RetryPolicy) (int, error) {
	close(f.started)
	<-ctx.Done()
	close(f.stopped)
	return 0, ctx.Err()
}
func TestPushWorkerCancelsAndJoinsDelivery(t *testing.T) {
	fake := deliveryWorkerFake{started: make(chan struct{}), stopped: make(chan struct{})}
	cfg := config.NotificationPushConfig{APNSEnabled: true, PageSize: 10, PollInterval: time.Hour, PageTimeout: time.Minute, Lease: 2 * time.Minute, Retry: notification.RetryPolicy{MaxAttempts: 3, InitialDelay: time.Minute, MaximumDelay: time.Hour}}
	stop := startNotificationPushWorker(context.Background(), fake, nil, cfg)
	select {
	case <-fake.started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}
	stop()
	select {
	case <-fake.stopped:
	default:
		t.Fatal("worker not joined")
	}
}
func TestPushWorkerDisabledDoesNotStart(t *testing.T) {
	fake := deliveryWorkerFake{started: make(chan struct{}), stopped: make(chan struct{})}
	stop := startNotificationPushWorker(context.Background(), fake, nil, config.NotificationPushConfig{})
	stop()
	select {
	case <-fake.started:
		t.Fatal("disabled worker ran")
	default:
	}
}

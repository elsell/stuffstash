package bootstrap

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/config"
	"testing"
	"time"
)

type retryNotificationSweep struct {
	cursors  chan notifications.GenerationCursor
	attempts int
}

func (f *retryNotificationSweep) GenerateSweepPage(ctx context.Context, cursor notifications.GenerationCursor, _ int) (notifications.GenerationCursor, bool, error) {
	f.attempts++
	f.cursors <- cursor
	switch f.attempts {
	case 1:
		return notifications.GenerationCursor{AfterRecipientID: "must-not-advance"}, false, errors.New("temporary repository failure")
	case 2:
		return notifications.GenerationCursor{AfterRecipientID: "completed"}, false, nil
	default:
		<-ctx.Done()
		return cursor, false, ctx.Err()
	}
}
func TestNotificationWorkerRetriesWithoutSkippingFailedPage(t *testing.T) {
	fake := &retryNotificationSweep{cursors: make(chan notifications.GenerationCursor, 3)}
	stop := startNotificationWorker(context.Background(), fake, nil, config.NotificationConfig{Enabled: true, PageSize: 10, PollInterval: 100 * time.Millisecond, PageTimeout: time.Minute})
	defer stop()
	for _, expected := range []string{"", "", "completed"} {
		select {
		case cursor := <-fake.cursors:
			if cursor.AfterRecipientID != expected {
				t.Fatalf("cursor %q want %q", cursor.AfterRecipientID, expected)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("worker did not retry")
		}
	}
}

type notificationSweepFake struct {
	started chan struct{}
	exited  chan struct{}
}

func (f notificationSweepFake) GenerateSweepPage(ctx context.Context, cursor notifications.GenerationCursor, _ int) (notifications.GenerationCursor, bool, error) {
	close(f.started)
	<-ctx.Done()
	close(f.exited)
	return cursor, false, ctx.Err()
}
func TestNotificationWorkerStopsBeforeReturning(t *testing.T) {
	fake := notificationSweepFake{started: make(chan struct{}), exited: make(chan struct{})}
	cfg := config.NotificationConfig{Enabled: true, PageSize: 10, PollInterval: time.Second, PageTimeout: time.Minute}
	stop := startNotificationWorker(context.Background(), fake, nil, cfg)
	select {
	case <-fake.started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start immediately")
	}
	stop()
	select {
	case <-fake.exited:
	default:
		t.Fatal("worker not joined")
	}
}
func TestDisabledNotificationWorkerDoesNotRun(t *testing.T) {
	fake := notificationSweepFake{started: make(chan struct{}), exited: make(chan struct{})}
	stop := startNotificationWorker(context.Background(), fake, nil, config.NotificationConfig{})
	stop()
	select {
	case <-fake.started:
		t.Fatal("disabled worker started")
	default:
	}
}

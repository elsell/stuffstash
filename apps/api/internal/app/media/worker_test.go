package media

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/worklimit"
	domain "github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestWorkerCompletesAndReleasesSharedCapacity(t *testing.T) {
	worker, queue, processor, admission := workerFixture(t)
	worked, err := worker.Drain(context.Background())
	if err != nil || !worked || !processor.prepared || queue.result.Status != ports.ThumbnailJobCompleted {
		t.Fatalf("work did not complete: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	release, err := admission.Acquire(ctx, ports.ImageWorkForeground)
	if err != nil {
		t.Fatal("worker leaked shared capacity")
	}
	release()
}

func TestWorkerRetriesThenExhausts(t *testing.T) {
	worker, queue, processor, _ := workerFixture(t)
	processor.failed = true
	if _, err := worker.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	if queue.result.Status != ports.ThumbnailJobPending || !queue.result.NextAttemptAt.Equal(queue.now.Add(time.Second)) {
		t.Fatal("failure did not schedule backoff")
	}
	queue.attempts = 3
	if _, err := worker.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	if queue.result.Status != ports.ThumbnailJobFailed {
		t.Fatal("exhausted work was retried")
	}
}

func TestWorkerDoesNotProcessBeyondCrashBudget(t *testing.T) {
	worker, queue, processor, _ := workerFixture(t)
	queue.attempts = 4
	if _, err := worker.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	if processor.prepared || queue.result.Status != ports.ThumbnailJobFailed {
		t.Fatal("crashed work bypassed attempt limit")
	}
}

func TestWorkerShutdownLeavesClaimRecoverable(t *testing.T) {
	worker, queue, processor, _ := workerFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	processor.cancel = cancel
	if _, err := worker.Drain(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("shutdown ignored: %v", err)
	}
	if queue.result.Status != "" {
		t.Fatal("shutdown acknowledged unfinished work")
	}
}

func TestWorkerCannotBypassForegroundCapacity(t *testing.T) {
	worker, queue, _, admission := workerFixture(t)
	release, _ := admission.Acquire(context.Background(), ports.ImageWorkForeground)
	defer release()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := worker.Drain(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("shared admission ignored: %v", err)
	}
	if queue.claimed {
		t.Fatal("worker claimed work without capacity")
	}
}

type workerQueue struct {
	now      time.Time
	attempts int
	claimed  bool
	result   ports.ThumbnailJobResolution
}

func (q *workerQueue) Now() time.Time { return q.now }
func (q *workerQueue) ClaimThumbnailJobs(_ context.Context, token string, _ int, _ time.Time, until time.Time) ([]ports.ClaimedThumbnailJob, error) {
	q.claimed = true
	return []ports.ClaimedThumbnailJob{{Job: domain.ThumbnailJob{AttachmentID: "image"}, ClaimID: token, ClaimedUntil: until, Attempts: q.attempts}}, nil
}
func (q *workerQueue) ResolveThumbnailJob(_ context.Context, _ ports.ClaimedThumbnailJob, result ports.ThumbnailJobResolution) error {
	q.result = result
	return result.Validate()
}

type workerIDs struct{}

func (workerIDs) NewID() string { return "unique-work-claim" }

type workerObserver struct{}

func (workerObserver) Record(context.Context, ports.Event) {}

type workerProcessor struct {
	prepared, failed bool
	cancel           context.CancelFunc
}

func (p *workerProcessor) ProcessThumbnailJob(ctx context.Context, _ ports.ClaimedThumbnailJob) error {
	if p.cancel != nil {
		p.cancel()
		return ctx.Err()
	}
	if p.failed {
		return errors.New("controlled image processing failure")
	}
	p.prepared = true
	return nil
}
func workerFixture(t *testing.T) (*Worker, *workerQueue, *workerProcessor, *worklimit.Limiter) {
	t.Helper()
	queue := &workerQueue{now: time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC), attempts: 1}
	processor := &workerProcessor{}
	admission, _ := worklimit.New(1)
	worker, err := NewWorker(queue, processor, admission, queue, workerIDs{}, workerObserver{}, WorkerConfig{MaxAttempts: 3, LeaseDuration: time.Minute, ProcessingTimeout: 30 * time.Second, RetryBase: time.Second, RetryMax: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	return worker, queue, processor, admission
}

package bootstrap

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

type thumbnailBlockingDrainer struct {
	entered chan struct{}
	exited  chan struct{}
}

func (d thumbnailBlockingDrainer) Drain(ctx context.Context) (bool, error) {
	d.entered <- struct{}{}
	<-ctx.Done()
	d.exited <- struct{}{}
	return false, ctx.Err()
}

type thumbnailTestObserver struct{}

func (thumbnailTestObserver) Record(context.Context, ports.Event) {}
func TestThumbnailWorkersStopAndJoinEveryLoop(t *testing.T) {
	cfg, err := config.LoadThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Concurrency = 2
	drainer := thumbnailBlockingDrainer{make(chan struct{}, 2), make(chan struct{}, 2)}
	stop := startThumbnailWorkers(context.Background(), drainer, thumbnailTestObserver{}, cfg)
	defer stop()
	for range 2 {
		select {
		case <-drainer.entered:
		case <-time.After(time.Second):
			t.Fatal("missing worker")
		}
	}
	done := make(chan struct{})
	go func() { stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not join workers")
	}
	if len(drainer.exited) != 2 {
		t.Fatal("shutdown returned before workers exited")
	}
}
func TestDisabledThumbnailWorkersDoNotDrain(t *testing.T) {
	cfg, err := config.LoadThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	cfg.WorkerEnabled = false
	drainer := thumbnailBlockingDrainer{make(chan struct{}, 2), make(chan struct{}, 2)}
	stop := startThumbnailWorkers(context.Background(), drainer, thumbnailTestObserver{}, cfg)
	stop()
	if len(drainer.entered) != 0 {
		t.Fatal("disabled worker drained")
	}
}

func TestThumbnailRuntimeKeepsForegroundReaderWhenWorkersDisabled(t *testing.T) {
	cfg, err := config.LoadThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	cfg.WorkerEnabled = false
	repositories, closeStore, err := buildRepositories(context.Background(), config.Config{RepositoryMode: "memory"})
	if err != nil {
		t.Fatal(err)
	}
	defer closeStore()
	reader, worker, err := buildThumbnailRuntime(repositories, cfg, thumbnailTestObserver{})
	if err != nil || reader == nil || worker == nil {
		t.Fatal("disabled background configuration lost foreground admission", err)
	}
}

type thumbnailCompletedBackfill struct{ calls chan struct{} }

func (b thumbnailCompletedBackfill) BackfillThumbnailJobs(context.Context, int, time.Time) (ports.ThumbnailBackfillProgress, error) {
	b.calls <- struct{}{}
	return ports.ThumbnailBackfillProgress{Complete: true}, nil
}
func TestThumbnailBackfillStopsAtCompletion(t *testing.T) {
	cfg, err := config.LoadThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	cfg.BackfillEnabled = true
	cfg.BackfillInterval = 100 * time.Millisecond
	backfill := thumbnailCompletedBackfill{calls: make(chan struct{}, 10)}
	stop := startThumbnailBackfill(context.Background(), backfill, ports.SystemClock{}, thumbnailTestObserver{}, cfg)
	defer stop()
	select {
	case <-backfill.calls:
	case <-time.After(time.Second):
		t.Fatal("backfill did not start")
	}
	select {
	case <-backfill.calls:
		t.Fatal("completed backfill kept scanning")
	case <-time.After(250 * time.Millisecond):
	}
}

type thumbnailControlledBackfill struct {
	calls     chan struct{}
	cancelled chan struct{}
	finish    chan struct{}
	failFirst bool
}

func (b *thumbnailControlledBackfill) BackfillThumbnailJobs(ctx context.Context, _ int, _ time.Time) (ports.ThumbnailBackfillProgress, error) {
	b.calls <- struct{}{}
	if b.failFirst {
		b.failFirst = false
		return ports.ThumbnailBackfillProgress{}, errors.New("controlled failure")
	}
	if b.cancelled != nil {
		<-ctx.Done()
		close(b.cancelled)
		<-b.finish
		return ports.ThumbnailBackfillProgress{}, ctx.Err()
	}
	return ports.ThumbnailBackfillProgress{Complete: true}, nil
}
func TestThumbnailBackfillShutdownWaitsForCancelledBatch(t *testing.T) {
	cfg, err := config.LoadThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	cfg.BackfillEnabled = true
	backfill := &thumbnailControlledBackfill{calls: make(chan struct{}, 2), cancelled: make(chan struct{}), finish: make(chan struct{})}
	stop := startThumbnailBackfill(context.Background(), backfill, ports.SystemClock{}, thumbnailTestObserver{}, cfg)
	select {
	case <-backfill.calls:
	case <-time.After(time.Second):
		t.Fatal("batch did not enter")
	}
	done := make(chan struct{})
	go func() { stop(); close(done) }()
	select {
	case <-backfill.cancelled:
	case <-time.After(time.Second):
		t.Fatal("batch not cancelled")
	}
	select {
	case <-done:
		t.Fatal("shutdown returned before batch unwound")
	default:
	}
	close(backfill.finish)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not join batch")
	}
}
func TestThumbnailBackfillRetriesFailedBatch(t *testing.T) {
	cfg, err := config.LoadThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	cfg.BackfillEnabled = true
	cfg.BackfillInterval = 100 * time.Millisecond
	backfill := &thumbnailControlledBackfill{calls: make(chan struct{}, 2), failFirst: true}
	stop := startThumbnailBackfill(context.Background(), backfill, ports.SystemClock{}, thumbnailTestObserver{}, cfg)
	defer stop()
	for range 2 {
		select {
		case <-backfill.calls:
		case <-time.After(time.Second):
			t.Fatal("failed backfill was not retried")
		}
	}
}

type observedCleanupQueue struct{ entered chan struct{} }

func (q observedCleanupQueue) ClaimBlobDeletionRechecks(ctx context.Context, _ string, _ int, _, _ time.Time, _ time.Duration) ([]ports.BlobDeletionEvent, error) {
	select {
	case q.entered <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return nil, nil
}
func (observedCleanupQueue) ResolveBlobDeletionRecheck(context.Context, ports.BlobDeletionEvent, time.Time, bool) error {
	return errors.New("no cleanup claim was returned")
}
func TestCleanupRunsWhenThumbnailGenerationDisabled(t *testing.T) {
	cfg, err := config.LoadThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	cfg.WorkerEnabled = false
	entered := make(chan struct{}, 1)
	stop, err := startBlobDeletionRechecks(context.Background(), repositories{blobs: memory.NewStore(), blobDeletionRechecks: observedCleanupQueue{entered: entered}}, thumbnailTestObserver{}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("cleanup was disabled with thumbnail generation")
	}
}

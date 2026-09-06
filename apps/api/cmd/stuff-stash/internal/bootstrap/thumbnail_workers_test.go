package bootstrap

import (
	"context"
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

package bootstrap

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"path/filepath"
	"testing"
	"time"
)

type blockingArchiveRuntime struct{ entered, exited chan struct{} }

func (b blockingArchiveRuntime) Drain(ctx context.Context) (bool, error) {
	b.entered <- struct{}{}
	<-ctx.Done()
	b.exited <- struct{}{}
	return false, ctx.Err()
}
func (b blockingArchiveRuntime) CleanupArchives(ctx context.Context, _ int) error {
	b.entered <- struct{}{}
	<-ctx.Done()
	b.exited <- struct{}{}
	return ctx.Err()
}
func TestArchiveWorkersCancelAndJoinExecutionAndCleanup(t *testing.T) {
	cfg, err := config.LoadArchives("sqlite")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Concurrency = 2
	runtime := blockingArchiveRuntime{make(chan struct{}, 3), make(chan struct{}, 3)}
	stop := startArchiveWorkers(context.Background(), runtime, runtime, thumbnailTestObserver{}, cfg)
	defer stop()
	for range 3 {
		select {
		case <-runtime.entered:
		case <-time.After(time.Second):
			t.Fatal("missing loop")
		}
	}
	done := make(chan struct{})
	go func() { stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shutdown hung")
	}
	if len(runtime.exited) != 3 {
		t.Fatal("shutdown did not join work")
	}
}

func TestArchiveRuntimeSupportsObservedPersistentRepositories(t *testing.T) {
	cfg, err := config.LoadArchives("sqlite")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ScratchDirectory = t.TempDir()
	repos, closeStore, err := buildRepositories(context.Background(), config.Config{RepositoryMode: "sqlite", DatabaseDSN: filepath.Join(t.TempDir(), "archive.db"), BlobStorageMode: "filesystem", BlobStoragePath: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer closeStore()
	repos = observeRepositories(repos, ports.NoopTelemetry{})
	service, worker, err := buildArchiveRuntime(cfg, repos, memory.NewAuthorizer(), thumbnailTestObserver{})
	if err != nil || service == nil || worker == nil {
		t.Fatalf("observed archive runtime: %v", err)
	}
}

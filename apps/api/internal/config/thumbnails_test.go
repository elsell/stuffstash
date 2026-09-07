package config

import (
	"testing"
	"time"
)

func TestThumbnailConfigurationDefaultsAndDisable(t *testing.T) {
	cfg, err := LoadThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.WorkerEnabled || cfg.Concurrency != 1 || cfg.ProcessingTimeout != time.Minute {
		t.Fatal("unexpected defaults", cfg)
	}
	t.Setenv("STUFF_STASH_THUMBNAIL_WORKER_ENABLED", "false")
	t.Setenv("STUFF_STASH_THUMBNAIL_CONCURRENCY", "2")
	cfg, err = LoadThumbnails()
	if err != nil || cfg.WorkerEnabled || cfg.Concurrency != 2 {
		t.Fatal("configuration ignored", cfg, err)
	}
}
func TestThumbnailConfigurationRejectsUnsafeValues(t *testing.T) {
	for _, input := range []struct{ key, value string }{
		{"CONCURRENCY", "0"}, {"CONCURRENCY", "9"}, {"CONCURRENCY", "bad"},
		{"WORKER_ENABLED", "maybe"}, {"POLL_INTERVAL", "0s"}, {"POLL_INTERVAL", "bad"},
		{"LEASE_DURATION", "30s"}, {"PUBLICATION_TIMEOUT", "60s"},
		{"MAX_ATTEMPTS", "101"}, {"RETRY_MAX", "1s"},
	} {
		t.Run(input.key+input.value, func(t *testing.T) {
			t.Setenv("STUFF_STASH_THUMBNAIL_"+input.key, input.value)
			if _, err := LoadThumbnails(); err == nil {
				t.Fatal("invalid setting accepted")
			}
		})
	}
}

func TestThumbnailBackfillConfiguration(t *testing.T) {
	cfg, err := LoadThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BackfillEnabled || cfg.BackfillBatchSize != 25 || cfg.BackfillInterval != 5*time.Second {
		t.Fatal("unsafe backfill defaults")
	}
	t.Setenv("STUFF_STASH_THUMBNAIL_BACKFILL_ENABLED", "true")
	t.Setenv("STUFF_STASH_THUMBNAIL_BACKFILL_BATCH_SIZE", "1001")
	if _, err := LoadThumbnails(); err == nil {
		t.Fatal("unbounded backfill accepted")
	}
	t.Setenv("STUFF_STASH_THUMBNAIL_BACKFILL_BATCH_SIZE", "10")
	cfg, err = LoadThumbnails()
	if err != nil || !cfg.BackfillEnabled || cfg.BackfillBatchSize != 10 {
		t.Fatal("backfill setting ignored", err)
	}
}

func TestCleanupRecheckIntervalBounds(t *testing.T) {
	cfg, err := LoadThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CleanupRecheckInterval != time.Hour {
		t.Fatal("unexpected cleanup interval")
	}
	t.Setenv("STUFF_STASH_THUMBNAIL_CLEANUP_RECHECK_INTERVAL", "1s")
	if _, err := LoadThumbnails(); err == nil {
		t.Fatal("too frequent recheck accepted")
	}
}

func TestForegroundCachePollingConfiguration(t *testing.T) {
	cfg, err := LoadThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ForegroundCachePollInterval != 250*time.Millisecond {
		t.Fatal("unexpected cache polling default")
	}
	for _, value := range []string{"0s", "99ms", "6s", "bad"} {
		t.Setenv("STUFF_STASH_THUMBNAIL_FOREGROUND_CACHE_POLL_INTERVAL", value)
		if _, err := LoadThumbnails(); err == nil {
			t.Fatal("unsafe cache polling accepted", value)
		}
	}
	t.Setenv("STUFF_STASH_THUMBNAIL_FOREGROUND_CACHE_POLL_INTERVAL", "500ms")
	cfg, err = LoadThumbnails()
	if err != nil || cfg.ForegroundCachePollInterval != 500*time.Millisecond {
		t.Fatal("cache polling setting ignored", err)
	}
}

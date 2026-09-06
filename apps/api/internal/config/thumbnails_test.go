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

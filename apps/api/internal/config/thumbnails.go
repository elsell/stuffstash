package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type ThumbnailConfig struct {
	CleanupRecheckInterval                                                                  time.Duration
	BackfillEnabled                                                                         bool
	BackfillBatchSize                                                                       int
	BackfillInterval                                                                        time.Duration
	WorkerEnabled                                                                           bool
	Concurrency, MaxAttempts                                                                int
	PollInterval, LeaseDuration, ProcessingTimeout, PublicationTimeout, RetryBase, RetryMax time.Duration
}

func LoadThumbnails() (ThumbnailConfig, error) {
	cfg := ThumbnailConfig{CleanupRecheckInterval: time.Hour, BackfillBatchSize: 25, BackfillInterval: 5 * time.Second, WorkerEnabled: true, Concurrency: 1, MaxAttempts: 5, PollInterval: time.Second, LeaseDuration: 90 * time.Second, ProcessingTimeout: time.Minute, PublicationTimeout: 15 * time.Second, RetryBase: 5 * time.Second, RetryMax: 5 * time.Minute}
	for _, field := range []struct {
		name   string
		target *bool
	}{{"WORKER_ENABLED", &cfg.WorkerEnabled}, {"BACKFILL_ENABLED", &cfg.BackfillEnabled}} {
		if value, exists := os.LookupEnv("STUFF_STASH_THUMBNAIL_" + field.name); exists {
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return cfg, fmt.Errorf("invalid thumbnail %s", field.name)
			}
			*field.target = parsed
		}
	}

	for _, field := range []struct {
		name   string
		target *int
	}{{"BACKFILL_BATCH_SIZE", &cfg.BackfillBatchSize}, {"CONCURRENCY", &cfg.Concurrency}, {"MAX_ATTEMPTS", &cfg.MaxAttempts}} {
		if value, exists := os.LookupEnv("STUFF_STASH_THUMBNAIL_" + field.name); exists {
			parsed, err := strconv.Atoi(value)
			if err != nil {
				return cfg, fmt.Errorf("invalid thumbnail %s", field.name)
			}
			*field.target = parsed
		}
	}
	for _, field := range []struct {
		name   string
		target *time.Duration
	}{
		{"CLEANUP_RECHECK_INTERVAL", &cfg.CleanupRecheckInterval}, {"BACKFILL_INTERVAL", &cfg.BackfillInterval}, {"POLL_INTERVAL", &cfg.PollInterval}, {"LEASE_DURATION", &cfg.LeaseDuration}, {"PROCESSING_TIMEOUT", &cfg.ProcessingTimeout},
		{"PUBLICATION_TIMEOUT", &cfg.PublicationTimeout}, {"RETRY_BASE", &cfg.RetryBase}, {"RETRY_MAX", &cfg.RetryMax},
	} {
		if value, exists := os.LookupEnv("STUFF_STASH_THUMBNAIL_" + field.name); exists {
			parsed, err := time.ParseDuration(value)
			if err != nil {
				return cfg, fmt.Errorf("invalid thumbnail %s", field.name)
			}
			*field.target = parsed
		}
	}
	return cfg, cfg.Validate()
}

func (c ThumbnailConfig) Validate() error {
	if c.BackfillBatchSize < 1 || c.BackfillBatchSize > 1000 {
		return errors.New("invalid thumbnail backfill batch bounds")
	}
	if c.Concurrency < 1 || c.Concurrency > 8 || c.MaxAttempts < 1 || c.MaxAttempts > 100 {
		return errors.New("invalid thumbnail concurrency or attempt bounds")
	}
	for _, field := range []struct {
		name            string
		value, min, max time.Duration
	}{
		{"CLEANUP_RECHECK_INTERVAL", c.CleanupRecheckInterval, time.Minute, 30 * 24 * time.Hour},
		{"BACKFILL_INTERVAL", c.BackfillInterval, 100 * time.Millisecond, time.Hour},
		{"POLL_INTERVAL", c.PollInterval, 100 * time.Millisecond, time.Minute},
		{"LEASE_DURATION", c.LeaseDuration, 2 * time.Second, 10 * time.Minute},
		{"PROCESSING_TIMEOUT", c.ProcessingTimeout, time.Second, 5 * time.Minute},
		{"PUBLICATION_TIMEOUT", c.PublicationTimeout, 100 * time.Millisecond, time.Minute},
		{"RETRY_BASE", c.RetryBase, 100 * time.Millisecond, time.Hour},
		{"RETRY_MAX", c.RetryMax, 100 * time.Millisecond, 24 * time.Hour},
	} {
		if field.value < field.min || field.value > field.max {
			return fmt.Errorf("invalid thumbnail %s bounds", field.name)
		}
	}
	if c.LeaseDuration <= c.ProcessingTimeout || c.PublicationTimeout >= c.ProcessingTimeout || c.RetryMax < c.RetryBase {
		return errors.New("invalid thumbnail timeout or retry relationship")
	}
	return nil
}

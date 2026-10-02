package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type ArchiveConfig struct {
	Enabled                                                                                 bool
	MaxBytes, ExpandedBytes, MetadataBytes, EntryBytes                                      int64
	MaxRecords, MaxEntries, Concurrency                                                     int
	Retention, Lease, Heartbeat, PollInterval, CleanupInterval, JobTimeout, TransferTimeout time.Duration
	ScratchDirectory                                                                        string
}

func LoadArchives(repositoryMode string) (ArchiveConfig, error) {
	persistent := strings.EqualFold(repositoryMode, "postgres") || strings.EqualFold(repositoryMode, "sqlite")
	c := ArchiveConfig{Enabled: persistent, MaxBytes: 1 << 30, ExpandedBytes: 4 << 30, MetadataBytes: 64 << 20, EntryBytes: 512 << 20, MaxRecords: 100000, MaxEntries: 100000, Concurrency: 1, Retention: 24 * time.Hour, Lease: 2 * time.Minute, Heartbeat: 20 * time.Second, PollInterval: 2 * time.Second, CleanupInterval: time.Minute, JobTimeout: 2 * time.Hour, TransferTimeout: 30 * time.Minute, ScratchDirectory: os.TempDir()}
	if raw, ok := os.LookupEnv("STUFF_STASH_ARCHIVE_ENABLED"); ok {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return c, fmt.Errorf("invalid archive ENABLED")
		}
		c.Enabled = value
	}
	if c.Enabled && !persistent {
		return c, fmt.Errorf("archive jobs require PostgreSQL or SQLite")
	}
	for _, f := range []struct {
		name   string
		target *int64
	}{{"MAX_BYTES", &c.MaxBytes}, {"EXPANDED_BYTES", &c.ExpandedBytes}, {"METADATA_BYTES", &c.MetadataBytes}, {"ENTRY_BYTES", &c.EntryBytes}} {
		if raw, ok := os.LookupEnv("STUFF_STASH_ARCHIVE_" + f.name); ok {
			n, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return c, fmt.Errorf("invalid archive %s", f.name)
			}
			*f.target = n
		}
	}
	for _, f := range []struct {
		name   string
		target *int
	}{{"MAX_RECORDS", &c.MaxRecords}, {"MAX_ENTRIES", &c.MaxEntries}, {"CONCURRENCY", &c.Concurrency}} {
		if raw, ok := os.LookupEnv("STUFF_STASH_ARCHIVE_" + f.name); ok {
			n, err := strconv.Atoi(raw)
			if err != nil {
				return c, fmt.Errorf("invalid archive %s", f.name)
			}
			*f.target = n
		}
	}
	for _, f := range []struct {
		name   string
		target *time.Duration
	}{{"RETENTION", &c.Retention}, {"LEASE", &c.Lease}, {"HEARTBEAT", &c.Heartbeat}, {"POLL_INTERVAL", &c.PollInterval}, {"CLEANUP_INTERVAL", &c.CleanupInterval}, {"JOB_TIMEOUT", &c.JobTimeout}, {"TRANSFER_TIMEOUT", &c.TransferTimeout}} {
		if raw, ok := os.LookupEnv("STUFF_STASH_ARCHIVE_" + f.name); ok {
			n, err := time.ParseDuration(raw)
			if err != nil {
				return c, fmt.Errorf("invalid archive %s", f.name)
			}
			*f.target = n
		}
	}
	if raw, ok := os.LookupEnv("STUFF_STASH_ARCHIVE_SCRATCH_DIRECTORY"); ok {
		c.ScratchDirectory = strings.TrimSpace(raw)
	}
	return c, c.Validate()
}
func (c ArchiveConfig) Validate() error {
	if c.MaxBytes < 1 || c.MaxBytes > 5<<30 || c.ExpandedBytes < 1 || c.ExpandedBytes > 100<<30 || c.MetadataBytes < 1 || c.MetadataBytes > c.MaxBytes || c.EntryBytes < 1 || c.EntryBytes > 5<<30 || c.MaxRecords < 1 || c.MaxRecords > 1000000 || c.MaxEntries < 2 || c.MaxEntries > 1000000 || c.Concurrency < 1 || c.Concurrency > 8 || c.ScratchDirectory == "" {
		return fmt.Errorf("invalid archive size, record, concurrency, or scratch limits")
	}
	if c.Retention < time.Hour || c.Retention > 30*24*time.Hour || c.Lease < time.Second || c.Lease > 10*time.Minute || c.Heartbeat < 100*time.Millisecond || c.Heartbeat >= c.Lease/2 || c.JobTimeout < time.Second || c.JobTimeout >= c.Retention || c.TransferTimeout < time.Second || c.TransferTimeout >= c.Retention || c.PollInterval < 100*time.Millisecond || c.PollInterval > time.Minute || c.CleanupInterval < time.Second || c.CleanupInterval > time.Hour {
		return fmt.Errorf("invalid archive scheduling or retention limits")
	}
	return nil
}

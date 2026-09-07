package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/config"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestThumbnailCommandRejectsInvalidArgumentsBeforeDatabaseAccess(t *testing.T) {
	for _, args := range [][]string{nil, {"unknown"}, {"status", "extra"}, {"retry-failed", "--limit", "0"}, {"retry-failed", "--limit", "1001"}, {"retry-failed", "--limit", "bad"}} {
		err := RunThumbnailJobsCommand(context.Background(), config.Config{}, args, io.Discard, thumbnailTestObserver{})
		if err == nil || err.Error() == "database dsn is required" {
			t.Fatal("invalid command reached database configuration", args, err)
		}
	}
}

func TestThumbnailStatusUsesExistingDatabaseAndOnlyReturnsAggregates(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "queue.db")
	cfg := config.Config{RepositoryMode: "sqlite", DatabaseDSN: path}
	var output bytes.Buffer
	if err := RunThumbnailJobsCommand(ctx, cfg, []string{"status"}, &output, thumbnailTestObserver{}); err == nil {
		t.Fatal("status created missing database")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("status mutated missing database", err)
	}
	_, closeStore, err := openSQLiteStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := closeStore(); err != nil {
		t.Fatal(err)
	}
	if err := RunThumbnailJobsCommand(ctx, cfg, []string{"status"}, &output, thumbnailTestObserver{}); err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 6 || result["pending"] != float64(0) || result["backfill_complete"] != false {
		t.Fatal("unexpected status output", result)
	}
}

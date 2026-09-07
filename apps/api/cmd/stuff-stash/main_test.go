package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestThumbnailCommandProcessOutput(t *testing.T) {
	if os.Getenv("STUFF_STASH_TEST_COMMAND_PROCESS") == "1" {
		os.Args = []string{os.Args[0], "thumbnail-jobs", "retry-failed"}
		main()
		os.Exit(0)
	}
	path := filepath.Join(t.TempDir(), "queue.db")
	db, err := gormstore.OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := gormstore.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := pool.Close(); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^TestThumbnailCommandProcessOutput$")
	command.Env = append(os.Environ(), "STUFF_STASH_TEST_COMMAND_PROCESS=1", "STUFF_STASH_REPOSITORY_MODE=sqlite", "STUFF_STASH_DATABASE_DSN="+path)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatal("command failed", err, stderr.String())
	}
	decoder := json.NewDecoder(&stdout)
	var result map[string]any
	if err := decoder.Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result["retried"] != float64(0) {
		t.Fatal("unexpected command result", result)
	}
	if err := decoder.Decode(&result); err != io.EOF {
		t.Fatal("stdout contained logs or another result", err)
	}
	if !strings.Contains(stderr.String(), "thumbnail_jobs.retried") {
		t.Fatal("operational event missing from stderr")
	}
}

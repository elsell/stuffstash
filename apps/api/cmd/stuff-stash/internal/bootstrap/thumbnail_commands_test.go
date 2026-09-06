package bootstrap

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/config"
	"io"
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

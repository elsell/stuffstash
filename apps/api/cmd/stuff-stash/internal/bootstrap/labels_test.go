package bootstrap

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/internal/config"
	"path/filepath"
	"testing"
)

func TestLabelsBootstrapCommandIsPersistentAndIdempotent(t *testing.T) {
	cfg := config.Config{RepositoryMode: "sqlite", DatabaseDSN: filepath.Join(t.TempDir(), "labels.db")}
	var first, second bytes.Buffer
	if err := RunLabelsCommand(context.Background(), cfg, []string{"bootstrap-instance"}, &first); err != nil {
		t.Fatal(err)
	}
	if err := RunLabelsCommand(context.Background(), cfg, []string{"bootstrap-instance"}, &second); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("bootstrap replaced persistent instance identity")
	}
}

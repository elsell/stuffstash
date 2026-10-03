package credentials

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestFileRefusesUnsafePermissionsAndOtherServer(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "credentials.json")
	store := File{Path: path}
	ctx := context.Background()
	s := ports.Session{Server: "https://one.example", IDToken: "private"}
	if err := store.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(ctx, "https://two.example"); err == nil {
		t.Fatal("accepted another server")
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(ctx, s.Server); err == nil {
		t.Fatal("read insecure credentials")
	}
	if err := store.Save(ctx, s); err == nil {
		t.Fatal("overwrote insecure credentials")
	}
}

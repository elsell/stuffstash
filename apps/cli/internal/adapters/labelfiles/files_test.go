package labelfiles

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPublicationPreservesExistingFilesAndLeavesNoPartialOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "label.png")
	content := []byte("verified image")
	if err := (Files{}).Publish(context.Background(), path, content); err != nil {
		t.Fatal(err)
	}
	if err := (Files{}).Publish(context.Background(), path, []byte("replacement")); err == nil {
		t.Fatal("overwrote existing file")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != string(content) {
		t.Fatalf("original changed: %q %v", data, err)
	}
	info, _ := os.Stat(path)
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatal("output is not private")
	}
	link := filepath.Join(dir, "link.png")
	if err := os.Symlink(path, link); err == nil {
		if err := (Files{}).Publish(context.Background(), link, []byte("replacement")); err == nil {
			t.Fatal("replaced symlink")
		}
		if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Fatal("symlink changed")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	canceled := filepath.Join(dir, "canceled.png")
	if err := (Files{}).Publish(ctx, canceled, content); err == nil {
		t.Fatal("canceled write succeeded")
	}
	if _, err := os.Stat(canceled); !os.IsNotExist(err) {
		t.Fatal("canceled output exists")
	}
	temporary, _ := filepath.Glob(filepath.Join(dir, ".stuffstash-label-*"))
	if len(temporary) != 0 {
		t.Fatal("temporary output leaked")
	}
}

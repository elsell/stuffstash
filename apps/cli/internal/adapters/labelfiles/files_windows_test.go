//go:build windows

package labelfiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsRefusesOutputWithoutPrivateFileAdapter(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "label.png")
	if err := (Files{}).Publish(context.Background(), path, []byte("private label")); err == nil {
		t.Fatal("file output succeeded without owner-only ACL verification")
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed output left files: %v %v", entries, err)
	}
}

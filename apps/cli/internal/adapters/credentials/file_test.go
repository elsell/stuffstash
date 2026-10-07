package credentials

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func TestCredentialFailuresDistinguishFileTypeFromPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := (File{Path: path}).Save(context.Background(), ports.Session{}); err == nil || !strings.Contains(err.Error(), "regular file") || strings.Contains(err.Error(), "0600") {
		t.Fatalf("file type guidance: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("private-session"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	var safe *ports.Error
	if _, err := (File{Path: path}).Load(context.Background(), "https://stash.example"); !errors.As(err, &safe) || safe.Category != "configuration" || !strings.Contains(err.Error(), "0600") || strings.Contains(err.Error(), "private-session") {
		t.Fatalf("permission guidance: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil || string(content) != "private-session" {
		t.Fatal("credential changed")
	}
}

//go:build linux

package usblp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/printstate"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestPhysicalLockExcludesWorkersWithDifferentJournalDirectories(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "device-node")
	original, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer original.Close()
	contender, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer contender.Close()
	// Journals intentionally use different roots: exclusion must come from the
	// underlying shared device inode, not coincidentally sharing recovery files.
	a, err := (printstate.Store{Directory: filepath.Join(directory, "a")}).Acquire(context.Background(), "physical-printer")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := (printstate.Store{Directory: filepath.Join(directory, "b")}).Acquire(context.Background(), "physical-printer")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if err = lockDevice(int(original.Fd())); err != nil {
		t.Fatal(err)
	}
	if err = lockDevice(int(contender.Fd())); !errors.Is(err, ports.ErrDeviceInUse) {
		t.Fatalf("second device lock accepted: %v", err)
	}
	if err = original.Close(); err != nil {
		t.Fatal(err)
	}
	if err = lockDevice(int(contender.Fd())); err != nil {
		t.Fatalf("lock survived process descriptor close: %v", err)
	}
}

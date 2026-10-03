//go:build linux

package printstate_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/printstate"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestDurableDeviceJournalAndExclusiveReservation(t *testing.T) {
	ctx := context.Background()
	directory := filepath.Join(t.TempDir(), "worker")
	store := printstate.Store{Directory: directory}
	lease, err := store.Acquire(ctx, "physical-device")
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	if _, err = store.Acquire(ctx, "physical-device"); !errors.Is(err, ports.ErrDeviceInUse) {
		t.Fatalf("competing lock: %v", err)
	}
	other, err := store.Acquire(ctx, "another-device")
	if err != nil {
		t.Fatal(err)
	}
	other.Close()
	if _, err = lease.Load(ctx); !errors.Is(err, ports.ErrJournalMissing) {
		t.Fatalf("missing is not idle: %v", err)
	}
	record := printing.JournalRecord{Binding: "registration", AttemptID: "attempt", JobID: "job", SessionID: "original-process", ClaimToken: "private-token", Phase: printing.JournalSubmitting, Copy: 1, CompletedCopies: 0}
	if err = lease.Save(ctx, &record); err != nil {
		t.Fatal(err)
	}
	if err = lease.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := store.Acquire(ctx, "physical-device")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	loaded, err := reopened.Load(ctx)
	if err != nil || loaded == nil || *loaded != record {
		t.Fatalf("lost durable evidence: %#v, %v", loaded, err)
	}
	if err = reopened.Save(ctx, nil); err != nil {
		t.Fatal(err)
	}
	idle, err := reopened.Load(ctx)
	if err != nil || idle != nil {
		t.Fatalf("initialized idle: %#v, %v", idle, err)
	}
}

func TestUnsafeOrCorruptJournalNeverLooksIdle(t *testing.T) {
	ctx := context.Background()
	directory := filepath.Join(t.TempDir(), "worker")
	store := printstate.Store{Directory: directory}
	lease, err := store.Acquire(ctx, "printer")
	if err != nil {
		t.Fatal(err)
	}
	if err = lease.Save(ctx, nil); err != nil {
		t.Fatal(err)
	}
	lease.Close()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			path := filepath.Join(directory, entry.Name())
			if err = os.WriteFile(path, []byte(`{"version":1,"device":"printer","record":null,"checksum":"wrong"}`), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	lease, err = store.Acquire(ctx, "printer")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = lease.Load(ctx); !errors.Is(err, ports.ErrJournalCorrupt) {
		t.Fatalf("corruption not rejected: %v", err)
	}
	lease.Close()
	if err = os.Chmod(directory, 0777); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Acquire(ctx, "printer"); err == nil {
		t.Fatal("world-writable journal accepted")
	}
}

package gormstore

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
)

func TestPostgresArchiveSnapshotAndJobPersistence(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if err = runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	tid := tenant.ID("archive-snapshot-tenant")
	iid := inventory.InventoryID("archive-snapshot-inventory")
	cleanup := func() {
		for _, m := range []any{&archiveJobModel{}, &inventoryModel{}, &tenantModel{}} {
			column := "tenant_id"
			if _, ok := m.(*tenantModel); ok {
				column = "id"
			}
			if err := db.Where(clause.Eq{Column: column, Value: tid.String()}).Delete(m).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	ctx := context.Background()
	store := NewStore(db)
	saveTenant(t, ctx, store, tid, "Archive")
	saveInventory(t, ctx, store, iid.String(), tid, "Before")
	err = store.WithArchiveSnapshot(ctx, tid, iid, func(repos ports.ArchiveSnapshotSources) error {
		original, found, err := repos.Inventories.InventoryByID(ctx, tid, iid)
		if err != nil {
			return err
		}
		if !found || original.Name.String() != "Before" {
			t.Fatal("missing initial inventory")
		}
		// The independent connection commits between reads. The captured snapshot
		// must retain the original name rather than mixing two database versions.
		saveInventory(t, ctx, store, iid.String(), tid, "After")
		again, found, err := repos.Inventories.InventoryByID(ctx, tid, iid)
		if err != nil {
			return err
		}
		if !found || again.Name != original.Name {
			t.Fatal("snapshot observed concurrent mutation")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	current, found, err := store.InventoryByID(ctx, tid, iid)
	if err != nil || !found || current.Name.String() != "After" {
		t.Fatalf("concurrent update missing: %v", err)
	}
	now := time.Date(2026, 10, 2, 0, 0, 0, 123456789, time.UTC)
	job, err := archivejob.New(archivejob.Request{RequestKey: "archive-test", ID: "archive-test", TenantID: tid.String(), PrincipalID: "owner", Kind: archivejob.Export, SourceInventoryID: iid.String()}, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.CreateArchiveJob(ctx, job); err != nil {
		t.Fatal(err)
	}
	read, found, err := store.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: tid.String(), SourceInventoryID: iid.String()}, job.ID)
	if err != nil || !found {
		t.Fatalf("timestamp roundtrip: %v", err)
	}
	claimed, err := read.Claim("worker", now.Add(time.Second), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := store.UpdateArchiveJob(ctx, claimed, read.Revision); err != nil || !ok {
		t.Fatalf("claim: %v", err)
	}
	if ok, err := store.UpdateArchiveJob(ctx, claimed, read.Revision); err != nil || ok {
		t.Fatalf("stale revision: %v", err)
	}
}

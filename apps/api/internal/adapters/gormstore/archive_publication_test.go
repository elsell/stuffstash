package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestArchivePublicationIsAtomicAndFenced(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(map[bool]string{false: "complete", true: "cancelled"}[cancelled], func(t *testing.T) {
			ctx := context.Background()
			s := newTestStore(t, ctx)
			saveTenant(t, ctx, s, "tenant", "Home")
			clock := &publicationClock{now: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)}
			job := createRestoreClaim(t, s, clock.now)
			if cancelled {
				next, err := job.Cancel(clock.now)
				if err != nil {
					t.Fatal(err)
				}
				if ok, err := s.UpdateArchiveJob(ctx, next, job.Revision); err != nil || !ok {
					t.Fatal(err)
				}
			}
			doc := ports.InventoryExportDocument{SchemaVersion: 2, TenantID: "tenant", InventoryID: "restored", InventoryName: "Restored", ExportedAt: clock.now, Assets: []ports.InventoryExportAsset{{ID: "child", Title: "Bottle", Kind: "item", ParentAssetID: "parent", LifecycleState: "archived", CreatedAt: clock.now, UpdatedAt: clock.now}, {ID: "parent", Title: "Shelf", Kind: "container", LifecycleState: "archived", CreatedAt: clock.now, UpdatedAt: clock.now}}}
			plan := ports.ArchiveRestorePlan{Document: doc}
			records, err := dataportability.BuildArchiveRestoreAudits(plan, job, &archiveAuditIDs{}, clock)
			if err != nil {
				t.Fatal(err)
			}
			request := ports.ArchiveRestorePublication{Job: job, Plan: plan, OwnerGrantEventID: "grant", AuditRecords: records}
			publisher := NewArchiveRestorePublisher(s, clock, 100)
			done, err := publisher.PublishArchiveRestore(ctx, request)
			_, found, readErr := s.InventoryByID(ctx, "tenant", "restored")
			if readErr != nil {
				t.Fatal(readErr)
			}
			if cancelled {
				if err == nil || found {
					t.Fatal("cancelled restore published")
				}
				return
			}
			if err != nil || !found || done.State != archivejob.Queued || done.Phase != archivejob.Finalization {
				t.Fatalf("publish: %v", err)
			}
			item, found, err := s.AssetByID(ctx, "tenant", "restored", "child")
			if err != nil || !found || item.ParentAssetID != "parent" || item.LifecycleState != "archived" {
				t.Fatalf("restored graph missing: %v", err)
			}
			events, err := s.ClaimPendingAuthorizationOutboxEvents(ctx, "grant-worker", 10, clock.now, clock.now.Add(time.Minute))
			if err != nil || len(events) != 1 || events[0].PrincipalID != "owner" {
				t.Fatalf("owner grant missing: %v", err)
			}
			if _, err = publisher.PublishArchiveRestore(ctx, request); err == nil {
				t.Fatal("replayed publication succeeded")
			}
		})
	}
}
func TestArchivePublicationFailureRollsBackInventory(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	saveTenant(t, ctx, s, "tenant", "Home")
	clock := &publicationClock{now: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)}
	job := createRestoreClaim(t, s, clock.now)
	// A preexisting event ID forces a late transaction failure after inserts.
	plan := ports.ArchiveRestorePlan{Document: ports.InventoryExportDocument{SchemaVersion: 2, TenantID: "tenant", InventoryID: "restored", InventoryName: "Restored", ExportedAt: clock.now}}
	records, err := dataportability.BuildArchiveRestoreAudits(plan, job, &archiveAuditIDs{}, clock)
	if err != nil {
		t.Fatal(err)
	}
	existing := records[0]
	existing.InventoryID = ""
	existing.TargetType = audit.TargetTenant
	existing.Action = audit.ActionTenantCreated
	existing.TargetID = "tenant"
	if err = s.SaveAuditRecord(ctx, existing); err != nil {
		t.Fatal(err)
	}
	request := ports.ArchiveRestorePublication{Job: job, Plan: plan, OwnerGrantEventID: "grant", AuditRecords: records}
	if _, err := NewArchiveRestorePublisher(s, clock, 100).PublishArchiveRestore(ctx, request); err == nil {
		t.Fatal("invalid audit accepted")
	}
	if _, found, err := s.InventoryByID(ctx, "tenant", "restored"); err != nil || found {
		t.Fatal("partial inventory escaped rollback")
	}
	read, found, err := s.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "tenant"}, job.ID)
	if err != nil || !found || read.State != archivejob.Running {
		t.Fatal("job completed despite rollback")
	}
}
func createRestoreClaim(t *testing.T, s Store, now time.Time) archivejob.Record {
	return createRestoreClaimFor(t, s, now, "tenant", "restored", "restore")
}
func createRestoreClaimFor(t *testing.T, s Store, now time.Time, tid, iid, id string) archivejob.Record {
	t.Helper()
	ctx := context.Background()
	job, err := archivejob.New(archivejob.Request{RequestKey: id, ID: id, TenantID: tid, PrincipalID: "owner", Kind: archivejob.Restore, SourceArtifactID: "upload", SourceSHA256: strings.Repeat("a", 64)}, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateArchiveJob(ctx, job); err != nil {
		t.Fatal(err)
	}
	persist := func(next archivejob.Record, err error) {
		if err != nil {
			t.Fatal(err)
		}
		if ok, err := s.UpdateArchiveJob(ctx, next, job.Revision); err != nil || !ok {
			t.Fatal(err)
		}
		job = next
	}
	persist(job.Claim("validate", now, now.Add(time.Minute)))
	persist(job.PreviewReady("validate", now, "plan", strings.Repeat("b", 64), iid))
	persist(job.Approve("Restored", now))
	persist(job.Claim("publish", now, now.Add(time.Minute)))
	return job
}

type archiveAuditIDs struct{ n int }

func (g *archiveAuditIDs) NewID() string { g.n++; return "restore-audit-" + strconv.Itoa(g.n) }

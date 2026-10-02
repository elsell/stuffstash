package gormstore

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
)

func TestArchiveJobAuditsAreAtomicAndIdempotent(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	saveTenant(t, ctx, s, "tenant", "Home")
	saveInventory(t, ctx, s, "inventory", "tenant", "Inventory")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	job, err := archivejob.New(archivejob.Request{RequestKey: "request", ID: "job", TenantID: "tenant", PrincipalID: "owner", Kind: archivejob.Export, SourceInventoryID: "inventory"}, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	record := auditRecord(t, "job-created", "tenant", "inventory", audit.ActionArchiveJobCreated)
	record.PrincipalID = "owner"
	record.TargetType = audit.TargetArchiveJob
	record.TargetID = job.ID
	if _, err = s.CreateArchiveJobAudited(ctx, job, record); err != nil {
		t.Fatal(err)
	}
	duplicate := job
	duplicate.ID = "duplicate"
	retryAudit := record
	retryAudit.ID = "retry-audit"
	retryAudit.TargetID = duplicate.ID
	if reused, err := s.CreateArchiveJobAudited(ctx, duplicate, retryAudit); err != nil || reused.ID != job.ID {
		t.Fatalf("duplicate request: %v", err)
	}
	var count int64
	if err = s.db.Model(&auditRecordModel{}).Where(clause.Eq{Column: "target_type", Value: audit.TargetArchiveJob.String()}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("duplicated creation audit: %d %v", count, err)
	}
	next, err := job.Cancel(now)
	if err != nil {
		t.Fatal(err)
	}
	// Colliding audit ID fails after the job update and must roll it back.
	changedAudit := record
	changedAudit.Action = audit.ActionArchiveJobUpdated
	if ok, err := s.UpdateArchiveJobAudited(ctx, next, job.Revision, changedAudit); err == nil || ok {
		t.Fatal("audit failure committed cancellation")
	}
	retained, found, err := s.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: "tenant", SourceInventoryID: "inventory"}, job.ID)
	if err != nil || !found || retained.State != archivejob.Queued {
		t.Fatal("job escaped transaction rollback")
	}
	changedAudit.ID = "job-cancelled"
	if ok, err := s.UpdateArchiveJobAudited(ctx, next, job.Revision, changedAudit); err != nil || !ok {
		t.Fatalf("audited cancellation: %v", err)
	}
}

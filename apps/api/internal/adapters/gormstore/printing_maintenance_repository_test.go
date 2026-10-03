package gormstore

import (
	"context"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestPrintMaintenanceRetainsUncertainOutputAndRemovesOnlyTerminalHistory(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	saveTenant(t, ctx, s, "tenant", "Home")
	saveInventory(t, ctx, s, "inventory", "tenant", "Garage")
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	statuses := []printing.JobStatus{printing.JobQueued, printing.JobClaimed, printing.JobPrinting, printing.JobUncertain, printing.JobCompleted}
	for i, status := range statuses {
		id := fmt.Sprint(i)
		p, _ := printingPrinterFromDomain(printing.Printer{ID: printing.PrinterID(id), Scope: scope, RequestKey: id, RequestFingerprint: id, Revision: 1, MediaFingerprint: "media", ActiveJobID: id, ReservationState: string(status)})
		if err := s.db.Create(&p).Error; err != nil {
			t.Fatal(err)
		}
		j := printing.Job{ID: printing.JobID(id), Scope: scope, PrinterID: printing.PrinterID(id), Status: status, Revision: 1, RequestedBy: "owner", IdempotencyKey: id, UpdatedAt: now.Add(-48 * time.Hour), Artifact: printing.Artifact{ExpiresAt: now.Add(-time.Hour)}, Attempts: []printing.Attempt{{ID: printing.AttemptID("attempt" + id), LeaseExpiresAt: now.Add(-time.Minute)}}}
		row, err := printJobModel(j, id, []byte("private-artifact"))
		if err != nil {
			t.Fatal(err)
		}
		if err = s.db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	in := ports.PrintJobMaintenance{Now: now, TerminalBefore: now.Add(-24 * time.Hour), Limit: 100, Audit: func(before, after printing.Job) (audit.Record, error) {
		r := auditRecord(t, "maintain-"+string(after.ID), "tenant", "inventory", audit.ActionPrintJobFailed)
		r.TargetID = string(after.ID)
		return r, nil
	}}
	page, err := s.MaintainPrintJobs(ctx, in)
	if err != nil || page.HasMore {
		t.Fatalf("maintenance: %+v %v", page, err)
	}
	for _, id := range []string{"0", "1", "2", "3"} {
		row, err := printJobByID(s.db, scope, printing.JobID(id))
		if err != nil {
			t.Fatal(err)
		}
		j, _ := row.domain()
		if id == "2" || id == "3" {
			if j.Status != printing.JobUncertain || len(row.ArtifactContent) == 0 {
				t.Fatalf("lost uncertain evidence: %+v", j)
			}
		} else if j.Status != printing.JobFailed || len(row.ArtifactContent) != 0 {
			t.Fatalf("expired job still printable: %+v", j)
		}
	}
	if _, err = printJobByID(s.db, scope, "4"); err == nil {
		t.Fatal("expired terminal metadata retained")
	}
	// A later maintenance pass must never age out unresolved physical output.
	in.Now = now.Add(365 * 24 * time.Hour)
	in.TerminalBefore = in.Now.Add(-24 * time.Hour)
	if _, err = s.MaintainPrintJobs(ctx, in); err != nil {
		t.Fatal(err)
	}
	for _, id := range []printing.JobID{"2", "3"} {
		if _, err = printJobByID(s.db, scope, id); err != nil {
			t.Fatal("uncertain reservation deleted", err)
		}
	}
}

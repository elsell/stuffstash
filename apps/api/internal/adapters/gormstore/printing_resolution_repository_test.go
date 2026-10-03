package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestResolutionReleasesReservationOnlyWithCommittedAudit(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	now := time.Now().UTC()
	saveTenant(t, ctx, s, "tenant", "Home")
	saveInventory(t, ctx, s, "inventory", "tenant", "Garage")
	scope := p.Scope{TenantID: "tenant", InventoryID: "inventory"}
	printer, _ := printingPrinterFromDomain(p.Printer{ID: "printer", Scope: scope, RequestKey: "printer", RequestFingerprint: "printer", Revision: 1, ActiveJobID: "job", ReservationState: string(p.JobUncertain)})
	if err := s.db.Create(&printer).Error; err != nil {
		t.Fatal(err)
	}
	job := p.Job{ID: "job", Scope: scope, PrinterID: "printer", Status: p.JobUncertain, Revision: 1, RequestedBy: "owner", IdempotencyKey: "job", Attempts: []p.Attempt{{ID: "attempt", IdleConfirmedAt: now, Outcome: p.Outcome{Kind: p.OutcomeUncertain, Reason: p.ReasonUnknown}}}}
	row, err := printJobModel(job, "job", []byte("artifact"))
	if err != nil {
		t.Fatal(err)
	}
	if err = s.db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	update := ports.PrintJobUpdate{Scope: scope, PrinterID: "printer", JobID: "job", Now: now, Change: func(j *p.Job, _ p.Printer) error { return j.Resolve("owner", p.ReportedUnknown, true, now, 1) }, Audit: func(before, after p.Job) (audit.Record, error) {
		return audit.Record{}, errors.New("audit unavailable")
	}}
	if _, err = s.UpdatePrintJob(ctx, update); err == nil {
		t.Fatal("missing audit accepted")
	}
	unchanged, err := s.GetPrintJob(ctx, scope, "job")
	if err != nil || unchanged.Status != p.JobUncertain || unchanged.Resolution != nil {
		t.Fatal("failed audit released job", err)
	}
	currentPrinter, err := s.GetPrinter(ctx, scope, "printer")
	if err != nil || currentPrinter.ActiveJobID != "job" {
		t.Fatal("failed audit released reservation", err)
	}
	update.Audit = func(before, after p.Job) (audit.Record, error) {
		r := auditRecord(t, "resolved", "tenant", "inventory", audit.ActionPrintJobResolved)
		r.TargetID = "job"
		return r, nil
	}
	resolved, err := s.UpdatePrintJob(ctx, update)
	if err != nil || resolved.Resolution == nil || resolved.Attempts[0].Outcome.Kind != p.OutcomeUncertain {
		t.Fatal("resolution lost evidence", err)
	}
	currentPrinter, err = s.GetPrinter(ctx, scope, "printer")
	if err != nil || currentPrinter.ActiveJobID != "" {
		t.Fatal("committed resolution did not release", err)
	}
}

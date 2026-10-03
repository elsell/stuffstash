package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestRetirementAtomicallyCancelsOnlyUnstartedOutputAndClearsDefault(t *testing.T) {
	for _, activeStatus := range []printing.JobStatus{printing.JobClaimed, printing.JobPrinting, printing.JobUncertain} {
		t.Run(string(activeStatus), func(t *testing.T) {
			ctx := context.Background()
			store := newTestStore(t, ctx)
			saveTenant(t, ctx, store, "tenant", "Home")
			saveInventory(t, ctx, store, "inventory", "tenant", "Garage")
			scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
			now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
			printer := printing.Printer{ID: "printer", Scope: scope, Revision: 1, RequestKey: "p", RequestFingerprint: "p", ActiveJobID: "active", ReservationState: string(activeStatus)}
			auditFor := func(id string, action audit.Action, target string) audit.Record {
				r := auditRecord(t, id, "tenant", "inventory", action)
				r.TargetID = target
				return r
			}
			if _, _, err := store.CreatePrinter(ctx, printer, auditFor("create", audit.ActionPrinterRegistered, "printer")); err != nil {
				t.Fatal(err)
			}
			for _, row := range []struct {
				id     printing.JobID
				status printing.JobStatus
			}{{"queued", printing.JobQueued}, {"active", activeStatus}} {
				job := printing.Job{ID: row.id, Scope: scope, PrinterID: printer.ID, Revision: 1, Status: row.status, RequestedBy: "human", IdempotencyKey: string(row.id)}
				model, err := printJobModel(job, string(row.id), nil)
				if err != nil {
					t.Fatal(err)
				}
				if err = store.db.Create(&model).Error; err != nil {
					t.Fatal(err)
				}
			}
			settings := printing.DefaultInventoryPrintSettings(scope)
			settings.DefaultPrinterID = printer.ID
			settings.PrintOnCreateDefault = true
			settings.Revision = 1
			model := settingsModel(settings)
			if err := store.db.Create(&model).Error; err != nil {
				t.Fatal(err)
			}
			effects := &ports.PrinterRetirement{Now: now, JobAudit: func(_, j printing.Job) (audit.Record, error) {
				return auditFor("cancel-"+string(j.ID), audit.ActionPrintJobCanceled, string(j.ID)), nil
			}, SettingsAudit: func(printing.InventoryPrintSettings) (audit.Record, error) {
				return auditFor("create", audit.ActionPrintSettingsUpdated, "inventory"), nil
			}}
			mutation := func(p *printing.Printer) error { p.Retired = true; p.Revision++; return nil }
			printerAudit := func(printing.Printer) (audit.Record, error) {
				return auditFor("retire", audit.ActionPrinterUpdated, "printer"), nil
			}
			if _, err := store.UpdatePrinter(ctx, scope, printer.ID, 1, mutation, printerAudit, effects); err == nil {
				t.Fatal("duplicate settings audit did not abort transaction")
			}
			unchanged, _ := store.GetPrinter(ctx, scope, printer.ID)
			queued, _ := store.GetPrintJob(ctx, scope, "queued")
			defaults, _ := store.GetPrintSettings(ctx, scope)
			if unchanged.Retired || queued.Status != printing.JobQueued || defaults.DefaultPrinterID != printer.ID {
				t.Fatal("failed retirement partially committed")
			}
			effects.SettingsAudit = func(printing.InventoryPrintSettings) (audit.Record, error) {
				return auditFor("clear-default", audit.ActionPrintSettingsUpdated, "inventory"), nil
			}
			retired, err := store.UpdatePrinter(ctx, scope, printer.ID, 1, mutation, printerAudit, effects)
			if err != nil {
				t.Fatal(err)
			}
			queued, _ = store.GetPrintJob(ctx, scope, "queued")
			active, _ := store.GetPrintJob(ctx, scope, "active")
			defaults, _ = store.GetPrintSettings(ctx, scope)
			if queued.Status != printing.JobCanceled || defaults.DefaultPrinterID != "" || defaults.PrintOnCreateDefault || defaults.Revision != 2 {
				t.Fatal("retirement did not settle pending work/default")
			}
			if activeStatus == printing.JobClaimed {
				if active.Status != printing.JobCanceled || retired.ActiveJobID != "" {
					t.Fatal("unstarted reservation not released")
				}
			} else if active.Status != activeStatus || retired.ActiveJobID != "active" {
				t.Fatal("possible physical output was discarded")
			}
		})
	}
}

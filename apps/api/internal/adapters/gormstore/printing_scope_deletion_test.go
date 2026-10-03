package gormstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestInventoryDeletionRetainsPrintingEvidenceAndPendingRevocation(t *testing.T) {
	exerciseInventoryPrintingDeletion(t, newTestStore(t, context.Background()))
}

func exerciseInventoryPrintingDeletion(t *testing.T, store Store) {
	ctx := context.Background()
	saveTenant(t, ctx, store, "tenant", "Home")
	saveInventory(t, ctx, store, "inventory", "tenant", "Garage")
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	now := time.Now().UTC()
	printer := printing.Printer{ID: "printer", Scope: scope, Revision: 1, RequestKey: "printer", RequestFingerprint: "printer", Readiness: printing.PrinterUnknown, ActiveJobID: "started", ReservationState: string(printing.JobPrinting)}
	if _, _, err := store.CreatePrinter(ctx, printer, auditRecord(t, "printer-created", "tenant", "inventory", audit.ActionPrinterRegistered)); err != nil {
		t.Fatal(err)
	}
	connector := printingConnectorFromDomain(printing.Connector{ID: "connector", Scope: scope, ServiceAccountID: "service", PublicKey: make([]byte, 32), State: printing.ConnectorActive, CredentialHash: "old", CredentialVersion: 1, Generation: 1, SyncedGeneration: 1, CredentialExpiresAt: now.Add(time.Hour)})
	binding := printingBindingFromDomain(printing.PrinterBinding{Scope: scope, PrinterID: printer.ID, ConnectorID: "connector", Generation: 1, SyncedGeneration: 1})
	for _, row := range []any{&connector, &binding} {
		if err := store.db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, status := range []printing.JobStatus{printing.JobQueued, printing.JobPrinting} {
		id := printing.JobID("queued")
		if status == printing.JobPrinting {
			id = "started"
		}
		job := printing.Job{ID: id, Scope: scope, PrinterID: printer.ID, Status: status, Revision: 1, RequestedBy: "human", IdempotencyKey: string(id), Copies: 2}
		if status == printing.JobPrinting {
			job.Attempts = []printing.Attempt{{Authority: printing.AttemptAuthority{AttemptID: "attempt", ConnectorID: "connector"}, Outcome: printing.Outcome{CompletedCopies: 1}}}
		}
		row, err := printJobModel(job, string(id), []byte("retained evidence"))
		if err != nil {
			t.Fatal(err)
		}
		if err = store.db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	settings := printing.DefaultInventoryPrintSettings(scope)
	settings.Revision = 1
	settings.DefaultPrinterID = printer.ID
	settings.PrintOnCreateDefault = true
	if _, err := store.SavePrintSettings(ctx, settings, 0, &printing.SettingsDestination{ID: printer.ID, Revision: 1}, auditRecord(t, "settings-created", "tenant", "inventory", audit.ActionPrintSettingsUpdated)); err != nil {
		t.Fatal(err)
	}
	deleted := auditRecord(t, "deleted", "tenant", "inventory", audit.ActionInventoryDeleted)
	effects := &ports.InventoryDeletionEffects{Audit: func(action audit.Action, target audit.TargetType, id string) (audit.Record, error) {
		r := auditRecord(t, "printer-created", "tenant", "inventory", action)
		r.TargetID = id
		r.TargetType = target
		return r, nil
	}}
	if err := store.DeleteInventory(ctx, "tenant", "inventory", deleted, effects); err == nil {
		t.Fatal("audit failure did not roll back deletion")
	}
	if _, found, _ := store.InventoryByID(ctx, "tenant", "inventory"); !found {
		t.Fatal("failed deletion removed inventory")
	}
	unchanged, _ := store.GetPrintConnector(ctx, scope, "connector")
	if unchanged.Connector.State != printing.ConnectorActive {
		t.Fatal("failed deletion revoked credential")
	}
	sequence := 0
	effects.Audit = func(action audit.Action, target audit.TargetType, id string) (audit.Record, error) {
		sequence++
		r := auditRecord(t, fmt.Sprintf("deletion-%d", sequence), "tenant", "inventory", action)
		r.TargetID = id
		r.TargetType = target
		return r, nil
	}
	var registration <-chan error
	if store.db.Dialector.Name() == "postgres" {
		result := make(chan error, 1)
		registration = result
		makeAudit := effects.Audit
		launched := false
		effects.Audit = func(action audit.Action, target audit.TargetType, id string) (audit.Record, error) {
			if !launched {
				launched = true
				started := make(chan struct{})
				go func() {
					close(started)
					late := printer
					late.ID = "late-printer"
					late.RequestKey = "late"
					_, _, err := store.CreatePrinter(ctx, late, auditRecord(t, "late-created", "tenant", "inventory", audit.ActionPrinterRegistered))
					result <- err
				}()
				<-started
			}
			return makeAudit(action, target, id)
		}
	}
	if err := store.DeleteInventory(ctx, "tenant", "inventory", deleted, effects); err != nil {
		t.Fatal(err)
	}
	if registration != nil {
		select {
		case err := <-registration:
			if !errors.Is(err, ports.ErrPrintNotFound) {
				t.Fatalf("concurrent registration resurrected deleted scope: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("registration/deletion lock order deadlocked")
		}
	}
	cleared, _ := store.GetPrintSettings(ctx, scope)
	if cleared.DefaultPrinterID != "" || cleared.PrintOnCreateDefault {
		t.Fatal("deleted scope retained active defaults")
	}

	c, err := store.GetPrintConnector(ctx, scope, "connector")
	if err != nil || c.Connector.State != printing.ConnectorRevoked || c.Connector.CredentialHash != "" || c.Connector.Generation == c.Connector.SyncedGeneration || !c.Bindings[0].Revoked {
		t.Fatalf("revocation not retained: %+v %v", c, err)
	}
	queued, _ := store.GetPrintJob(ctx, scope, "queued")
	started, _ := store.GetPrintJob(ctx, scope, "started")
	if queued.Status != printing.JobCanceled || started.Status != printing.JobUncertain || started.Attempts[0].Outcome.CompletedCopies != 1 {
		t.Fatalf("physical evidence lost: %+v %+v", queued, started)
	}
	pending, err := store.PendingPrintConnectorScopes(ctx, 10)
	if err != nil || len(pending) != 1 {
		t.Fatal("deleted scope lost pending relationship removal")
	}
	if err = store.DeleteTenant(ctx, "tenant", auditRecord(t, "tenant-deleted", "tenant", "", audit.ActionTenantDeleted)); err != nil {
		t.Fatal(err)
	}
	if _, err = store.GetPrintJob(ctx, scope, "started"); err != nil {
		t.Fatal("tenant deletion purged uncertain output")
	}
}

package memory

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestPrintingQueueFakeAtomicallyReservesPrinterAcrossJobs(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	s := NewStore()
	s.printingPrinters = map[printing.PrinterID]printing.Printer{"printer": {ID: "printer", Scope: scope, Revision: 1, MediaFingerprint: "media"}}
	// Registry state is real in-memory state. No scripted call expectations.
	s.printingConnectors = map[printing.ConnectorID]printing.Connector{"connector": {ID: "connector", Scope: scope, ServiceAccountID: "worker", State: printing.ConnectorActive, CredentialVersion: 1, LastSeenAt: &now, CredentialExpiresAt: now.Add(time.Hour), Generation: 1, SyncedGeneration: 1}}
	s.printingBindings = map[string]printing.PrinterBinding{"connector:printer": {Scope: scope, ConnectorID: "connector", PrinterID: "printer", Generation: 1, SyncedGeneration: 1}}
	s.printingReports = map[string]printing.PrinterReport{"connector:printer": {Scope: scope, ConnectorID: "connector", PrinterID: "printer", State: printing.PrinterReady, ReportedAt: now}}
	auditFor := func(before, after printing.Job) (audit.Record, error) {
		return audit.Record{ID: audit.ID(fmt.Sprintf("%s-%d", after.ID, after.Revision)), TenantID: "tenant", InventoryID: "inventory", TargetID: string(after.ID)}, nil
	}
	for i := 0; i < 2; i++ {
		job := printing.Job{ID: printing.JobID(fmt.Sprint(i)), Scope: scope, PrinterID: "printer", Status: printing.JobQueued, Revision: 1, Copies: 1, MediaFingerprint: "media", IdempotencyKey: fmt.Sprint(i), RequestedBy: "human", CreatedAt: now}
		rec, _ := auditFor(printing.Job{}, job)
		if _, _, err := s.CreatePrintJob(ctx, ports.PrintJobCreate{Job: job, PrinterRevision: 1, RequestFingerprint: fmt.Sprint(i), Audit: rec}); err != nil {
			t.Fatal(err)
		}
	}
	authority := printing.ConsumerAuthority{Scope: scope, ConnectorID: "connector", ServiceAccountID: "worker", CredentialVersion: 1, PrinterID: "printer", BindingGeneration: 1}
	var wg sync.WaitGroup
	results := make(chan printing.Job, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := printing.AttemptID(fmt.Sprint(i))
			job, found, err := s.ClaimPrintJob(ctx, ports.PrintClaim{Authority: authority, Owner: printing.AttemptAuthority{AttemptID: id, ConnectorID: "connector", SessionID: printing.SessionID(id), TokenDigest: sha256.Sum256([]byte(id))}, Now: now, Lease: time.Minute, ReportMaxAge: time.Minute, Audit: auditFor})
			if err != nil {
				t.Error(err)
			}
			if found {
				results <- job
			}
		}(i)
	}
	wg.Wait()
	close(results)
	if len(results) != 1 {
		t.Fatalf("claimed %d jobs for one physical printer", len(results))
	}
	claimed := <-results
	owner := claimed.Attempts[0].Authority
	started, err := s.UpdatePrintJob(ctx, ports.PrintJobUpdate{Scope: scope, PrinterID: "printer", JobID: claimed.ID, Authority: &authority, Now: now, Change: func(j *printing.Job, p printing.Printer) error { return j.Start(owner, now, j.Revision) }, Audit: auditFor})
	if err != nil {
		t.Fatal(err)
	}
	_, found, err := s.ClaimPrintJob(ctx, ports.PrintClaim{Authority: authority, Owner: printing.AttemptAuthority{AttemptID: "later", ConnectorID: "connector", SessionID: "new", TokenDigest: sha256.Sum256([]byte("new"))}, Now: now.Add(time.Minute), Lease: time.Minute, ReportMaxAge: time.Hour, Audit: auditFor})
	if err != nil || found {
		t.Fatal("expired started attempt must hold printer, not dispatch another job")
	}
	stored, err := s.GetPrintJob(ctx, scope, started.ID)
	if err != nil || stored.Status != printing.JobUncertain {
		t.Fatal("expiry must persist uncertainty")
	}
	if _, err := s.GetPrintJob(ctx, printing.Scope{TenantID: "other", InventoryID: "inventory"}, started.ID); err == nil {
		t.Fatal("cross tenant read succeeded")
	}
}

package gormstore

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestPrintingQueuePersistsIdempotencyAndRollsBackAuditFailure(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	if err := s.db.AutoMigrate(&printingPrinterModel{}, &printingJobModel{}); err != nil {
		t.Fatal(err)
	}
	saveTenant(t, ctx, s, "tenant", "Home")
	saveInventory(t, ctx, s, "inventory", "tenant", "Garage")
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	p, err := printingPrinterFromDomain(printing.Printer{ID: "printer", Scope: scope, Revision: 1, MediaFingerprint: "media"})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	job := printing.Job{ID: "job", Scope: scope, PrinterID: "printer", Status: printing.JobQueued, Revision: 1, Copies: 1, MediaFingerprint: "media", RequestedBy: "human", IdempotencyKey: "request", CreatedAt: time.Now().UTC()}
	record := auditRecord(t, "queue-created", "tenant", "inventory", audit.ActionAssetCreated)
	record.TargetID = "job"
	input := ports.PrintJobCreate{Job: job, PrinterRevision: 1, RequestFingerprint: "first", Audit: record}
	if _, created, err := s.CreatePrintJob(ctx, input); err != nil || !created {
		t.Fatalf("create: %v %v", created, err)
	}
	// Replay survives a restart and later printer configuration changes.
	s = NewStore(s.db)
	if err = s.db.Model(&printingPrinterModel{}).Where("id = ?", "printer").Update("retired", true).Error; err != nil {
		t.Fatal(err)
	}
	if got, created, err := s.CreatePrintJob(ctx, input); err != nil || created || got.ID != job.ID {
		t.Fatalf("replay: %v %v", created, err)
	}
	input.RequestFingerprint = "different"
	if _, _, err := s.CreatePrintJob(ctx, input); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("changed replay: %v", err)
	}
	if _, err := s.GetPrintJob(ctx, printing.Scope{TenantID: "other", InventoryID: "inventory"}, job.ID); !errors.Is(err, ports.ErrPrintJobNotFound) {
		t.Fatalf("cross tenant: %v", err)
	}
	if err = s.db.Model(&printingPrinterModel{}).Where("id = ?", "printer").Update("retired", false).Error; err != nil {
		t.Fatal(err)
	}
	input.Job.ID = "second"
	input.Job.IdempotencyKey = "second"
	input.Audit.TargetID = "second"
	if _, _, err := s.CreatePrintJob(ctx, input); err == nil {
		t.Fatal("duplicate audit should reject mutation")
	}
	if _, err := s.GetPrintJob(ctx, scope, "second"); !errors.Is(err, ports.ErrPrintJobNotFound) {
		t.Fatalf("audit failure leaked job: %v", err)
	}
}

func TestPrintingQueueDurableClaimExpiryAndPrinterHold(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	if err := s.db.AutoMigrate(&printingPrinterModel{}, &printingConnectorModel{}, &printingBindingModel{}, &printingReportModel{}, &printingJobModel{}, &printingAttemptIndex{}); err != nil {
		t.Fatal(err)
	}
	saveTenant(t, ctx, s, "tenant", "Home")
	saveInventory(t, ctx, s, "inventory", "tenant", "Garage")
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	p, err := printingPrinterFromDomain(printing.Printer{ID: "printer", Scope: scope, Revision: 1, MediaFingerprint: "media"})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []any{&p, &printingConnectorModel{ID: "connector", TenantID: "tenant", InventoryID: "inventory", ServiceAccountID: "worker", State: string(printing.ConnectorActive), CredentialVersion: 1, CredentialExpiresAt: now.Add(time.Hour), LastSeenAt: &now, Generation: 1, SyncedGeneration: 1}, &printingBindingModel{ConnectorID: "connector", PrinterID: "printer", TenantID: "tenant", InventoryID: "inventory", Generation: 1, SyncedGeneration: 1}, &printingReportModel{ConnectorID: "connector", PrinterID: "printer", TenantID: "tenant", InventoryID: "inventory", State: string(printing.PrinterReady), ReportedAt: now}} {
		if err = s.db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	auditFor := func(before, after printing.Job) (audit.Record, error) {
		r := auditRecord(t, fmt.Sprintf("%s-%d", after.ID, after.Revision), "tenant", "inventory", audit.ActionAssetCreated)
		r.TargetID = string(after.ID)
		return r, nil
	}
	for _, id := range []printing.JobID{"first", "second"} {
		j := printing.Job{ID: id, Scope: scope, PrinterID: "printer", Status: printing.JobQueued, Revision: 1, Copies: 1, MediaFingerprint: "media", IdempotencyKey: string(id), RequestedBy: "human", CreatedAt: now}
		r, _ := auditFor(printing.Job{}, j)
		if _, _, err = s.CreatePrintJob(ctx, ports.PrintJobCreate{Job: j, PrinterRevision: 1, RequestFingerprint: string(id), Audit: r}); err != nil {
			t.Fatal(err)
		}
	}
	authority := printing.ConsumerAuthority{Scope: scope, ConnectorID: "connector", ServiceAccountID: "worker", CredentialVersion: 1, PrinterID: "printer", BindingGeneration: 1}
	owner := printing.AttemptAuthority{AttemptID: "attempt", ConnectorID: "connector", SessionID: "session", TokenDigest: sha256.Sum256([]byte("claim"))}
	input := ports.PrintClaim{Authority: authority, Owner: owner, Now: now, Lease: time.Minute, ReportMaxAge: time.Hour, Audit: auditFor}
	job, found, err := s.ClaimPrintJob(ctx, input)
	if err != nil || !found {
		t.Fatalf("claim: %v %v", found, err)
	}
	s = NewStore(s.db)
	if recovered, err := s.FindPrintAttempt(ctx, scope, "connector", owner.AttemptID); err != nil || recovered.ID != job.ID {
		t.Fatalf("lost response recovery: %v", err)
	}
	if _, err = s.FindPrintAttempt(ctx, scope, "other", owner.AttemptID); !errors.Is(err, ports.ErrPrintJobNotFound) {
		t.Fatal("attempt leaked to other connector")
	}
	input.Owner.AttemptID = "competing"
	if _, found, err = s.ClaimPrintJob(ctx, input); err != nil || found {
		t.Fatalf("printer reservation lost: %v %v", found, err)
	}
	if _, err = s.UpdatePrintJob(ctx, ports.PrintJobUpdate{Scope: scope, PrinterID: "printer", JobID: job.ID, Authority: &authority, Now: now, Change: func(j *printing.Job, p printing.Printer) error { return j.Start(owner, now, j.Revision) }, Audit: auditFor}); err != nil {
		t.Fatal(err)
	}
	input.Now = now.Add(time.Minute)
	if _, found, err = s.ClaimPrintJob(ctx, input); err != nil || found {
		t.Fatalf("uncertain printer redispatched: %v %v", found, err)
	}
	stored, err := s.GetPrintJob(ctx, scope, job.ID)
	if err != nil || stored.Status != printing.JobUncertain {
		t.Fatalf("uncertainty not durable: %v", err)
	}
	if err = s.db.Where("id = ?", "printer").First(&p).Error; err != nil || p.ActiveJobID != string(job.ID) {
		t.Fatalf("uncertain hold lost: %v", err)
	}
}

package gormstore

import (
	"context"
	"crypto/sha256"
	"gorm.io/gorm/clause"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestPostgresPrintingClaimsSerializeAcrossConnectors(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err = runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	s := NewStore(db)
	ctx := context.Background()
	ids := idgen.ULIDGenerator{}
	now := time.Now().UTC()
	scope := printing.Scope{TenantID: ids.NewID(), InventoryID: ids.NewID()}
	pid := printing.PrinterID(ids.NewID())
	saveTenant(t, ctx, s, tenant.ID(scope.TenantID), "Printing concurrency")
	saveInventory(t, ctx, s, scope.InventoryID, tenant.ID(scope.TenantID), "Inventory")
	t.Cleanup(func() {
		for _, m := range []any{&printingAttemptIndex{}, &printingJobModel{}, &printingReportModel{}, &printingBindingModel{}, &printingConnectorModel{}, &printingPrinterModel{}, &auditRecordModel{}, &inventoryModel{}} {
			if e := db.Where(clause.Eq{Column: "tenant_id", Value: scope.TenantID}).Delete(m).Error; e != nil {
				t.Error(e)
			}
		}
		if e := db.Where(clause.Eq{Column: "id", Value: scope.TenantID}).Delete(&tenantModel{}).Error; e != nil {
			t.Error(e)
		}
	})
	p, e := printingPrinterFromDomain(printing.Printer{ID: pid, Scope: scope, Name: "Printer", AdapterID: "brother-ql800", RequestKey: ids.NewID(), RequestFingerprint: "registration", MediaFingerprint: "media", Revision: 1, Readiness: printing.PrinterReady, CreatedAt: now, UpdatedAt: now})
	if e != nil {
		t.Fatal(e)
	}
	if err = db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	authorities := make([]printing.ConsumerAuthority, 2)
	for i := range authorities {
		cid := printing.ConnectorID(ids.NewID())
		sid := printing.ServiceAccountID(ids.NewID())
		authorities[i] = printing.ConsumerAuthority{Scope: scope, ConnectorID: cid, ServiceAccountID: sid, CredentialVersion: 1, PrinterID: pid, BindingGeneration: 1}
		for _, row := range []any{&printingConnectorModel{ID: string(cid), TenantID: scope.TenantID, InventoryID: scope.InventoryID, ServiceAccountID: string(sid), Name: "Consumer", State: string(printing.ConnectorActive), PublicKey: []byte{}, CredentialVersion: 1, CredentialExpiresAt: now.Add(time.Hour), ActivationDeadline: now, LastSeenAt: &now, Generation: 1, SyncedGeneration: 1, CreatedAt: now, UpdatedAt: now}, &printingBindingModel{ConnectorID: string(cid), PrinterID: string(pid), TenantID: scope.TenantID, InventoryID: scope.InventoryID, DeviceID: string(cid), Generation: 1, SyncedGeneration: 1}, &printingReportModel{ConnectorID: string(cid), PrinterID: string(pid), TenantID: scope.TenantID, InventoryID: scope.InventoryID, State: string(printing.PrinterReady), ReportedAt: now}} {
			if err = db.Create(row).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	auditFor := func(before, after printing.Job) (audit.Record, error) {
		r := auditRecord(t, ids.NewID(), tenant.ID(scope.TenantID), inventory.InventoryID(scope.InventoryID), audit.ActionAssetCreated)
		r.TargetID = string(after.ID)
		return r, nil
	}
	for range 2 {
		j := printing.Job{ID: printing.JobID(ids.NewID()), Scope: scope, PrinterID: pid, Status: printing.JobQueued, Revision: 1, Copies: 1, MediaFingerprint: "media", RequestedBy: "human", IdempotencyKey: ids.NewID(), CreatedAt: now, UpdatedAt: now}
		r, _ := auditFor(printing.Job{}, j)
		if _, _, err = s.CreatePrintJob(ctx, ports.PrintJobCreate{Job: j, PrinterRevision: 1, RequestFingerprint: string(j.ID), Audit: r}); err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	results := make(chan printing.Job, 2)
	var wg sync.WaitGroup
	for _, authority := range authorities {
		wg.Add(1)
		go func(a printing.ConsumerAuthority) {
			defer wg.Done()
			<-start
			owner := printing.AttemptAuthority{AttemptID: printing.AttemptID(ids.NewID()), ConnectorID: a.ConnectorID, SessionID: printing.SessionID(ids.NewID()), TokenDigest: sha256.Sum256([]byte(ids.NewID()))}
			j, found, e := s.ClaimPrintJob(ctx, ports.PrintClaim{Authority: a, Owner: owner, Now: now, Lease: time.Minute, ReportMaxAge: time.Minute, Audit: auditFor})
			if e != nil {
				t.Error(e)
			}
			if found {
				results <- j
			}
		}(authority)
	}
	close(start)
	wg.Wait()
	close(results)
	if len(results) != 1 {
		t.Fatalf("expected one printer reservation, got %d", len(results))
	}
	job := <-results
	owner := job.Attempts[0].Authority
	var authority printing.ConsumerAuthority
	for _, a := range authorities {
		if a.ConnectorID == owner.ConnectorID {
			authority = a
		}
	}
	if _, err = s.UpdatePrintJob(ctx, ports.PrintJobUpdate{Scope: scope, PrinterID: pid, JobID: job.ID, Authority: &authority, Now: now, Change: func(j *printing.Job, p printing.Printer) error { return j.Start(owner, now, j.Revision) }, Audit: auditFor}); err != nil {
		t.Fatal(err)
	}
	owner.AttemptID = printing.AttemptID(ids.NewID())
	if _, found, e := s.ClaimPrintJob(ctx, ports.PrintClaim{Authority: authority, Owner: owner, Now: now.Add(time.Minute), Lease: time.Minute, ReportMaxAge: time.Hour, Audit: auditFor}); e != nil || found {
		t.Fatalf("uncertain output must retain printer: %v %v", found, e)
	}
	stored, e := s.GetPrintJob(ctx, scope, job.ID)
	if e != nil || stored.Status != printing.JobUncertain {
		t.Fatalf("uncertainty lost: %v", e)
	}
}

package gormstore

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestArchiveJobsPersistAndFenceStaleWorkers(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	saveTenant(t, ctx, s, tenant.ID("tenant"), "Home")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	request := archivejob.Request{RequestKey: "request", ID: "job", TenantID: "tenant", PrincipalID: "owner", Kind: archivejob.Export, SourceInventoryID: "inventory", Photos: true, OtherFiles: true}
	job, err := archivejob.New(request, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	stored, err := s.CreateArchiveJob(ctx, job)
	if err != nil {
		t.Fatal(err)
	}
	request.ID = "duplicate"
	duplicate, err := archivejob.New(request, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.CreateArchiveJob(ctx, duplicate)
	if err != nil || again.ID != stored.ID {
		t.Fatalf("duplicate request created another job: %v", err)
	}
	for _, scope := range []ports.ArchiveJobScope{{TenantID: "other", SourceInventoryID: "inventory"}, {TenantID: "tenant", SourceInventoryID: "other"}, {TenantID: "tenant"}} {
		if _, found, err := s.ArchiveJobByID(ctx, scope, job.ID); err != nil || found {
			t.Fatalf("scope leak: %v", err)
		}
	}
	claim, err := stored.Claim("first", now, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := s.UpdateArchiveJob(ctx, claim, stored.Revision); err != nil || !ok {
		t.Fatalf("claim: %v", err)
	}
	// Simulate a process restart using a new repository object over the same DB.
	restarted := NewStore(s.db)
	queued, err := restarted.ListRunnableArchiveJobs(ctx, now.Add(2*time.Minute), 10)
	if err != nil || len(queued) != 1 {
		t.Fatalf("expired claim not recovered: %v", err)
	}
	next, err := queued[0].Claim("second", now.Add(2*time.Minute), now.Add(3*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := restarted.UpdateArchiveJob(ctx, next, claim.Revision); err != nil || !ok {
		t.Fatalf("reclaim: %v", err)
	}
	stale, err := claim.Complete("first", now.Add(time.Second), "old-result")
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := s.UpdateArchiveJob(ctx, stale, claim.Revision); err != nil || ok {
		t.Fatalf("stale completion accepted: %v", err)
	}
	cancelled, err := next.Cancel(now.Add(2 * time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := s.UpdateArchiveJob(ctx, cancelled, next.Revision); err != nil || !ok {
		t.Fatalf("cancel: %v", err)
	}
	pending, err := s.ListRunnableArchiveJobs(ctx, now.Add(4*time.Minute), 10)
	if err != nil || len(pending) != 0 {
		t.Fatalf("cancelled job runnable: %v", err)
	}
}

func TestArchiveJobRejectsMutationOutsideTransition(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	saveTenant(t, ctx, s, tenant.ID("tenant"), "Home")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	job, err := archivejob.New(archivejob.Request{RequestKey: "immutable", ID: "immutable", TenantID: "tenant", PrincipalID: "owner", Kind: archivejob.Export, SourceInventoryID: "inventory"}, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateArchiveJob(ctx, job); err != nil {
		t.Fatal(err)
	}
	claim, err := job.Claim("worker", now, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*archivejob.Record){
		"expiry":      func(r *archivejob.Record) { r.ExpiresAt = r.ExpiresAt.Add(time.Hour) },
		"creation":    func(r *archivejob.Record) { r.CreatedAt = r.CreatedAt.Add(-time.Hour) },
		"request":     func(r *archivejob.Record) { r.Photos = true },
		"approval":    func(r *archivejob.Record) { r.DestinationInventoryID = "unapproved" },
		"publication": func(r *archivejob.Record) { r.State = archivejob.Ready; r.ResultArtifactID = "unclaimed" },
	} {
		t.Run(name, func(t *testing.T) {
			changed := claim
			mutate(&changed)
			if ok, err := s.UpdateArchiveJob(ctx, changed, job.Revision); err == nil || ok {
				t.Fatalf("invalid mutation persisted: %v", err)
			}
		})
	}
	if ok, err := s.UpdateArchiveJob(ctx, claim, job.Revision); err != nil || !ok {
		t.Fatalf("valid claim rejected: %v", err)
	}
}

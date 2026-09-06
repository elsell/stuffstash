package gormstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestAttachmentAndThumbnailJobCommitTogether(t *testing.T) {
	ctx := context.Background()
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	if err := store.SaveAttachment(ctx, attachment, record, &job); err != nil {
		t.Fatal(err)
	}
	claims, err := store.ClaimThumbnailJobs(ctx, "worker-one", 1, job.CreatedAt, job.CreatedAt.Add(time.Minute))
	if err != nil || len(claims) != 1 {
		t.Fatalf("durable work missing: %v %#v", err, claims)
	}
	if !claims[0].Job.Matches(attachment) {
		t.Fatal("claimed work lost attachment scope")
	}
	other, err := store.ClaimThumbnailJobs(ctx, "worker-two", 1, job.CreatedAt, job.CreatedAt.Add(time.Minute))
	if err != nil || len(other) != 0 {
		t.Fatal("claimed work was concurrently available")
	}
}

func TestAttachmentFailureDoesNotLeaveThumbnailJob(t *testing.T) {
	ctx := context.Background()
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	// Force the audit insert to fail after attachment insertion inside the transaction.
	if err := store.SaveAuditRecord(ctx, record); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveAttachment(ctx, attachment, record, &job); err == nil {
		t.Fatal("duplicate audit should roll back attachment creation")
	}
	_, found, err := store.AttachmentByID(ctx, tenant.ID(attachment.TenantID), inventory.InventoryID(attachment.InventoryID), asset.ID(attachment.AssetID), attachment.ID)
	if err != nil || found {
		t.Fatal("failed creation left an attachment")
	}
	claims, err := store.ClaimThumbnailJobs(ctx, "worker", 1, job.CreatedAt, job.CreatedAt.Add(time.Minute))
	if err != nil || len(claims) != 0 {
		t.Fatal("failed creation left derivative work")
	}
}

func TestThumbnailClaimRecoveryAndStaleAcknowledgement(t *testing.T) {
	ctx := context.Background()
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	if err := store.SaveAttachment(ctx, attachment, record, &job); err != nil {
		t.Fatal(err)
	}
	first, err := store.ClaimThumbnailJobs(ctx, "worker", 1, job.CreatedAt, job.CreatedAt.Add(time.Minute))
	if err != nil || len(first) != 1 {
		t.Fatal("initial claim failed")
	}
	now := job.CreatedAt.Add(2 * time.Minute)
	second, err := store.ClaimThumbnailJobs(ctx, "worker", 1, now, now.Add(time.Minute))
	if err != nil || len(second) != 1 {
		t.Fatal("expired work was not recovered")
	}
	if first[0].Attempts != 1 || second[0].Attempts != 2 {
		t.Fatal("crashed attempt was not counted")
	}
	done := ports.ThumbnailJobResolution{Status: ports.ThumbnailJobCompleted, At: now}
	if err := store.ResolveThumbnailJob(ctx, first[0], done); !errors.Is(err, ports.ErrOutboxClaimLost) {
		t.Fatalf("stale acknowledgement accepted: %v", err)
	}
	if err := store.ResolveThumbnailJob(ctx, second[0], done); err != nil {
		t.Fatal(err)
	}
	claims, err := store.ClaimThumbnailJobs(ctx, "third-worker", 1, now.Add(2*time.Minute), now.Add(3*time.Minute))
	if err != nil || len(claims) != 0 {
		t.Fatal("completed work reappeared")
	}
}

func TestThumbnailRetryEligibilityAndExhaustion(t *testing.T) {
	ctx := context.Background()
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	if err := store.SaveAttachment(ctx, attachment, record, &job); err != nil {
		t.Fatal(err)
	}
	claims, _ := store.ClaimThumbnailJobs(ctx, "worker-one", 1, job.CreatedAt, job.CreatedAt.Add(time.Minute))
	if len(claims) != 1 {
		t.Fatal("initial claim failed")
	}
	retryAt := job.CreatedAt.Add(time.Minute)
	retry := ports.ThumbnailJobResolution{Status: ports.ThumbnailJobPending, At: job.CreatedAt, NextAttemptAt: retryAt, Failure: ports.ThumbnailFailureProcessing}
	if err := store.ResolveThumbnailJob(ctx, claims[0], retry); err != nil {
		t.Fatal(err)
	}
	early, err := store.ClaimThumbnailJobs(ctx, "early", 1, job.CreatedAt, retryAt)
	if err != nil || len(early) != 0 {
		t.Fatal("backoff ignored")
	}
	claims, err = store.ClaimThumbnailJobs(ctx, "retry", 1, retryAt, retryAt.Add(time.Minute))
	if err != nil || len(claims) != 1 || claims[0].Attempts != 2 {
		t.Fatal("retry state was not durable")
	}
	failure := ports.ThumbnailJobResolution{Status: ports.ThumbnailJobFailed, At: retryAt, Failure: ports.ThumbnailFailureProcessing}
	if err := store.ResolveThumbnailJob(ctx, claims[0], failure); err != nil {
		t.Fatal(err)
	}
	claims, err = store.ClaimThumbnailJobs(ctx, "later", 1, retryAt.Add(time.Hour), retryAt.Add(2*time.Hour))
	if err != nil || len(claims) != 0 {
		t.Fatal("exhausted work was retried automatically")
	}
}

func TestThumbnailQueueRejectsWrongAttachmentScope(t *testing.T) {
	ctx := context.Background()
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	job.TenantID = "other-tenant"
	if err := store.SaveAttachment(ctx, attachment, record, &job); err == nil {
		t.Fatal("cross-tenant derivative work accepted")
	}
}

func thumbnailQueueFixture(t *testing.T, ctx context.Context) (Store, media.Attachment, audit.Record, media.ThumbnailJob) {
	t.Helper()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("thumbnail-tenant")
	inventoryID := inventory.InventoryID("thumbnail-inventory")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Photos")
	item := assetItem("thumbnail-asset", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	attachment := testAttachment(t, "thumbnail-attachment", item, "photo.jpg", media.ContentTypeJPEG, now)
	record := auditRecord(t, "thumbnail-created", tenantID, inventoryID, audit.ActionAttachmentCreated)
	job, err := media.NewThumbnailJob(attachment, media.ThumbnailJobNewImage, now)
	if err != nil {
		t.Fatal(err)
	}
	return store, attachment, record, job
}

func plannedThumbnailJob(t *testing.T, attachment media.Attachment) *media.ThumbnailJob {
	t.Helper()
	job, err := media.PlanThumbnailJob(attachment)
	if err != nil {
		t.Fatal(err)
	}
	return job
}

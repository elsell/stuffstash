package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestArchiveArtifactCleanupRecoversOrphansAndPreservesPublishedMedia(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	saveTenant(t, ctx, s, "tenant", "Home")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	orphan, err := archivejob.New(archivejob.Request{ID: "orphan", TenantID: "tenant", PrincipalID: "owner", RequestKey: "orphan", Kind: archivejob.Restore, SourceArtifactID: "archives/tenant/orphan/source.zip", SourceSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	key := media.StorageKey(orphan.SourceArtifactID)
	if err = s.RegisterArchiveArtifact(ctx, orphan, key, ports.ArchiveArtifactSource, now); err != nil {
		t.Fatal(err)
	}
	due, err := s.ListDueArchiveArtifacts(ctx, now, 10)
	if err != nil || len(due) != 0 {
		t.Fatal("live upload expired")
	}
	due, err = s.ListDueArchiveArtifacts(ctx, now.Add(time.Hour), 10)
	if err != nil || len(due) != 1 {
		t.Fatal("orphan not discoverable")
	}
	if err = s.db.Create(&blobDeletionEventModel{ID: "delete-orphan", StorageKey: "other", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if err = s.RetireArchiveArtifact(ctx, due[0], now.Add(time.Hour), "delete-orphan"); err == nil {
		t.Fatal("outbox collision committed retirement")
	}
	retained, err := s.ListDueArchiveArtifacts(ctx, now.Add(time.Hour), 10)
	if err != nil || len(retained) != 1 {
		t.Fatal("failed retirement lost journal")
	}
	if err = s.RegisterArchiveArtifact(ctx, orphan, key, ports.ArchiveArtifactSource, now); err != nil {
		t.Fatal("failed retirement leaked key tombstone")
	}
	if err = s.db.Delete(&blobDeletionEventModel{ID: "delete-orphan"}).Error; err != nil {
		t.Fatal(err)
	}
	if err = s.RetireArchiveArtifact(ctx, due[0], now.Add(time.Hour), "delete-orphan"); err != nil {
		t.Fatal(err)
	}
	events, err := s.ClaimPendingBlobDeletionEvents(ctx, "cleanup", 10, now.Add(time.Hour), now.Add(2*time.Hour))
	if err != nil || len(events) != 1 || events[0].StorageKey != key {
		t.Fatal("orphan deletion not durable")
	}
	if err = s.RegisterArchiveArtifact(ctx, orphan, key, ports.ArchiveArtifactSource, now); err == nil {
		t.Fatal("retired key reused")
	}
	// Existing attachment preservation is checked even for archived records.
	job := createRestoreClaim(t, s, now)
	restored := media.StorageKey("tenant/restored/asset/photo")
	if err = s.RegisterArchiveArtifact(ctx, job, restored, ports.ArchiveArtifactRestoredMedia, now); err != nil {
		t.Fatal(err)
	}
	artifact := ports.ArchiveArtifact{Key: restored, JobID: job.ID, TenantID: "tenant", Kind: ports.ArchiveArtifactRestoredMedia, ExpiresAt: job.ExpiresAt}
	if err = s.RetireArchiveArtifact(ctx, artifact, job.ExpiresAt, "too-early"); err == nil {
		t.Fatal("unfenced job retired")
	}
	expired, err := job.Expire(job.ExpiresAt)
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := s.UpdateArchiveJob(ctx, expired, job.Revision); err != nil || !ok {
		t.Fatal("expire")
	}
	saveInventory(t, ctx, s, "restored", "tenant", "Restored")
	// Create the minimal referenced row through the same internal model used by
	// atomic restore publication; no source blob is needed for reference safety.
	if err = s.db.Create(&assetModel{ID: "asset", TenantID: "tenant", InventoryID: "restored", Title: "Camera", Kind: "item", LifecycleState: "active", CustomFields: "{}", CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if err = s.db.Create(&attachmentModel{ID: "photo", TenantID: "tenant", InventoryID: "restored", AssetID: "asset", StorageKey: restored.String(), FileName: "photo.jpg", ContentType: "image/jpeg", SHA256: orphan.SourceSHA256, SizeBytes: 1, CreatedAt: now, LifecycleState: "archived"}).Error; err != nil {
		t.Fatal(err)
	}
	if err = s.RetireArchiveArtifact(ctx, artifact, job.ExpiresAt, "do-not-delete"); err != nil {
		t.Fatal(err)
	}
	events, err = s.ClaimPendingBlobDeletionEvents(ctx, "later", 10, job.ExpiresAt, job.ExpiresAt.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.StorageKey == restored {
			t.Fatal("published media queued for deletion")
		}
	}
}

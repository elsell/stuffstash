package media

import (
	"testing"
	"time"
)

func TestThumbnailJobBindsImageIdentityAndRevision(t *testing.T) {
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	attachment := Attachment{ID: "photo", TenantID: "tenant", InventoryID: "inventory", AssetID: "asset", StorageKey: "original", SHA256: SHA256("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), ContentType: ContentTypeJPEG, LifecycleState: LifecycleStateActive}
	job, err := NewThumbnailJob(attachment, ThumbnailJobNewImage, now)
	if err != nil {
		t.Fatal(err)
	}
	if !job.Matches(attachment) || job.Revision != CurrentThumbnailRevision || !job.CreatedAt.Equal(now) {
		t.Fatal("job lost image identity or creation time")
	}
	for _, changed := range []Attachment{
		{ID: "different", TenantID: attachment.TenantID, InventoryID: attachment.InventoryID, AssetID: attachment.AssetID, SHA256: attachment.SHA256, StorageKey: attachment.StorageKey},
		{ID: attachment.ID, TenantID: "other", InventoryID: attachment.InventoryID, AssetID: attachment.AssetID, SHA256: attachment.SHA256, StorageKey: attachment.StorageKey},
		{ID: attachment.ID, TenantID: attachment.TenantID, InventoryID: "other", AssetID: attachment.AssetID, SHA256: attachment.SHA256, StorageKey: attachment.StorageKey},
		{ID: attachment.ID, TenantID: attachment.TenantID, InventoryID: attachment.InventoryID, AssetID: "other", SHA256: attachment.SHA256, StorageKey: attachment.StorageKey},
		{ID: attachment.ID, TenantID: attachment.TenantID, InventoryID: attachment.InventoryID, AssetID: attachment.AssetID, SHA256: "changed", StorageKey: attachment.StorageKey},
		{ID: attachment.ID, TenantID: attachment.TenantID, InventoryID: attachment.InventoryID, AssetID: attachment.AssetID, SHA256: attachment.SHA256, StorageKey: "changed"},
	} {
		if job.Matches(changed) {
			t.Fatal("job matched changed identity or content")
		}
	}
	attachment.ContentType = ContentTypePDF
	if _, err := NewThumbnailJob(attachment, ThumbnailJobNewImage, now); err == nil {
		t.Fatal("non-image queued")
	}
}

func TestThumbnailJobRejectsIncompleteScopeAndUnknownPriority(t *testing.T) {
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	if _, err := NewThumbnailJob(Attachment{ContentType: ContentTypeJPEG}, ThumbnailJobNewImage, now); err == nil {
		t.Fatal("unscoped job accepted")
	}
	attachment := Attachment{ID: "photo", TenantID: "tenant", InventoryID: "inventory", AssetID: "asset", StorageKey: "original", SHA256: SHA256("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), ContentType: ContentTypeJPEG, LifecycleState: LifecycleStateActive}
	if _, err := NewThumbnailJob(attachment, ThumbnailJobPriority("unknown"), now); err == nil {
		t.Fatal("unknown scheduling priority accepted")
	}
	if _, err := NewThumbnailJob(attachment, ThumbnailJobBackfill, time.Time{}); err == nil {
		t.Fatal("missing job time accepted")
	}
	if _, err := NewThumbnailJob(attachment, ThumbnailJobBackfill, now); err != nil {
		t.Fatal(err)
	}
}

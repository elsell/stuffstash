package media

import (
	"errors"
	"time"
)

type ThumbnailRevision int

const CurrentThumbnailRevision ThumbnailRevision = 1

type ThumbnailJobPriority string

const (
	ThumbnailJobNewImage ThumbnailJobPriority = "new_image"
	ThumbnailJobBackfill ThumbnailJobPriority = "backfill"
)

// ThumbnailJob identifies durable derivative work, never a user credential or URL.
type ThumbnailJob struct {
	AttachmentID ID
	TenantID     TenantID
	InventoryID  InventoryID
	AssetID      AssetID
	StorageKey   StorageKey
	SHA256       SHA256
	Revision     ThumbnailRevision
	Priority     ThumbnailJobPriority
	CreatedAt    time.Time
}

func NewThumbnailJob(attachment Attachment, priority ThumbnailJobPriority, now time.Time) (ThumbnailJob, error) {
	if !attachment.ContentType.IsImage() || attachment.ID == "" || attachment.TenantID == "" || attachment.InventoryID == "" || attachment.AssetID == "" || attachment.StorageKey == "" || attachment.SHA256 == "" || now.IsZero() {
		return ThumbnailJob{}, errors.New("thumbnail job requires an identified image and creation time")
	}
	if priority != ThumbnailJobNewImage && priority != ThumbnailJobBackfill {
		return ThumbnailJob{}, errors.New("invalid thumbnail job priority")
	}
	return ThumbnailJob{
		AttachmentID: attachment.ID, TenantID: attachment.TenantID,
		InventoryID: attachment.InventoryID, AssetID: attachment.AssetID,
		StorageKey: attachment.StorageKey, SHA256: attachment.SHA256,
		Revision: CurrentThumbnailRevision, Priority: priority, CreatedAt: now.UTC(),
	}, nil
}

// Matches binds work to the authoritative attachment identity and immutable content.
// Lifecycle eligibility must be checked separately when claiming or publishing.
func (j ThumbnailJob) Matches(attachment Attachment) bool {
	return j.AttachmentID == attachment.ID && j.TenantID == attachment.TenantID &&
		j.InventoryID == attachment.InventoryID && j.AssetID == attachment.AssetID &&
		j.StorageKey == attachment.StorageKey && j.SHA256 == attachment.SHA256
}

// PlanThumbnailJob returns the work that must commit with a newly created attachment.
func PlanThumbnailJob(attachment Attachment) (*ThumbnailJob, error) {
	if !attachment.ContentType.IsImage() {
		return nil, nil
	}
	job, err := NewThumbnailJob(attachment, ThumbnailJobNewImage, attachment.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func ValidatePlannedThumbnailJob(attachment Attachment, job *ThumbnailJob) error {
	if !attachment.ContentType.IsImage() && job == nil {
		return nil
	}
	if job == nil || !attachment.ContentType.IsImage() || !job.Matches(attachment) || job.Revision != CurrentThumbnailRevision || job.CreatedAt.IsZero() || (job.Priority != ThumbnailJobNewImage && job.Priority != ThumbnailJobBackfill) {
		return errors.New("thumbnail work does not match attachment")
	}
	return nil
}

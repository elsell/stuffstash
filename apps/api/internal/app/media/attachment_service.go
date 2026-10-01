package media

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type AttachmentDependencies struct {
	Access                      ports.ActiveInventoryAccess
	Audit                       ports.AuditRepository
	Assets                      ports.AssetRepository
	AttachmentUnitOfWork        ports.AttachmentUnitOfWork
	Attachments                 ports.AttachmentRepository
	BlobDeletionClaimLease      time.Duration
	BlobDeletionMaxAttempts     int
	BlobDeletionOutbox          ports.BlobDeletionOutbox
	Blobs                       ports.BlobStorage
	Clock                       ports.Clock
	DefaultPageLimit            int
	DirectUploadTTL             time.Duration
	DirectUploads               ports.DirectAttachmentUploader
	IDs                         ports.IDGenerator
	ImageProcessor              ports.ImageProcessor
	MaxAttachmentBytes          int
	MaxPageLimit                int
	Observer                    ports.Observer
	PrimaryThumbnailWarmLimit   int
	PrimaryThumbnailWarmTimeout time.Duration
	ThumbnailGenerationState    *ThumbnailGenerationState
	ThumbnailReader             ports.ThumbnailReader
	ThumbnailWarmState          *PrimaryThumbnailWarmState
}
type AttachmentService struct{ deps AttachmentDependencies }

func NewAttachmentService(deps AttachmentDependencies) AttachmentService {
	return AttachmentService{deps: deps}
}

type auditRecordInput = appsupport.AuditRecordInput

func (a AttachmentService) newAuditRecord(input auditRecordInput) (audit.Record, error) {
	return appsupport.NewAuditRecord(a.deps.IDs, a.deps.Clock, input)
}
func (a AttachmentService) saveReadAuditRecord(ctx context.Context, input auditRecordInput) error {
	return appsupport.SaveReadAuditRecord(ctx, a.deps.Audit, a.deps.IDs, a.deps.Clock, input)
}

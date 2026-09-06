package media

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	domain "github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Processor struct {
	attachments        ports.AttachmentReader
	blobs              ports.BlobStorage
	images             ports.ImageBatchProcessor
	guard              ports.ThumbnailPublicationGuard
	publicationTimeout time.Duration
}

var _ ports.ThumbnailJobProcessor = (*Processor)(nil)

func NewProcessor(attachments ports.AttachmentReader, blobs ports.BlobStorage, images ports.ImageBatchProcessor, guard ports.ThumbnailPublicationGuard, publicationTimeout time.Duration) (*Processor, error) {
	if attachments == nil || blobs == nil || images == nil || guard == nil || publicationTimeout <= 0 {
		return nil, errors.New("thumbnail processor dependencies and publication timeout are required")
	}
	return &Processor{attachments: attachments, blobs: blobs, images: images, guard: guard, publicationTimeout: publicationTimeout}, nil
}

// ProcessThumbnailJob is called with shared image admission already held by Worker.
func (p *Processor) ProcessThumbnailJob(ctx context.Context, claim ports.ClaimedThumbnailJob) error {
	job := claim.Job
	if job.AttachmentID == "" || job.TenantID == "" || job.InventoryID == "" || job.AssetID == "" || job.Revision != domain.CurrentThumbnailRevision || claim.ClaimID == "" {
		return ports.ErrOutboxClaimLost
	}
	attachment, found, err := p.attachments.AttachmentByID(ctx, tenant.ID(job.TenantID), inventory.InventoryID(job.InventoryID), asset.ID(job.AssetID), job.AttachmentID)
	if err != nil {
		return err
	}
	if !found || !job.Matches(attachment) || !attachment.ContentType.IsImage() {
		return ports.ErrBlobNotFound
	}
	variants := []domain.ThumbnailVariant{}
	for _, variant := range []domain.ThumbnailVariant{domain.ThumbnailVariantSmall, domain.ThumbnailVariantMedium, domain.ThumbnailVariantLarge} {
		key, err := ThumbnailCacheKey(attachment.StorageKey, variant)
		if err != nil {
			return err
		}
		if _, err := readCachedThumbnail(ctx, p.blobs, key); errors.Is(err, ports.ErrBlobNotFound) {
			variants = append(variants, variant)
		} else if err != nil {
			return err
		}
	}
	if len(variants) == 0 {
		return nil
	}
	return p.generate(ctx, attachment, &claim, variants, nil)
}

func (p *Processor) generate(ctx context.Context, attachment domain.Attachment, claim *ports.ClaimedThumbnailJob, variants []domain.ThumbnailVariant, ready func(ports.ImageDerivative)) error {
	content, err := p.blobs.GetBlob(ctx, attachment.StorageKey)
	if err != nil {
		return err
	}
	return p.images.CreateThumbnails(ctx, ports.ImageDerivativesRequest{Attachment: attachment, ContentType: attachment.ContentType, Content: content, Variants: variants}, func(variant domain.ThumbnailVariant, derivative ports.ImageDerivative) error {
		if !derivative.ContentType.IsImage() || len(derivative.Content) == 0 {
			return errors.New("invalid generated thumbnail")
		}
		key, err := ThumbnailCacheKey(attachment.StorageKey, variant)
		if err != nil {
			return err
		}
		publishing, cancel := context.WithTimeout(ctx, p.publicationTimeout)
		defer cancel()
		err = p.guard.Publish(publishing, attachment, claim, func(writeCtx context.Context) error {
			if err := p.blobs.PutBlob(writeCtx, key, derivative.ContentType, derivative.Content); err != nil {
				return err
			}
			return p.blobs.PutBlob(writeCtx, thumbnailMetadataKey(key), domain.ContentType("text/plain"), []byte(derivative.ContentType.String()))
		})
		if err != nil {
			return err
		}
		if ready != nil {
			ready(derivative)
		}
		return nil
	})
}

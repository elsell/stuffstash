package media

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/blobstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/adapters/worklimit"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	domain "github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestProcessorPublishesMissingVariantsAndReusesCompletedCache(t *testing.T) {
	processor, source, blobs, guard, claim := processorFixture(t)
	if err := processor.ProcessThumbnailJob(context.Background(), claim); err != nil {
		t.Fatal(err)
	}
	if guard.writes != 3 {
		t.Fatal("three derivatives were not guarded")
	}
	for _, variant := range []domain.ThumbnailVariant{domain.ThumbnailVariantSmall, domain.ThumbnailVariantMedium, domain.ThumbnailVariantLarge} {
		key, _ := ThumbnailCacheKey(source.attachment.StorageKey, variant)
		if data, err := blobs.GetBlob(context.Background(), key); err != nil || len(data) == 0 {
			t.Fatal("missing derivative")
		}
	}
	// Removing the original proves a fully cached job needs no source download.
	if err := blobs.DeleteBlob(context.Background(), source.attachment.StorageKey); err != nil {
		t.Fatal(err)
	}
	if err := processor.ProcessThumbnailJob(context.Background(), claim); err != nil {
		t.Fatal(err)
	}
	if guard.writes != 3 {
		t.Fatal("completed variants were regenerated")
	}
}

func TestProcessorRetriesPartialPublicationWithoutRepeatingReadyVariant(t *testing.T) {
	processor, _, _, guard, claim := processorFixture(t)
	guard.failAt = 2
	if err := processor.ProcessThumbnailJob(context.Background(), claim); err == nil {
		t.Fatal("publication failure hidden")
	}
	guard.failAt = 0
	if err := processor.ProcessThumbnailJob(context.Background(), claim); err != nil {
		t.Fatal(err)
	}
	if guard.writes != 4 {
		t.Fatalf("partial retry regenerated ready output: %d writes", guard.writes)
	}
}

func TestProcessorRejectsMissingOrChangedAttachment(t *testing.T) {
	for _, missing := range []bool{false, true} {
		processor, source, _, guard, claim := processorFixture(t)
		if missing {
			source.missing = true
		} else {
			source.attachment.SHA256 = "changed"
		}
		if err := processor.ProcessThumbnailJob(context.Background(), claim); err == nil {
			t.Fatal("stale attachment was processed")
		}
		if guard.writes != 0 {
			t.Fatal("stale attachment was published")
		}
	}
}

type processorSource struct {
	attachment domain.Attachment
	missing    bool
}

func (s *processorSource) AttachmentByID(_ context.Context, t tenant.ID, i inventory.InventoryID, a asset.ID, id domain.ID) (domain.Attachment, bool, error) {
	if s.missing || string(t) != s.attachment.TenantID.String() || string(i) != s.attachment.InventoryID.String() || string(a) != s.attachment.AssetID.String() || id != s.attachment.ID {
		return domain.Attachment{}, false, nil
	}
	return s.attachment, true, nil
}

type processorGuard struct {
	source         *processorSource
	before         func()
	writes, failAt int
}

func (g *processorGuard) Publish(ctx context.Context, a domain.Attachment, claim *ports.ClaimedThumbnailJob, publish func(context.Context) error) error {
	if g.before != nil {
		g.before()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if g.source.missing || (claim != nil && (!claim.Job.Matches(a) || !claim.Job.Matches(g.source.attachment))) {
		return ports.ErrOutboxClaimLost
	}
	g.writes++
	if g.failAt == g.writes {
		return errors.New("controlled storage failure")
	}
	return publish(ctx)
}
func processorFixture(t *testing.T) (*Processor, *processorSource, *memory.Store, *processorGuard, ports.ClaimedThumbnailJob) {
	t.Helper()
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	attachment := domain.Attachment{ID: "image", TenantID: "tenant", InventoryID: "inventory", AssetID: "asset", StorageKey: "original", SHA256: "source-hash", ContentType: domain.ContentTypePNG, CreatedAt: now}
	source := &processorSource{attachment: attachment}
	blobs := memory.NewStore()
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	if err := blobs.PutBlob(context.Background(), attachment.StorageKey, attachment.ContentType, data.Bytes()); err != nil {
		t.Fatal(err)
	}
	guard := &processorGuard{source: source}
	processor, err := NewProcessor(source, blobs, blobstore.StandardImageProcessor{}, guard, worklimit.NewThumbnailReadiness(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	job, err := domain.NewThumbnailJob(attachment, domain.ThumbnailJobNewImage, now)
	if err != nil {
		t.Fatal(err)
	}
	return processor, source, blobs, guard, ports.ClaimedThumbnailJob{Job: job, ClaimID: "claim", ClaimedUntil: now.Add(time.Minute), Attempts: 1}
}

func TestProcessorRepairsMetadataWriteFailure(t *testing.T) {
	processor, source, blobs, _, claim := processorFixture(t)
	key, _ := ThumbnailCacheKey(source.attachment.StorageKey, domain.ThumbnailVariantSmall)
	failing := &processorFailingBlobs{BlobStorage: blobs, failKey: thumbnailMetadataKey(key)}
	processor.blobs = failing
	if err := processor.ProcessThumbnailJob(context.Background(), claim); err == nil {
		t.Fatal("metadata failure hidden")
	}
	if _, err := blobs.GetBlob(context.Background(), key); err != nil {
		t.Fatal("test did not leave derivative bytes behind", err)
	}
	if _, err := readCachedThumbnail(context.Background(), blobs, key); !errors.Is(err, ports.ErrBlobNotFound) {
		t.Fatal("incomplete derivative treated as ready", err)
	}
	failing.failKey = ""
	if err := processor.ProcessThumbnailJob(context.Background(), claim); err != nil {
		t.Fatal(err)
	}
	if _, err := readCachedThumbnail(context.Background(), blobs, key); err != nil {
		t.Fatal("retry did not repair metadata", err)
	}
}

func TestProcessorRejectsCancellationAndLostClaimAtPublication(t *testing.T) {
	for _, cancelDuringPublication := range []bool{false, true} {
		processor, source, blobs, guard, claim := processorFixture(t)
		ctx, cancel := context.WithCancel(context.Background())
		guard.before = func() {
			if cancelDuringPublication {
				cancel()
			} else {
				source.missing = true
			}
		}
		err := processor.ProcessThumbnailJob(ctx, claim)
		cancel()
		expected := ports.ErrOutboxClaimLost
		if cancelDuringPublication {
			expected = context.Canceled
		}
		if !errors.Is(err, expected) {
			t.Fatalf("publication error = %v, want %v", err, expected)
		}
		key, _ := ThumbnailCacheKey(source.attachment.StorageKey, domain.ThumbnailVariantSmall)
		if _, err := blobs.GetBlob(context.Background(), key); !errors.Is(err, ports.ErrBlobNotFound) {
			t.Fatal("invalid publication wrote bytes")
		}
	}
}

type processorFailingBlobs struct {
	ports.BlobStorage
	failKey domain.StorageKey
}

func (b *processorFailingBlobs) PutBlob(ctx context.Context, key domain.StorageKey, kind domain.ContentType, data []byte) error {
	if key == b.failKey {
		return errors.New("controlled metadata storage failure")
	}
	return b.BlobStorage.PutBlob(ctx, key, kind, data)
}

func TestProcessorNotifiesOnlySuccessfullyPublishedVariants(t *testing.T) {
	processor, source, _, guard, claim := processorFixture(t)
	smallKey, _ := ThumbnailCacheKey(source.attachment.StorageKey, domain.ThumbnailVariantSmall)
	mediumKey, _ := ThumbnailCacheKey(source.attachment.StorageKey, domain.ThumbnailVariantMedium)
	small, stopSmall := processor.readiness.Watch(smallKey)
	defer stopSmall()
	medium, stopMedium := processor.readiness.Watch(mediumKey)
	defer stopMedium()
	guard.failAt = 2
	if err := processor.ProcessThumbnailJob(context.Background(), claim); err == nil {
		t.Fatal("publication failure hidden")
	}
	select {
	case <-small:
	default:
		t.Fatal("published small did not notify")
	}
	select {
	case <-medium:
		t.Fatal("failed publication notified")
	default:
	}
	guard.failAt = 0
	if err := processor.ProcessThumbnailJob(context.Background(), claim); err != nil {
		t.Fatal(err)
	}
	select {
	case <-medium:
	default:
		t.Fatal("repaired publication did not notify")
	}
}

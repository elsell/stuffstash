package observability

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type observedBlobs struct {
	delegate  ports.BlobStorage
	telemetry ports.Telemetry
}

func ObserveBlobs(delegate ports.BlobStorage, telemetry ports.Telemetry) ports.BlobStorage {
	if delegate == nil {
		return nil
	}
	if telemetry == nil {
		telemetry = ports.NoopTelemetry{}
	}
	return observedBlobs{delegate, telemetry}
}
func (b observedBlobs) GetBlob(ctx context.Context, key media.StorageKey) (content []byte, err error) {
	ctx, finish := b.telemetry.Start(ctx, ports.OperationBlobRead)
	defer func() { finish(err) }()
	return b.delegate.GetBlob(ctx, key)
}
func (b observedBlobs) PutBlob(ctx context.Context, key media.StorageKey, kind media.ContentType, content []byte) (err error) {
	ctx, finish := b.telemetry.Start(ctx, ports.OperationBlobWrite)
	defer func() { finish(err) }()
	return b.delegate.PutBlob(ctx, key, kind, content)
}
func (b observedBlobs) DeleteBlob(ctx context.Context, key media.StorageKey) (err error) {
	ctx, finish := b.telemetry.Start(ctx, ports.OperationBlobDelete)
	defer func() { finish(err) }()
	return b.delegate.DeleteBlob(ctx, key)
}

type observedImages struct {
	delegate  ports.ImageProcessor
	telemetry ports.Telemetry
}

func ObserveImages(delegate ports.ImageProcessor, telemetry ports.Telemetry) ports.ImageProcessor {
	if delegate == nil {
		return nil
	}
	if telemetry == nil {
		telemetry = ports.NoopTelemetry{}
	}
	return observedImages{delegate, telemetry}
}
func (p observedImages) CreateThumbnail(ctx context.Context, request ports.ImageDerivativeRequest) (result ports.ImageDerivative, err error) {
	ctx, finish := p.telemetry.Start(ctx, ports.OperationThumbnailGenerate)
	defer func() { finish(err) }()
	return p.delegate.CreateThumbnail(ctx, request)
}
func (p observedImages) PrepareImageForModelUse(ctx context.Context, request ports.ModelImageRequest) (result ports.ModelImage, err error) {
	ctx, finish := p.telemetry.Start(ctx, ports.OperationModelImagePrepare)
	defer func() { finish(err) }()
	return p.delegate.PrepareImageForModelUse(ctx, request)
}

type observedUploads struct {
	delegate  ports.DirectAttachmentUploader
	telemetry ports.Telemetry
}

func ObserveUploads(delegate ports.DirectAttachmentUploader, telemetry ports.Telemetry) ports.DirectAttachmentUploader {
	if delegate == nil {
		return nil
	}
	if telemetry == nil {
		telemetry = ports.NoopTelemetry{}
	}
	return observedUploads{delegate, telemetry}
}
func (u observedUploads) CreateDirectAttachmentUpload(ctx context.Context, request ports.DirectAttachmentUploadRequest) (result ports.DirectAttachmentUpload, err error) {
	ctx, finish := u.telemetry.Start(ctx, ports.OperationUploadInitiate)
	defer func() { finish(err) }()
	return u.delegate.CreateDirectAttachmentUpload(ctx, request)
}
func (u observedUploads) CompleteDirectAttachmentUpload(ctx context.Context, id string) (result ports.CompletedDirectAttachmentUpload, err error) {
	ctx, finish := u.telemetry.Start(ctx, ports.OperationUploadVerify)
	defer func() { finish(err) }()
	return u.delegate.CompleteDirectAttachmentUpload(ctx, id)
}

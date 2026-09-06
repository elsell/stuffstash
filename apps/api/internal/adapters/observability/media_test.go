package observability

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/blobstore"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestMediaDecoratorsPreserveBytesAndProcessing(t *testing.T) {
	telemetry := &recordingOperations{}
	blobs := ObserveBlobs(blobstore.NewFileSystemStoreWithMaxBytes(t.TempDir(), 1024*1024), telemetry)
	key, _ := media.NewStorageKey("private/asset/original")
	var source bytes.Buffer
	if err := png.Encode(&source, image.NewRGBA(image.Rect(0, 0, 32, 24))); err != nil {
		t.Fatal(err)
	}
	if err := blobs.PutBlob(context.Background(), key, media.ContentTypePNG, source.Bytes()); err != nil {
		t.Fatal(err)
	}
	content, err := blobs.GetBlob(context.Background(), key)
	if err != nil || !bytes.Equal(content, source.Bytes()) {
		t.Fatal("blob contents changed")
	}
	processor := ObserveImages(blobstore.StandardImageProcessor{}, telemetry)
	derivative, err := processor.CreateThumbnail(context.Background(), ports.ImageDerivativeRequest{Content: content, ContentType: media.ContentTypePNG, Variant: media.ThumbnailVariantSmall})
	if err != nil || len(derivative.Content) == 0 || derivative.ContentType != media.ContentTypeJPEG {
		t.Fatal("image processing changed")
	}
	if err := blobs.DeleteBlob(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	if _, err := blobs.GetBlob(context.Background(), key); err == nil {
		t.Fatal("expected missing blob")
	}
	expected := []ports.Operation{ports.OperationBlobWrite, ports.OperationBlobRead, ports.OperationThumbnailGenerate, ports.OperationBlobDelete, ports.OperationBlobRead}
	if len(telemetry.operations) != len(expected) {
		t.Fatal("missing operation scopes")
	}
	for i, op := range expected {
		if telemetry.operations[i] != op {
			t.Fatal("wrong operation")
		}
	}
	if len(telemetry.results) != 5 || telemetry.results[4] == nil {
		t.Fatal("missing operation error")
	}
}

type recordingOperations struct {
	operations []ports.Operation
	results    []error
}

func (r *recordingOperations) Start(ctx context.Context, op ports.Operation) (context.Context, func(error)) {
	r.operations = append(r.operations, op)
	return ctx, func(err error) { r.results = append(r.results, err) }
}

package blobstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestLocalDirectAttachmentUploaderCompletesFromStoredBlob(t *testing.T) {
	store := NewFileSystemStoreWithMaxBytes(t.TempDir(), 32)
	uploader := NewLocalDirectAttachmentUploader(store)
	request := directUploadRequest(t, time.Now().Add(time.Hour))
	content := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

	upload, err := uploader.CreateDirectAttachmentUpload(context.Background(), request)
	if err != nil {
		t.Fatalf("create direct upload: %v", err)
	}
	if upload.Method != "PUT" || upload.URL == "" || upload.ExpiresAt != request.ExpiresAt {
		t.Fatalf("unexpected upload target: %+v", upload)
	}
	if err := store.PutBlob(context.Background(), request.StorageKey, request.ContentType, content); err != nil {
		t.Fatalf("put uploaded blob: %v", err)
	}

	completed, err := uploader.CompleteDirectAttachmentUpload(context.Background(), request.UploadID)
	if err != nil {
		t.Fatalf("complete direct upload: %v", err)
	}
	hashBytes := sha256.Sum256(content)
	if completed.StorageKey != request.StorageKey || completed.SizeBytes != int64(len(content)) || completed.SHA256.String() != hex.EncodeToString(hashBytes[:]) || completed.ExpiresAt != request.ExpiresAt {
		t.Fatalf("unexpected completion: %+v", completed)
	}
}

func TestLocalDirectAttachmentUploaderRejectsUnknownUpload(t *testing.T) {
	uploader := NewLocalDirectAttachmentUploader(NewFileSystemStore(t.TempDir()))

	_, err := uploader.CompleteDirectAttachmentUpload(context.Background(), "missing")
	if !errors.Is(err, ports.ErrDirectUploadIncomplete) {
		t.Fatalf("expected incomplete direct upload for unknown upload, got %v", err)
	}
}

func TestS3DirectAttachmentUploaderCreatesBoundedPostPolicy(t *testing.T) {
	store, err := NewS3Store(S3Config{
		Endpoint:  "127.0.0.1:3900",
		AccessKey: "access",
		SecretKey: "secret",
		Bucket:    "stuffstash",
		Region:    "garage",
		MaxBytes:  32,
	})
	if err != nil {
		t.Fatalf("create s3 store: %v", err)
	}
	request := directUploadRequest(t, time.Now().Add(time.Hour))
	uploader := NewS3DirectAttachmentUploader(store)

	upload, err := uploader.CreateDirectAttachmentUpload(context.Background(), request)
	if err != nil {
		t.Fatalf("create s3 direct upload: %v", err)
	}
	if upload.Method != "POST" || upload.URL == "" {
		t.Fatalf("expected presigned POST target, got %+v", upload)
	}
	if upload.UploadID == request.UploadID {
		t.Fatalf("expected opaque signed upload token, got raw upload id")
	}
	if upload.FormFields["key"] != request.StorageKey.String() || upload.FormFields["Content-Type"] != request.ContentType.String() {
		t.Fatalf("expected bounded form fields, got %+v", upload.FormFields)
	}
	if upload.Headers == nil {
		t.Fatalf("expected headers map to be present")
	}
}

func TestS3DirectUploadTokenSurvivesReconstructionAndRejectsTampering(t *testing.T) {
	store, err := NewS3Store(S3Config{
		Endpoint:  "127.0.0.1:3900",
		AccessKey: "access",
		SecretKey: "secret",
		Bucket:    "stuffstash",
		Region:    "garage",
		MaxBytes:  32,
	})
	if err != nil {
		t.Fatalf("create s3 store: %v", err)
	}
	request := directUploadRequest(t, time.Now().Add(time.Hour))
	token, err := encodeDirectUploadToken(request, store.directUploadSigningKey)
	if err != nil {
		t.Fatalf("encode token: %v", err)
	}

	decoded, err := decodeDirectUploadToken(token, store.directUploadSigningKey)
	if err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if decoded.AttachmentID != request.AttachmentID || decoded.StorageKey != request.StorageKey || decoded.SizeBytes != request.SizeBytes {
		t.Fatalf("unexpected decoded request: %+v", decoded)
	}
	if _, err := decodeDirectUploadToken(token+"tampered", store.directUploadSigningKey); !errors.Is(err, ports.ErrDirectUploadInvalid) {
		t.Fatalf("expected tampered token to fail safely, got %v", err)
	}
}

func TestStandardImageProcessorPreparesBoundedImageData(t *testing.T) {
	processor := StandardImageProcessor{}
	content := testPNG(t, 512, 128)

	thumbnail, err := processor.CreateThumbnail(context.Background(), ports.ImageDerivativeRequest{
		ContentType: media.ContentTypePNG,
		Content:     content,
	})
	if err != nil {
		t.Fatalf("create thumbnail: %v", err)
	}
	if thumbnail.ContentType != media.ContentTypePNG || len(thumbnail.Content) == 0 || len(thumbnail.Content) >= len(content) {
		t.Fatalf("unexpected thumbnail: %+v", thumbnail)
	}
	content[0] = 0
	if thumbnail.Content[0] == 0 {
		t.Fatalf("expected thumbnail bytes to be copied")
	}

	modelImage, err := processor.PrepareImageForModelUse(context.Background(), ports.ModelImageRequest{
		ContentType: media.ContentTypePNG,
		Content:     thumbnail.Content,
	})
	if err != nil {
		t.Fatalf("prepare model image: %v", err)
	}
	if modelImage.SizeBytes != int64(len(thumbnail.Content)) || modelImage.SHA256.String() == "" {
		t.Fatalf("unexpected model image: %+v", modelImage)
	}
	if modelImage.Width > thumbnailMaxDimension || modelImage.Height > thumbnailMaxDimension {
		t.Fatalf("expected thumbnail-sized model image dimensions, got %dx%d", modelImage.Width, modelImage.Height)
	}
}

func TestStandardImageProcessorHandlesWebP(t *testing.T) {
	processor := StandardImageProcessor{}
	content := testWebP(t)

	thumbnail, err := processor.CreateThumbnail(context.Background(), ports.ImageDerivativeRequest{
		ContentType: media.ContentTypeWEBP,
		Content:     content,
	})
	if err != nil {
		t.Fatalf("create webp thumbnail: %v", err)
	}
	if thumbnail.ContentType != media.ContentTypePNG || len(thumbnail.Content) == 0 {
		t.Fatalf("unexpected webp thumbnail: %+v", thumbnail)
	}

	modelImage, err := processor.PrepareImageForModelUse(context.Background(), ports.ModelImageRequest{
		ContentType: media.ContentTypeWEBP,
		Content:     content,
	})
	if err != nil {
		t.Fatalf("prepare webp model image: %v", err)
	}
	if modelImage.Width <= 0 || modelImage.Height <= 0 || modelImage.SHA256.String() == "" {
		t.Fatalf("unexpected webp model image: %+v", modelImage)
	}
}

func directUploadRequest(t *testing.T, expiresAt time.Time) ports.DirectAttachmentUploadRequest {
	t.Helper()

	attachmentID, ok := media.NewID("attachment-one")
	if !ok {
		t.Fatal("invalid attachment id")
	}
	storageKey, ok := media.NewStorageKey("tenant-one/inventory-one/asset-one/attachment-one")
	if !ok {
		t.Fatal("invalid storage key")
	}
	fileName, ok := media.NewFileName("receipt.png")
	if !ok {
		t.Fatal("invalid file name")
	}
	return ports.DirectAttachmentUploadRequest{
		UploadID:     "upload-one",
		AttachmentID: attachmentID,
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		AssetID:      asset.ID("asset-one"),
		StorageKey:   storageKey,
		FileName:     fileName,
		ContentType:  media.ContentTypePNG,
		SizeBytes:    8,
		ExpiresAt:    expiresAt,
	}
}

func testWebP(t *testing.T) []byte {
	t.Helper()

	content, err := base64.StdEncoding.DecodeString("UklGRrIBAABXRUJQVlA4TKUBAAAvSsAYAA8w//M///MfeJAkbXvaSG7m8Q3GfYSBJekwQztm/IcZlgwnmWImn2BK7aFmBtnVir6q//8VOkFE/xm4baTIu8c48ArEo6+B3zFKYln3pqClSCKX0begFTAXFOLXHSyF8cCNcZEG4OywuA4KVVfJCiArU7GAgJI8+lJP/OKMT/fBAjevg1cYB7YVkFuWga2lyPi5I0HFy5YTpWIHg0RZpkniRVW9odHAKOwosWuOGdxIyn2OvaCDvhg/we6TwadPBPbqBV58MsLmMJ8yZnOWk8SRz4N+QoyPL+MnamzMvcE1rHNEr91F9GKZPVUcS9w7PhhH36suB9qPeYb/oLk6cuTiJ0wOK3m5h1cKjW6EVZCYMK7dxcKCBdgP9HkKr9gkAO2P8GKZGWVdIAatQa+1IDpt6qyorVwdy01xdW8Jkfk6xjEXmVQQ+HQdFr6OKhIN34dXWq0+0qr6EJSCeeVLH9+gvGTLyqM65PQ44ihzlTXxQKjKbAvshXgir7Lil9w4L2bvMycmjQcqXaMCO6BlY28i+FOLzbfI1vEqxAhotocAAA==")
	if err != nil {
		t.Fatalf("decode webp fixture: %v", err)
	}
	return content
}

func testPNG(t *testing.T, width int, height int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 120, A: 255})
		}
	}
	buffer := bytes.Buffer{}
	if err := png.Encode(&buffer, img); err != nil {
		t.Fatalf("encode test png: %v", err)
	}
	return buffer.Bytes()
}

package blobstore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestS3StoreAgainstGarage(t *testing.T) {
	endpoint := os.Getenv("STUFF_STASH_TEST_S3_ENDPOINT")
	accessKey := os.Getenv("STUFF_STASH_TEST_S3_ACCESS_KEY")
	secretKey := os.Getenv("STUFF_STASH_TEST_S3_SECRET_KEY")
	bucket := os.Getenv("STUFF_STASH_TEST_S3_BUCKET")
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		t.Skip("Garage S3 test environment is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	store, err := NewS3Store(S3Config{
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Bucket:    bucket,
		Region:    "garage",
		Secure:    false,
		MaxBytes:  1024,
	})
	if err != nil {
		t.Fatalf("create S3 store: %v", err)
	}
	key := storageKey(t, "tenant/inventory/asset/garage-test")
	content := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

	if err := store.PutBlob(ctx, key, media.ContentTypePNG, content); err != nil {
		t.Fatalf("put blob: %v", err)
	}
	read, err := store.GetBlob(ctx, key)
	if err != nil {
		t.Fatalf("get blob: %v", err)
	}
	if !bytes.Equal(read, content) {
		t.Fatalf("expected %q, got %q", string(content), string(read))
	}
	if err := store.DeleteBlob(ctx, key); err != nil {
		t.Fatalf("delete blob: %v", err)
	}
	if _, err := store.GetBlob(ctx, key); !errors.Is(err, ports.ErrBlobNotFound) {
		t.Fatalf("expected deleted blob to be missing, got %v", err)
	}
}

func TestS3DirectUploadAgainstGarage(t *testing.T) {
	endpoint := os.Getenv("STUFF_STASH_TEST_S3_ENDPOINT")
	accessKey := os.Getenv("STUFF_STASH_TEST_S3_ACCESS_KEY")
	secretKey := os.Getenv("STUFF_STASH_TEST_S3_SECRET_KEY")
	bucket := os.Getenv("STUFF_STASH_TEST_S3_BUCKET")
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		t.Skip("Garage S3 test environment is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	store, err := NewS3Store(S3Config{
		Endpoint:       endpoint,
		PublicEndpoint: endpoint,
		AccessKey:      accessKey,
		SecretKey:      secretKey,
		Bucket:         bucket,
		Region:         "garage",
		Secure:         false,
		MaxBytes:       1024,
	})
	if err != nil {
		t.Fatalf("create S3 store: %v", err)
	}
	uploader := NewS3DirectAttachmentUploader(store)
	content := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	key := storageKey(t, "tenant/inventory/asset/direct-garage-test")
	attachmentID, ok := media.NewID("attachment-direct-garage")
	if !ok {
		t.Fatal("invalid attachment ID")
	}
	fileName, ok := media.NewFileName("direct.png")
	if !ok {
		t.Fatal("invalid file name")
	}

	upload, err := uploader.CreateDirectAttachmentUpload(ctx, ports.DirectAttachmentUploadRequest{
		UploadID:     "upload-direct-garage",
		AttachmentID: attachmentID,
		TenantID:     tenant.ID("tenant-home"),
		InventoryID:  inventory.InventoryID("inventory-home"),
		AssetID:      asset.ID("asset-home"),
		StorageKey:   key,
		FileName:     fileName,
		ContentType:  media.ContentTypePNG,
		SizeBytes:    int64(len(content)),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	})
	if err != nil {
		t.Fatalf("create direct upload: %v", err)
	}
	if upload.Method != http.MethodPost || len(upload.FormFields) == 0 {
		t.Fatalf("expected presigned POST form upload, got method=%s fields=%v", upload.Method, upload.FormFields)
	}

	status, body, err := postDirectUploadForm(ctx, upload.URL, upload.FormFields, "direct.png", media.ContentTypePNG.String(), content)
	if err != nil {
		t.Fatalf("post direct upload form: %v", err)
	}
	if status < 200 || status >= 300 {
		t.Fatalf("expected successful Garage direct upload, got HTTP %d with body %s", status, body)
	}

	completed, err := uploader.CompleteDirectAttachmentUpload(ctx, upload.UploadID)
	if err != nil {
		t.Fatalf("complete direct upload: %v", err)
	}
	if completed.AttachmentID != attachmentID || completed.SizeBytes != int64(len(content)) {
		t.Fatalf("unexpected completed direct upload: %+v", completed)
	}
}

func postDirectUploadForm(ctx context.Context, uploadURL string, fields map[string]string, fileName string, contentType string, content []byte) (int, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return 0, "", err
		}
	}
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return 0, "", err
	}
	if _, err := part.Write(content); err != nil {
		return 0, "", err
	}
	if err := writer.Close(); err != nil {
		return 0, "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, body)
	if err != nil {
		return 0, "", err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 0, "", err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 4096))
	if err != nil {
		return response.StatusCode, "", err
	}
	return response.StatusCode, string(responseBody), nil
}

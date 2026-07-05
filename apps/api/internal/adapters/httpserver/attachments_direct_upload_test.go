package httpserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestAttachmentDirectUploadRejectsUndecodableImageAtHTTPBoundary(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	content := truncatedPNGAttachmentContent()
	directUploads := &httpFakeDirectAttachmentUploader{content: content}
	server := NewServer(":0", newSeededMediaTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"asset-one", "op-asset-one", "audit-asset-one", "upload-one", "attachment-one"},
	}, directUploads, nil))
	assetResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	createdAsset := decodeAsset(t, assetResponse)
	initiate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/direct-uploads", "Bearer dev:owner", map[string]any{
		"fileName":    "receipt.png",
		"contentType": "image/png",
		"sizeBytes":   len(content),
	})
	upload := decodeDirectUpload(t, initiate)

	complete := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/direct-uploads/"+upload.Data.UploadID+"/complete", "Bearer dev:owner", nil)
	if complete.Code != http.StatusBadRequest {
		t.Fatalf("expected undecodable direct upload status %d, got %d with body %s", http.StatusBadRequest, complete.Code, complete.Body.String())
	}
	assertSafeError(t, complete, "invalid_request", "Attachment content does not match its file type.")

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:owner", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected attachment list status %d, got %d with body %s", http.StatusOK, list.Code, list.Body.String())
	}
	if body := decodeAttachmentList(t, list); len(body.Data) != 0 {
		t.Fatalf("expected failed direct upload to avoid metadata persistence, got %+v", body.Data)
	}
}

type httpFakeDirectAttachmentUploader struct {
	request ports.DirectAttachmentUploadRequest
	blobs   ports.BlobStorage
	content []byte
	err     error
}

func (f *httpFakeDirectAttachmentUploader) CreateDirectAttachmentUpload(_ context.Context, request ports.DirectAttachmentUploadRequest) (ports.DirectAttachmentUpload, error) {
	f.request = request
	return ports.DirectAttachmentUpload{
		UploadID:     request.UploadID,
		AttachmentID: request.AttachmentID,
		Method:       "PUT",
		URL:          "https://uploads.example.test/" + request.UploadID,
		Headers:      map[string]string{"content-type": request.ContentType.String()},
		ExpiresAt:    request.ExpiresAt,
	}, nil
}

func (f *httpFakeDirectAttachmentUploader) CompleteDirectAttachmentUpload(_ context.Context, uploadID string) (ports.CompletedDirectAttachmentUpload, error) {
	if f.err != nil {
		return ports.CompletedDirectAttachmentUpload{}, f.err
	}
	if f.request.UploadID != uploadID {
		return ports.CompletedDirectAttachmentUpload{}, app.ErrInvalidInput
	}
	content := f.content
	if len(content) == 0 {
		content = pngAttachmentContent()
	}
	hashBytes := sha256.Sum256(content)
	hash, ok := media.NewSHA256(hex.EncodeToString(hashBytes[:]))
	if !ok {
		return ports.CompletedDirectAttachmentUpload{}, app.ErrInvalidInput
	}
	if f.blobs != nil {
		if err := f.blobs.PutBlob(context.Background(), f.request.StorageKey, f.request.ContentType, content); err != nil {
			return ports.CompletedDirectAttachmentUpload{}, err
		}
	}
	return ports.CompletedDirectAttachmentUpload{
		UploadID:     uploadID,
		AttachmentID: f.request.AttachmentID,
		TenantID:     f.request.TenantID,
		InventoryID:  f.request.InventoryID,
		AssetID:      f.request.AssetID,
		StorageKey:   f.request.StorageKey,
		FileName:     f.request.FileName,
		ContentType:  f.request.ContentType,
		SizeBytes:    int64(len(content)),
		SHA256:       hash,
		ExpiresAt:    f.request.ExpiresAt,
	}, nil
}

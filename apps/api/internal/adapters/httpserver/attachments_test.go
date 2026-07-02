package httpserver

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
	"net/http"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/blobstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestAttachmentUploadListAndDownloadFlow(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"asset-one", "op-asset-one", "audit-asset-one", "attachment-one", "audit-attachment-one"},
	}))

	assetResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if assetResponse.Code != http.StatusCreated {
		t.Fatalf("expected asset status %d, got %d with body %s", http.StatusCreated, assetResponse.Code, assetResponse.Body.String())
	}
	createdAsset := decodeAsset(t, assetResponse)

	content := pngAttachmentContent()
	createAttachment := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:owner", map[string]any{
		"fileName":      "receipt.png",
		"contentType":   "image/png",
		"contentBase64": base64.StdEncoding.EncodeToString(content),
	})
	if createAttachment.Code != http.StatusCreated {
		t.Fatalf("expected attachment status %d, got %d with body %s", http.StatusCreated, createAttachment.Code, createAttachment.Body.String())
	}
	attachment := decodeAttachment(t, createAttachment)
	hash := sha256.Sum256(content)
	if attachment.Data.ID != "attachment-one" || attachment.Data.AssetID != createdAsset.Data.ID || attachment.Data.FileName != "receipt.png" || attachment.Data.SizeBytes != int64(len(content)) || attachment.Data.SHA256 != hex.EncodeToString(hash[:]) {
		t.Fatalf("unexpected attachment response: %+v", attachment.Data)
	}

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:owner", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, list.Code, list.Body.String())
	}
	listBody := decodeAttachmentList(t, list)
	if len(listBody.Data) != 1 || listBody.Data[0].ID != attachment.Data.ID {
		t.Fatalf("expected attachment in list, got %+v", listBody.Data)
	}
	if listBody.Meta.Pagination == nil || listBody.Meta.Pagination.Limit != 50 || listBody.Meta.Pagination.HasMore {
		t.Fatalf("unexpected pagination metadata: %+v", listBody.Meta.Pagination)
	}

	download := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID+"/content", "Bearer dev:owner", nil)
	if download.Code != http.StatusOK {
		t.Fatalf("expected download status %d, got %d with body %s", http.StatusOK, download.Code, download.Body.String())
	}
	if download.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("expected image/png content type, got %q", download.Header().Get("Content-Type"))
	}
	if !bytes.Equal(download.Body.Bytes(), content) {
		t.Fatalf("expected downloaded content %q, got %q", string(content), download.Body.String())
	}

	archiveAsset := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/archive", "Bearer dev:owner", nil)
	if archiveAsset.Code != http.StatusOK {
		t.Fatalf("expected asset archive status %d, got %d with body %s", http.StatusOK, archiveAsset.Code, archiveAsset.Body.String())
	}
	uploadToArchivedAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:owner", map[string]any{
		"fileName":      "later.png",
		"contentType":   "image/png",
		"contentBase64": base64.StdEncoding.EncodeToString(content),
	})
	if uploadToArchivedAsset.Code != http.StatusNotFound {
		t.Fatalf("expected archived asset upload status %d, got %d with body %s", http.StatusNotFound, uploadToArchivedAsset.Code, uploadToArchivedAsset.Body.String())
	}
	assertSafeError(t, uploadToArchivedAsset, "resource_not_found", "Resource not found.")
	listArchivedAssetAttachments := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:owner", nil)
	if listArchivedAssetAttachments.Code != http.StatusNotFound {
		t.Fatalf("expected archived asset attachment list status %d, got %d with body %s", http.StatusNotFound, listArchivedAssetAttachments.Code, listArchivedAssetAttachments.Body.String())
	}
	assertSafeError(t, listArchivedAssetAttachments, "resource_not_found", "Resource not found.")
	detailArchivedAssetAttachment := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID, "Bearer dev:owner", nil)
	if detailArchivedAssetAttachment.Code != http.StatusNotFound {
		t.Fatalf("expected archived asset attachment detail status %d, got %d with body %s", http.StatusNotFound, detailArchivedAssetAttachment.Code, detailArchivedAssetAttachment.Body.String())
	}
	assertSafeError(t, detailArchivedAssetAttachment, "resource_not_found", "Resource not found.")
	downloadArchivedAssetAttachment := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID+"/content", "Bearer dev:owner", nil)
	if downloadArchivedAssetAttachment.Code != http.StatusNotFound {
		t.Fatalf("expected archived asset attachment download status %d, got %d with body %s", http.StatusNotFound, downloadArchivedAssetAttachment.Code, downloadArchivedAssetAttachment.Body.String())
	}
	assertSafeError(t, downloadArchivedAssetAttachment, "resource_not_found", "Resource not found.")
}

func TestAttachmentDirectUploadFlow(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	directUploads := &httpFakeDirectAttachmentUploader{}
	server := NewServer(":0", newSeededMediaTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"asset-one", "op-asset-one", "audit-asset-one", "upload-one", "attachment-one", "audit-attachment-one"},
	}, directUploads, nil))
	assetResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if assetResponse.Code != http.StatusCreated {
		t.Fatalf("expected asset status %d, got %d with body %s", http.StatusCreated, assetResponse.Code, assetResponse.Body.String())
	}
	createdAsset := decodeAsset(t, assetResponse)

	initiate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/direct-uploads", "Bearer dev:owner", map[string]any{
		"fileName":    "receipt.png",
		"contentType": "image/png",
		"sizeBytes":   len(pngAttachmentContent()),
	})
	if initiate.Code != http.StatusCreated {
		t.Fatalf("expected direct upload status %d, got %d with body %s", http.StatusCreated, initiate.Code, initiate.Body.String())
	}
	upload := decodeDirectUpload(t, initiate)
	if upload.Data.UploadID != "upload-one" || upload.Data.AttachmentID != "attachment-one" || upload.Data.Method != "PUT" || strings.Contains(upload.Data.URL, "tenant") {
		t.Fatalf("unexpected direct upload response: %+v", upload.Data)
	}

	complete := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/direct-uploads/"+upload.Data.UploadID+"/complete", "Bearer dev:owner", nil)
	if complete.Code != http.StatusCreated {
		t.Fatalf("expected direct upload completion status %d, got %d with body %s", http.StatusCreated, complete.Code, complete.Body.String())
	}
	attachment := decodeAttachment(t, complete)
	if attachment.Data.ID != "attachment-one" || attachment.Data.FileName != "receipt.png" || attachment.Data.SizeBytes != int64(len(pngAttachmentContent())) {
		t.Fatalf("unexpected completed attachment: %+v", attachment.Data)
	}
}

func TestAttachmentDirectUploadCompletionFailureReturnsSafeEnvelope(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	directUploads := &httpFakeDirectAttachmentUploader{err: ports.ErrDirectUploadMismatch}
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
		"sizeBytes":   len(pngAttachmentContent()),
	})
	upload := decodeDirectUpload(t, initiate)

	complete := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/direct-uploads/"+upload.Data.UploadID+"/complete", "Bearer dev:owner", nil)
	if complete.Code != http.StatusBadRequest {
		t.Fatalf("expected completion failure status %d, got %d with body %s", http.StatusBadRequest, complete.Code, complete.Body.String())
	}
	assertSafeError(t, complete, "invalid_request", "Invalid request.")
}

func TestAttachmentThumbnailEndpointUsesImageProcessor(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	processor := &httpFakeImageProcessor{thumbnailContent: []byte("thumbnail")}
	server := NewServer(":0", newSeededMediaTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"asset-one", "op-asset-one", "audit-asset-one", "attachment-one", "audit-attachment-one", "audit-thumbnail"},
	}, nil, processor))
	assetResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if assetResponse.Code != http.StatusCreated {
		t.Fatalf("expected asset status %d, got %d with body %s", http.StatusCreated, assetResponse.Code, assetResponse.Body.String())
	}
	createdAsset := decodeAsset(t, assetResponse)
	createAttachment := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:owner", map[string]any{
		"fileName":      "receipt.png",
		"contentType":   "image/png",
		"contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent()),
	})
	if createAttachment.Code != http.StatusCreated {
		t.Fatalf("expected attachment status %d, got %d with body %s", http.StatusCreated, createAttachment.Code, createAttachment.Body.String())
	}
	attachment := decodeAttachment(t, createAttachment)

	thumbnail := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID+"/thumbnail?variant=small", "Bearer dev:owner", nil)
	if thumbnail.Code != http.StatusOK {
		t.Fatalf("expected thumbnail status %d, got %d with body %s", http.StatusOK, thumbnail.Code, thumbnail.Body.String())
	}
	if thumbnail.Header().Get("Content-Type") != "image/png" || thumbnail.Body.String() != "thumbnail" || !processor.thumbnailCalled {
		t.Fatalf("unexpected thumbnail response contentType=%q body=%q called=%t", thumbnail.Header().Get("Content-Type"), thumbnail.Body.String(), processor.thumbnailCalled)
	}
}

func TestAttachmentRealImageUploadDownloadAndThumbnailFlow(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededMediaTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"asset-one", "op-asset-one", "audit-asset-one", "attachment-one", "audit-attachment-one", "audit-download", "audit-thumbnail", "viewer-grant-event", "audit-viewer-grant", "viewer-claim", "audit-viewer-download", "audit-viewer-thumbnail"},
	}, nil, blobstore.StandardImageProcessor{}))
	assetResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if assetResponse.Code != http.StatusCreated {
		t.Fatalf("expected asset status %d, got %d with body %s", http.StatusCreated, assetResponse.Code, assetResponse.Body.String())
	}
	createdAsset := decodeAsset(t, assetResponse)
	content := realJPEGAttachmentContent(t)

	createAttachment := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:owner", map[string]any{
		"fileName":      "workbench.jpg",
		"contentType":   "image/jpeg",
		"contentBase64": base64.StdEncoding.EncodeToString(content),
	})
	if createAttachment.Code != http.StatusCreated {
		t.Fatalf("expected attachment status %d, got %d with body %s", http.StatusCreated, createAttachment.Code, createAttachment.Body.String())
	}
	attachment := decodeAttachment(t, createAttachment)

	download := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID+"/content", "Bearer dev:owner", nil)
	if download.Code != http.StatusOK {
		t.Fatalf("expected download status %d, got %d with body %s", http.StatusOK, download.Code, download.Body.String())
	}
	if !bytes.Equal(download.Body.Bytes(), content) {
		t.Fatalf("expected downloaded content to match uploaded image")
	}

	thumbnail := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID+"/thumbnail?variant=small", "Bearer dev:owner", nil)
	if thumbnail.Code != http.StatusOK {
		t.Fatalf("expected thumbnail status %d, got %d with body %s", http.StatusOK, thumbnail.Code, thumbnail.Body.String())
	}
	if thumbnail.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("expected image/jpeg thumbnail content type, got %q", thumbnail.Header().Get("Content-Type"))
	}
	thumbnailImage, _, err := image.Decode(bytes.NewReader(thumbnail.Body.Bytes()))
	if err != nil {
		t.Fatalf("expected decodable thumbnail image: %v", err)
	}
	if thumbnailImage.Bounds().Dx() > 256 || thumbnailImage.Bounds().Dy() > 256 {
		t.Fatalf("expected bounded thumbnail dimensions, got %dx%d", thumbnailImage.Bounds().Dx(), thumbnailImage.Bounds().Dy())
	}
	if len(thumbnail.Body.Bytes()) >= len(content) {
		t.Fatalf("expected thumbnail to be smaller than uploaded image")
	}
	mediumThumbnail := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID+"/thumbnail?variant=medium", "Bearer dev:owner", nil)
	if mediumThumbnail.Code != http.StatusOK {
		t.Fatalf("expected medium thumbnail status %d, got %d with body %s", http.StatusOK, mediumThumbnail.Code, mediumThumbnail.Body.String())
	}
	mediumThumbnailImage, _, err := image.Decode(bytes.NewReader(mediumThumbnail.Body.Bytes()))
	if err != nil {
		t.Fatalf("expected decodable medium thumbnail image: %v", err)
	}
	if mediumThumbnailImage.Bounds().Dx() != 768 || mediumThumbnailImage.Bounds().Dy() != 512 {
		t.Fatalf("expected medium thumbnail to preserve aspect ratio at medium bounds, got %dx%d", mediumThumbnailImage.Bounds().Dx(), mediumThumbnailImage.Bounds().Dy())
	}

	grantViewer := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]any{
		"principalId":  "viewer",
		"relationship": "viewer",
	})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, grantViewer.Code, grantViewer.Body.String())
	}
	viewerDownload := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID+"/content", "Bearer dev:viewer", nil)
	if viewerDownload.Code != http.StatusOK {
		t.Fatalf("expected viewer image download status %d, got %d with body %s", http.StatusOK, viewerDownload.Code, viewerDownload.Body.String())
	}
	viewerThumbnail := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID+"/thumbnail?variant=small", "Bearer dev:viewer", nil)
	if viewerThumbnail.Code != http.StatusOK {
		t.Fatalf("expected viewer thumbnail status %d, got %d with body %s", http.StatusOK, viewerThumbnail.Code, viewerThumbnail.Body.String())
	}
	intruderThumbnail := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID+"/thumbnail?variant=small", "Bearer dev:intruder", nil)
	if intruderThumbnail.Code != http.StatusForbidden {
		t.Fatalf("expected intruder thumbnail status %d, got %d with body %s", http.StatusForbidden, intruderThumbnail.Code, intruderThumbnail.Body.String())
	}
	assertSafeError(t, intruderThumbnail, "forbidden", "Forbidden.")
	missingAuthThumbnail := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID+"/thumbnail?variant=small", "", nil)
	if missingAuthThumbnail.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing auth thumbnail status %d, got %d with body %s", http.StatusUnauthorized, missingAuthThumbnail.Code, missingAuthThumbnail.Body.String())
	}
	assertSafeError(t, missingAuthThumbnail, "authentication_required", "Authentication required.")
}

func TestAttachmentListIsPaginated(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"asset-one", "op-asset-one", "audit-asset-one", "attachment-one", "audit-attachment-one", "attachment-two", "audit-attachment-two"},
	}))
	assetResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if assetResponse.Code != http.StatusCreated {
		t.Fatalf("expected asset status %d, got %d with body %s", http.StatusCreated, assetResponse.Code, assetResponse.Body.String())
	}
	createdAsset := decodeAsset(t, assetResponse)

	for _, name := range []string{"first.png", "second.png"} {
		createAttachment := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:owner", map[string]any{
			"fileName":      name,
			"contentType":   "image/png",
			"contentBase64": base64.StdEncoding.EncodeToString(append(pngAttachmentContent(), []byte(name)...)),
		})
		if createAttachment.Code != http.StatusCreated {
			t.Fatalf("expected attachment status %d, got %d with body %s", http.StatusCreated, createAttachment.Code, createAttachment.Body.String())
		}
	}

	firstPage := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments?limit=1", "Bearer dev:owner", nil)
	if firstPage.Code != http.StatusOK {
		t.Fatalf("expected first page status %d, got %d with body %s", http.StatusOK, firstPage.Code, firstPage.Body.String())
	}
	firstBody := decodeAttachmentList(t, firstPage)
	if len(firstBody.Data) != 1 || firstBody.Data[0].ID != "attachment-one" || firstBody.Meta.Pagination == nil || !firstBody.Meta.Pagination.HasMore || firstBody.Meta.Pagination.NextCursor == nil {
		t.Fatalf("unexpected first page: %+v", firstBody)
	}

	secondPage := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments?limit=1&cursor="+*firstBody.Meta.Pagination.NextCursor, "Bearer dev:owner", nil)
	if secondPage.Code != http.StatusOK {
		t.Fatalf("expected second page status %d, got %d with body %s", http.StatusOK, secondPage.Code, secondPage.Body.String())
	}
	secondBody := decodeAttachmentList(t, secondPage)
	if len(secondBody.Data) != 1 || secondBody.Data[0].ID != "attachment-two" || secondBody.Meta.Pagination == nil || secondBody.Meta.Pagination.HasMore {
		t.Fatalf("unexpected second page: %+v", secondBody)
	}
}

func TestAttachmentEndpointsEnforceAuthenticationAndAuthorization(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"asset-one", "op-asset-one", "audit-asset-one", "viewer-grant-event", "audit-viewer-grant", "viewer-claim", "attachment-one", "audit-attachment-one"},
	}))
	assetResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if assetResponse.Code != http.StatusCreated {
		t.Fatalf("expected asset status %d, got %d with body %s", http.StatusCreated, assetResponse.Code, assetResponse.Body.String())
	}
	createdAsset := decodeAsset(t, assetResponse)

	grantViewer := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]any{
		"principalId":  "viewer",
		"relationship": "viewer",
	})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("expected grant status %d, got %d with body %s", http.StatusCreated, grantViewer.Code, grantViewer.Body.String())
	}

	uploadBody := map[string]any{
		"fileName":      "receipt.png",
		"contentType":   "image/png",
		"contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent()),
	}
	directUploadBody := map[string]any{
		"fileName":    "receipt.png",
		"contentType": "image/png",
		"sizeBytes":   len(pngAttachmentContent()),
	}
	viewerUpload := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:viewer", uploadBody)
	if viewerUpload.Code != http.StatusForbidden {
		t.Fatalf("expected viewer upload status %d, got %d with body %s", http.StatusForbidden, viewerUpload.Code, viewerUpload.Body.String())
	}
	assertSafeError(t, viewerUpload, "forbidden", "Forbidden.")
	viewerDirectUpload := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/direct-uploads", "Bearer dev:viewer", directUploadBody)
	if viewerDirectUpload.Code != http.StatusForbidden {
		t.Fatalf("expected viewer direct upload status %d, got %d with body %s", http.StatusForbidden, viewerDirectUpload.Code, viewerDirectUpload.Body.String())
	}
	assertSafeError(t, viewerDirectUpload, "forbidden", "Forbidden.")

	ownerUpload := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:owner", uploadBody)
	if ownerUpload.Code != http.StatusCreated {
		t.Fatalf("expected owner upload status %d, got %d with body %s", http.StatusCreated, ownerUpload.Code, ownerUpload.Body.String())
	}
	attachment := decodeAttachment(t, ownerUpload)

	authCases := []struct {
		name          string
		method        string
		path          string
		authorization string
		body          any
		status        int
		code          string
	}{
		{name: "missing auth upload", method: http.MethodPost, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments", body: uploadBody, status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "malformed auth upload", method: http.MethodPost, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments", authorization: "Bearer nope", body: uploadBody, status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "intruder upload", method: http.MethodPost, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments", authorization: "Bearer dev:intruder", body: uploadBody, status: http.StatusForbidden, code: "forbidden"},
		{name: "missing auth direct upload", method: http.MethodPost, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments/direct-uploads", body: directUploadBody, status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "malformed auth direct upload", method: http.MethodPost, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments/direct-uploads", authorization: "Bearer nope", body: directUploadBody, status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "intruder direct upload", method: http.MethodPost, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments/direct-uploads", authorization: "Bearer dev:intruder", body: directUploadBody, status: http.StatusForbidden, code: "forbidden"},
		{name: "missing auth list", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments", status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "malformed auth list", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments", authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "intruder list", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments", authorization: "Bearer dev:intruder", status: http.StatusForbidden, code: "forbidden"},
		{name: "missing auth download", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments/" + attachment.Data.ID + "/content", status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "malformed auth download", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments/" + attachment.Data.ID + "/content", authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "intruder download", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments/" + attachment.Data.ID + "/content", authorization: "Bearer dev:intruder", status: http.StatusForbidden, code: "forbidden"},
	}
	for _, tc := range authCases {
		t.Run(tc.name, func(t *testing.T) {
			response := performRequest(server, tc.method, tc.path, tc.authorization, tc.body)
			if response.Code != tc.status {
				t.Fatalf("expected status %d, got %d with body %s", tc.status, response.Code, response.Body.String())
			}
			assertErrorCode(t, response, tc.code)
		})
	}

	viewerList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:viewer", nil)
	if viewerList.Code != http.StatusOK {
		t.Fatalf("expected viewer list status %d, got %d with body %s", http.StatusOK, viewerList.Code, viewerList.Body.String())
	}

	viewerDownload := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments/"+attachment.Data.ID+"/content", "Bearer dev:viewer", nil)
	if viewerDownload.Code != http.StatusOK {
		t.Fatalf("expected viewer download status %d, got %d with body %s", http.StatusOK, viewerDownload.Code, viewerDownload.Body.String())
	}

}

func TestAttachmentEndpointsHideCrossTenantAndCrossInventoryResources(t *testing.T) {
	const tenantOneID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const tenantTwoID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const inventoryOneID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const inventoryTwoID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantOneID, name: "Home", owner: "owner-one"},
			{id: tenantTwoID, name: "Cabin", owner: "owner-two"},
		},
		inventories: []seedInventory{
			{id: inventoryOneID, tenantID: tenantOneID, name: "Tools", owner: "owner-one"},
			{id: inventoryTwoID, tenantID: tenantTwoID, name: "Supplies", owner: "owner-two"},
		},
		ids: []string{"asset-one", "op-asset-one", "audit-asset-one", "asset-two", "op-asset-two", "audit-asset-two", "attachment-two", "audit-attachment-two"},
	}))
	assetOneResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantOneID+"/inventories/"+inventoryOneID+"/assets", "Bearer dev:owner-one", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if assetOneResponse.Code != http.StatusCreated {
		t.Fatalf("expected asset one status %d, got %d with body %s", http.StatusCreated, assetOneResponse.Code, assetOneResponse.Body.String())
	}
	assetOne := decodeAsset(t, assetOneResponse)
	assetTwoResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantTwoID+"/inventories/"+inventoryTwoID+"/assets", "Bearer dev:owner-two", map[string]any{
		"kind":  "item",
		"title": "Cabin Drill",
	})
	if assetTwoResponse.Code != http.StatusCreated {
		t.Fatalf("expected asset two status %d, got %d with body %s", http.StatusCreated, assetTwoResponse.Code, assetTwoResponse.Body.String())
	}
	assetTwo := decodeAsset(t, assetTwoResponse)
	attachmentTwoResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantTwoID+"/inventories/"+inventoryTwoID+"/assets/"+assetTwo.Data.ID+"/attachments", "Bearer dev:owner-two", map[string]any{
		"fileName":      "cabin.png",
		"contentType":   "image/png",
		"contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent()),
	})
	if attachmentTwoResponse.Code != http.StatusCreated {
		t.Fatalf("expected attachment two status %d, got %d with body %s", http.StatusCreated, attachmentTwoResponse.Code, attachmentTwoResponse.Body.String())
	}
	attachmentTwo := decodeAttachment(t, attachmentTwoResponse)

	crossAssetDownload := performRequest(server, http.MethodGet, "/tenants/"+tenantOneID+"/inventories/"+inventoryOneID+"/assets/"+assetOne.Data.ID+"/attachments/"+attachmentTwo.Data.ID+"/content", "Bearer dev:owner-one", nil)
	if crossAssetDownload.Code != http.StatusNotFound {
		t.Fatalf("expected cross asset status %d, got %d with body %s", http.StatusNotFound, crossAssetDownload.Code, crossAssetDownload.Body.String())
	}
	assertSafeError(t, crossAssetDownload, "resource_not_found", "Resource not found.")

	crossTenantList := performRequest(server, http.MethodGet, "/tenants/"+tenantTwoID+"/inventories/"+inventoryTwoID+"/assets/"+assetTwo.Data.ID+"/attachments", "Bearer dev:owner-one", nil)
	if crossTenantList.Code != http.StatusForbidden {
		t.Fatalf("expected cross tenant status %d, got %d with body %s", http.StatusForbidden, crossTenantList.Code, crossTenantList.Body.String())
	}
	assertSafeError(t, crossTenantList, "forbidden", "Forbidden.")
}

func TestAttachmentUploadRejectsUnsafeInput(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"asset-one", "op-asset-one", "audit-asset-one"},
	}))
	assetResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if assetResponse.Code != http.StatusCreated {
		t.Fatalf("expected asset status %d, got %d with body %s", http.StatusCreated, assetResponse.Code, assetResponse.Body.String())
	}
	createdAsset := decodeAsset(t, assetResponse)
	path := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + createdAsset.Data.ID + "/attachments"

	cases := []struct {
		name        string
		body        map[string]any
		wantMessage string
	}{
		{
			name:        "invalid base64",
			body:        map[string]any{"fileName": "receipt.png", "contentType": "image/png", "contentBase64": "not base64"},
			wantMessage: "Attachment content could not be read.",
		},
		{
			name: "unsupported content type",
			body: map[string]any{"fileName": "receipt.txt", "contentType": "text/plain", "contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent())},
		},
		{
			name:        "empty content",
			body:        map[string]any{"fileName": "receipt.png", "contentType": "image/png", "contentBase64": ""},
			wantMessage: "Attachment content is empty.",
		},
		{
			name:        "content type mismatch",
			body:        map[string]any{"fileName": "receipt.png", "contentType": "image/png", "contentBase64": base64.StdEncoding.EncodeToString([]byte("not a png"))},
			wantMessage: "Attachment content does not match its file type.",
		},
		{
			name:        "unsafe file name",
			body:        map[string]any{"fileName": "../receipt.png", "contentType": "image/png", "contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent())},
			wantMessage: "Invalid attachment file name.",
		},
		{
			name:        "too large",
			body:        map[string]any{"fileName": "large.png", "contentType": "image/png", "contentBase64": base64.StdEncoding.EncodeToString(append(pngAttachmentContent(), []byte(strings.Repeat("x", 25*1024*1024))...))},
			wantMessage: "Attachment is too large.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := performRequest(server, http.MethodPost, path, "Bearer dev:owner", tc.body)
			if response.Code != http.StatusBadRequest && response.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected status %d or %d, got %d with body %s", http.StatusBadRequest, http.StatusUnprocessableEntity, response.Code, response.Body.String())
			}
			if tc.wantMessage == "" {
				assertErrorCode(t, response, "invalid_request")
				return
			}
			assertSafeError(t, response, "invalid_request", tc.wantMessage)
		})
	}
}

func TestAttachmentUploadReturnsSafeStorageErrors(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestAppWithBlob(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"asset-one", "op-asset-one", "audit-asset-one", "attachment-one", "audit-attachment-one"},
	}, failingBlobStorage{}))
	assetResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if assetResponse.Code != http.StatusCreated {
		t.Fatalf("expected asset status %d, got %d with body %s", http.StatusCreated, assetResponse.Code, assetResponse.Body.String())
	}
	createdAsset := decodeAsset(t, assetResponse)

	response := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:owner", map[string]any{
		"fileName":      "receipt.png",
		"contentType":   "image/png",
		"contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent()),
	})
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected storage failure status %d, got %d with body %s", http.StatusInternalServerError, response.Code, response.Body.String())
	}
	assertSafeError(t, response, "internal_error", "Internal server error.")
}

type failingBlobStorage struct{}

func (failingBlobStorage) PutBlob(context.Context, media.StorageKey, media.ContentType, []byte) error {
	return errors.New("filesystem path /secret leaked")
}

func (failingBlobStorage) GetBlob(context.Context, media.StorageKey) ([]byte, error) {
	return nil, errors.New("filesystem path /secret leaked")
}

func (failingBlobStorage) DeleteBlob(context.Context, media.StorageKey) error {
	return nil
}

func pngAttachmentContent() []byte {
	return []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
}

func realPNGAttachmentContent(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 512, 128))
	for y := range 128 {
		for x := range 512 {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 140, A: 255})
		}
	}
	buffer := bytes.Buffer{}
	if err := png.Encode(&buffer, img); err != nil {
		t.Fatalf("encode png fixture: %v", err)
	}
	return buffer.Bytes()
}

func newSeededMediaTestApp(t *testing.T, state seededState, directUploads ports.DirectAttachmentUploader, imageProcessor ports.ImageProcessor) app.App {
	t.Helper()

	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	seedMemoryStore(t, context.Background(), store, authorizer, state)
	if fakeDirectUploads, ok := directUploads.(*httpFakeDirectAttachmentUploader); ok {
		fakeDirectUploads.blobs = store
	}

	return app.New(app.Dependencies{
		Observer:                  &fakeObserver{},
		Auth:                      auth.NewLocalDevAuthenticator(),
		Authorizer:                authorizer,
		Tenants:                   store,
		TenantUnitOfWork:          store,
		Inventories:               store,
		InventoryUnitOfWork:       store,
		InventoryAccess:           store,
		InventoryAccessUnitOfWork: store,
		CustomAssetTypes:          store,
		CustomAssetTypeUnitOfWork: store,
		CustomFields:              store,
		CustomFieldUnitOfWork:     store,
		Assets:                    store,
		AssetUnitOfWork:           store,
		Undoables:                 store,
		Search:                    store,
		Attachments:               store,
		AttachmentUnitOfWork:      store,
		Blobs:                     store,
		DirectUploads:             directUploads,
		ImageProcessor:            imageProcessor,
		BlobDeletionOutbox:        store,
		Audit:                     store,
		Outbox:                    store,
		ProviderProfiles:          store,
		ProviderProfileUnitOfWork: store,
		VoiceProviderConfigs:      store,
		ProviderCredentialVault:   httpTestCredentialVault{repository: store, sealer: httpTestCredentialSealer{}},
		ProviderProfileTester:     httpTestProviderProfileTester{},
		RealtimeSessions:          store,
		IDs:                       &fakeIDGenerator{ids: state.ids},
	})
}

type httpFakeDirectAttachmentUploader struct {
	request ports.DirectAttachmentUploadRequest
	blobs   ports.BlobStorage
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
	content := pngAttachmentContent()
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

type httpFakeImageProcessor struct {
	thumbnailCalled  bool
	thumbnailContent []byte
}

func (f *httpFakeImageProcessor) CreateThumbnail(_ context.Context, request ports.ImageDerivativeRequest) (ports.ImageDerivative, error) {
	f.thumbnailCalled = true
	return ports.ImageDerivative{ContentType: request.ContentType, Content: append([]byte(nil), f.thumbnailContent...)}, nil
}

func (f *httpFakeImageProcessor) PrepareImageForModelUse(_ context.Context, request ports.ModelImageRequest) (ports.ModelImage, error) {
	hashBytes := sha256.Sum256(request.Content)
	hash, ok := media.NewSHA256(hex.EncodeToString(hashBytes[:]))
	if !ok {
		return ports.ModelImage{}, app.ErrInvalidInput
	}
	return ports.ModelImage{
		ContentType: request.ContentType,
		Content:     append([]byte(nil), request.Content...),
		SizeBytes:   int64(len(request.Content)),
		SHA256:      hash,
		Width:       1,
		Height:      1,
	}, nil
}

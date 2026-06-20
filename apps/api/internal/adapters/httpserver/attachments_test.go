package httpserver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
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
	viewerUpload := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID+"/attachments", "Bearer dev:viewer", uploadBody)
	if viewerUpload.Code != http.StatusForbidden {
		t.Fatalf("expected viewer upload status %d, got %d with body %s", http.StatusForbidden, viewerUpload.Code, viewerUpload.Body.String())
	}
	assertSafeError(t, viewerUpload, "forbidden", "Forbidden.")

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
		name string
		body map[string]any
	}{
		{
			name: "invalid base64",
			body: map[string]any{"fileName": "receipt.png", "contentType": "image/png", "contentBase64": "not base64"},
		},
		{
			name: "unsupported content type",
			body: map[string]any{"fileName": "receipt.txt", "contentType": "text/plain", "contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent())},
		},
		{
			name: "empty content",
			body: map[string]any{"fileName": "receipt.png", "contentType": "image/png", "contentBase64": ""},
		},
		{
			name: "content type mismatch",
			body: map[string]any{"fileName": "receipt.png", "contentType": "image/png", "contentBase64": base64.StdEncoding.EncodeToString([]byte("not a png"))},
		},
		{
			name: "unsafe file name",
			body: map[string]any{"fileName": "../receipt.png", "contentType": "image/png", "contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent())},
		},
		{
			name: "too large",
			body: map[string]any{"fileName": "large.png", "contentType": "image/png", "contentBase64": base64.StdEncoding.EncodeToString(append(pngAttachmentContent(), []byte(strings.Repeat("x", 5*1024*1024))...))},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := performRequest(server, http.MethodPost, path, "Bearer dev:owner", tc.body)
			if response.Code != http.StatusBadRequest && response.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected status %d or %d, got %d with body %s", http.StatusBadRequest, http.StatusUnprocessableEntity, response.Code, response.Body.String())
			}
			assertErrorCode(t, response, "invalid_request")
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

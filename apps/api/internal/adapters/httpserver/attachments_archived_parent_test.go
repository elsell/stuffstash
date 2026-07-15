package httpserver

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/blobstore"
)

func TestArchivedAssetAttachmentReadsPreserveAuthorizationAndScope(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededMediaTestApp(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "owner"}},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"}},
	}, &httpFakeDirectAttachmentUploader{}, blobstore.StandardImageProcessor{}))

	assetResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind": "item", "title": "Archived drill",
	})
	if assetResponse.Code != http.StatusCreated {
		t.Fatalf("expected asset status %d, got %d with body %s", http.StatusCreated, assetResponse.Code, assetResponse.Body.String())
	}
	archivedAsset := decodeAsset(t, assetResponse)
	otherAssetResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind": "item", "title": "Other drill",
	})
	if otherAssetResponse.Code != http.StatusCreated {
		t.Fatalf("expected other asset status %d, got %d with body %s", http.StatusCreated, otherAssetResponse.Code, otherAssetResponse.Body.String())
	}
	otherAsset := decodeAsset(t, otherAssetResponse)
	content := realJPEGAttachmentContent(t)
	attachmentResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+archivedAsset.Data.ID+"/attachments", "Bearer dev:owner", map[string]any{
		"fileName": "archived-drill.jpg", "contentType": "image/jpeg", "contentBase64": base64.StdEncoding.EncodeToString(content),
	})
	if attachmentResponse.Code != http.StatusCreated {
		t.Fatalf("expected attachment status %d, got %d with body %s", http.StatusCreated, attachmentResponse.Code, attachmentResponse.Body.String())
	}
	attachment := decodeAttachment(t, attachmentResponse)
	grantViewer := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]any{
		"principalId": "viewer", "relationship": "viewer",
	})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, grantViewer.Code, grantViewer.Body.String())
	}
	archive := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+archivedAsset.Data.ID+"/archive", "Bearer dev:owner", nil)
	if archive.Code != http.StatusOK {
		t.Fatalf("expected archive status %d, got %d with body %s", http.StatusOK, archive.Code, archive.Body.String())
	}

	contentPath := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + archivedAsset.Data.ID + "/attachments/" + attachment.Data.ID + "/content"
	thumbnailPath := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + archivedAsset.Data.ID + "/attachments/" + attachment.Data.ID + "/thumbnail?variant=small"
	for _, principal := range []string{"owner", "viewer"} {
		download := performRequest(server, http.MethodGet, contentPath, "Bearer dev:"+principal, nil)
		if download.Code != http.StatusOK || download.Header().Get("Content-Type") != "image/jpeg" || !bytes.Equal(download.Body.Bytes(), content) {
			t.Fatalf("expected %s archived attachment download status/type/bytes, got %d %q", principal, download.Code, download.Header().Get("Content-Type"))
		}
		thumbnail := performRequest(server, http.MethodGet, thumbnailPath, "Bearer dev:"+principal, nil)
		if thumbnail.Code != http.StatusOK || thumbnail.Header().Get("Content-Type") != "image/jpeg" {
			t.Fatalf("expected %s archived thumbnail status/type, got %d %q with body %s", principal, thumbnail.Code, thumbnail.Header().Get("Content-Type"), thumbnail.Body.String())
		}
	}

	for _, tc := range []struct {
		name, path, auth, code string
		status                 int
	}{
		{name: "missing authentication", path: thumbnailPath, status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "intruder", path: thumbnailPath, auth: "Bearer dev:intruder", status: http.StatusForbidden, code: "forbidden"},
		{name: "wrong asset scope", path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + otherAsset.Data.ID + "/attachments/" + attachment.Data.ID + "/thumbnail?variant=small", auth: "Bearer dev:owner", status: http.StatusNotFound, code: "resource_not_found"},
		{name: "wrong attachment scope", path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + archivedAsset.Data.ID + "/attachments/missing-attachment/thumbnail?variant=small", auth: "Bearer dev:owner", status: http.StatusNotFound, code: "resource_not_found"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := performRequest(server, http.MethodGet, tc.path, tc.auth, nil)
			if response.Code != tc.status {
				t.Fatalf("expected status %d, got %d with body %s", tc.status, response.Code, response.Body.String())
			}
			assertErrorCode(t, response, tc.code)
		})
	}

	for _, mutation := range []struct {
		name, method, path string
		body               any
	}{
		{name: "upload", method: http.MethodPost, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + archivedAsset.Data.ID + "/attachments", body: map[string]any{"fileName": "later.jpg", "contentType": "image/jpeg", "contentBase64": base64.StdEncoding.EncodeToString(content)}},
		{name: "direct upload", method: http.MethodPost, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + archivedAsset.Data.ID + "/attachments/direct-uploads", body: map[string]any{"fileName": "later.jpg", "contentType": "image/jpeg", "sizeBytes": len(content)}},
		{name: "archive attachment", method: http.MethodPatch, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + archivedAsset.Data.ID + "/attachments/" + attachment.Data.ID + "/archive"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			response := performRequest(server, mutation.method, mutation.path, "Bearer dev:owner", mutation.body)
			if response.Code != http.StatusNotFound {
				t.Fatalf("expected archived-parent mutation status %d, got %d with body %s", http.StatusNotFound, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "resource_not_found", "Resource not found.")
		})
	}
}

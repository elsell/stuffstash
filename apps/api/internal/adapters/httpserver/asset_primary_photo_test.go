package httpserver

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
)

func TestAssetListIncludesSafePrimaryPhotoSummary(t *testing.T) {
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

	assetList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?limit=50", "Bearer dev:owner", nil)
	if assetList.Code != http.StatusOK {
		t.Fatalf("expected asset list status %d, got %d with body %s", http.StatusOK, assetList.Code, assetList.Body.String())
	}
	assetListBody := decodeAssetList(t, assetList)
	if len(assetListBody.Data) != 1 || assetListBody.Data[0].PrimaryPhoto == nil {
		t.Fatalf("expected asset list primary photo, got %+v", assetListBody.Data)
	}
	primaryPhoto := assetListBody.Data[0].PrimaryPhoto
	if primaryPhoto.ID != attachment.Data.ID || primaryPhoto.FileName != "receipt.png" || primaryPhoto.ContentType != "image/png" || primaryPhoto.SizeBytes != int64(len(content)) {
		t.Fatalf("unexpected primary photo summary: %+v", primaryPhoto)
	}
	if strings.Contains(primaryPhoto.Thumbnails.Small, "storage") || !strings.Contains(primaryPhoto.Thumbnails.Small, "/attachments/"+attachment.Data.ID+"/thumbnail?variant=small") {
		t.Fatalf("expected safe thumbnail API path, got %q", primaryPhoto.Thumbnails.Small)
	}

	assetDetail := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+createdAsset.Data.ID, "Bearer dev:owner", nil)
	if assetDetail.Code != http.StatusOK {
		t.Fatalf("expected asset detail status %d, got %d with body %s", http.StatusOK, assetDetail.Code, assetDetail.Body.String())
	}
	assetDetailBody := decodeAsset(t, assetDetail)
	if assetDetailBody.Data.PrimaryPhoto == nil {
		t.Fatalf("expected asset detail primary photo, got %+v", assetDetailBody.Data)
	}
	if assetDetailBody.Data.PrimaryPhoto.ID != attachment.Data.ID {
		t.Fatalf("expected detail primary photo %q, got %+v", attachment.Data.ID, assetDetailBody.Data.PrimaryPhoto)
	}
	if strings.Contains(assetDetailBody.Data.PrimaryPhoto.Thumbnails.Small, "storage") || !strings.Contains(assetDetailBody.Data.PrimaryPhoto.Thumbnails.Small, "/attachments/"+attachment.Data.ID+"/thumbnail?variant=small") {
		t.Fatalf("expected safe detail thumbnail API path, got %q", assetDetailBody.Data.PrimaryPhoto.Thumbnails.Small)
	}
}

package httpserver

import (
	"net/http/httptest"
	"testing"
)

type assetResponse struct {
	ID                string                 `json:"id"`
	TenantID          string                 `json:"tenantId"`
	InventoryID       string                 `json:"inventoryId"`
	ParentAssetID     string                 `json:"parentAssetId,omitempty"`
	CustomAssetTypeID string                 `json:"customAssetTypeId,omitempty"`
	Kind              string                 `json:"kind"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	CustomFields      map[string]interface{} `json:"customFields"`
	LifecycleState    string                 `json:"lifecycleState"`
	CreatedAt         string                 `json:"createdAt"`
	UpdatedAt         string                 `json:"updatedAt"`
}

type assetBody struct {
	Data assetResponse `json:"data"`
	Meta responseMeta  `json:"meta"`
}

type assetListBody struct {
	Data []assetResponse `json:"data"`
	Meta responseMeta    `json:"meta"`
}

func decodeAsset(t *testing.T, response *httptest.ResponseRecorder) assetBody {
	t.Helper()

	var body assetBody
	decodeBody(t, response, &body)
	return body
}

func decodeAssetList(t *testing.T, response *httptest.ResponseRecorder) assetListBody {
	t.Helper()

	var body assetListBody
	decodeBody(t, response, &body)
	return body
}

func assetListContainsID(items []assetResponse, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

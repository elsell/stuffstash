package httpserver

import (
	"net/http/httptest"
	"testing"
)

type searchAssetResultResponse struct {
	Type      string `json:"type"`
	TenantID  string `json:"tenantId"`
	Inventory struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"inventory"`
	Asset struct {
		ID                string             `json:"id"`
		InventoryID       string             `json:"inventoryId"`
		ParentAssetID     string             `json:"parentAssetId,omitempty"`
		CustomAssetTypeID string             `json:"customAssetTypeId,omitempty"`
		Kind              string             `json:"kind"`
		Title             string             `json:"title"`
		Description       string             `json:"description"`
		CustomFields      map[string]any     `json:"customFields"`
		LifecycleState    string             `json:"lifecycleState"`
		PrimaryPhoto      *assetPrimaryPhoto `json:"primaryPhoto,omitempty"`
		CurrentCheckout   *currentCheckout   `json:"currentCheckout,omitempty"`
	} `json:"asset"`
	Matches []struct {
		Field string `json:"field"`
		Value string `json:"value"`
	} `json:"matches"`
}

type searchAssetListBody struct {
	Data []searchAssetResultResponse `json:"data"`
	Meta responseMeta                `json:"meta"`
}

func decodeAssetSearch(t *testing.T, response *httptest.ResponseRecorder) searchAssetListBody {
	t.Helper()

	var body searchAssetListBody
	decodeBody(t, response, &body)
	return body
}

func assetSearchContainsTitle(results []searchAssetResultResponse, title string) bool {
	for _, result := range results {
		if result.Asset.Title == title {
			return true
		}
	}
	return false
}

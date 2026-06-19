package httpserver

import (
	"net/http/httptest"
	"testing"
)

type customAssetTypeResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	InventoryID string `json:"inventoryId,omitempty"`
	Scope       string `json:"scope"`
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type customAssetTypeBody struct {
	Data customAssetTypeResponse `json:"data"`
	Meta responseMeta            `json:"meta"`
}

type customAssetTypeListBody struct {
	Data []customAssetTypeResponse `json:"data"`
	Meta responseMeta              `json:"meta"`
}

func decodeCustomAssetType(t *testing.T, response *httptest.ResponseRecorder) customAssetTypeBody {
	t.Helper()

	var body customAssetTypeBody
	decodeBody(t, response, &body)
	return body
}

func decodeCustomAssetTypeList(t *testing.T, response *httptest.ResponseRecorder) customAssetTypeListBody {
	t.Helper()

	var body customAssetTypeListBody
	decodeBody(t, response, &body)
	return body
}

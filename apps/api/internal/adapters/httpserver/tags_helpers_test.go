package httpserver

import (
	"net/http/httptest"
	"testing"
)

type tagResponse struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenantId"`
	InventoryID    string `json:"inventoryId"`
	Key            string `json:"key"`
	DisplayName    string `json:"displayName"`
	Color          string `json:"color,omitempty"`
	LifecycleState string `json:"lifecycleState"`
}

type tagBody struct {
	Data tagResponse `json:"data"`
}

func decodeScenarioTag(t *testing.T, response *httptest.ResponseRecorder) tagBody {
	t.Helper()

	var body tagBody
	decodeBody(t, response, &body)
	return body
}

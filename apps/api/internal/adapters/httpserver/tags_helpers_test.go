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

type tagListBody struct {
	Data []tagResponse `json:"data"`
	Meta struct {
		Pagination *struct {
			Limit      int     `json:"limit"`
			NextCursor *string `json:"nextCursor"`
			HasMore    bool    `json:"hasMore"`
		} `json:"pagination"`
	} `json:"meta"`
}

func decodeScenarioTag(t *testing.T, response *httptest.ResponseRecorder) tagBody {
	t.Helper()

	var body tagBody
	decodeBody(t, response, &body)
	return body
}

func decodeScenarioTagList(t *testing.T, response *httptest.ResponseRecorder) tagListBody {
	t.Helper()

	var body tagListBody
	decodeBody(t, response, &body)
	return body
}

package httpserver

import (
	"net/http/httptest"
	"testing"
)

type inventoryResponse struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenantId"`
	Name           string `json:"name"`
	LifecycleState string `json:"lifecycleState"`
	Access         accessResponse
}

type inventoryBody struct {
	Data inventoryResponse `json:"data"`
	Meta responseMeta      `json:"meta"`
}

type inventoryListResponse struct {
	Data []inventoryResponse `json:"data"`
	Meta responseMeta        `json:"meta"`
}

type expectedInventory struct {
	id       string
	tenantID string
	name     string
}

func decodeInventoryList(t *testing.T, response *httptest.ResponseRecorder) []inventoryResponse {
	t.Helper()

	var body inventoryListResponse
	decodeBody(t, response, &body)
	return body.Data
}

func decodeInventoryListBody(t *testing.T, response *httptest.ResponseRecorder) inventoryListResponse {
	t.Helper()

	var body inventoryListResponse
	decodeBody(t, response, &body)
	return body
}

func assertInventories(t *testing.T, inventories []inventoryResponse, expected ...expectedInventory) {
	t.Helper()

	if len(inventories) != len(expected) {
		t.Fatalf("expected inventories %v, got %+v", expected, inventories)
	}

	seen := map[string]struct {
		tenantID string
		name     string
	}{}
	for _, item := range inventories {
		seen[item.ID] = struct {
			tenantID string
			name     string
		}{tenantID: item.TenantID, name: item.Name}
	}
	for _, item := range expected {
		actual, ok := seen[item.id]
		if !ok {
			t.Fatalf("expected inventory ID %q in %+v", item.id, inventories)
		}
		if actual.tenantID != item.tenantID || actual.name != item.name {
			t.Fatalf("expected inventory %q tenant/name %q/%q, got %q/%q", item.id, item.tenantID, item.name, actual.tenantID, actual.name)
		}
	}
}

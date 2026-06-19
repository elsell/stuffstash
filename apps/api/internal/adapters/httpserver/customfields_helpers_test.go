package httpserver

import (
	"net/http/httptest"
	"testing"
)

type customFieldDefinitionResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	InventoryID string `json:"inventoryId,omitempty"`
	Scope       string `json:"scope"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
}

type customFieldDefinitionBody struct {
	Data customFieldDefinitionResponse `json:"data"`
	Meta responseMeta                  `json:"meta"`
}

type customFieldDefinitionListBody struct {
	Data []customFieldDefinitionResponse `json:"data"`
	Meta responseMeta                    `json:"meta"`
}

func decodeCustomFieldDefinition(t *testing.T, response *httptest.ResponseRecorder) customFieldDefinitionBody {
	t.Helper()

	var body customFieldDefinitionBody
	decodeBody(t, response, &body)
	return body
}

func decodeCustomFieldDefinitionList(t *testing.T, response *httptest.ResponseRecorder) customFieldDefinitionListBody {
	t.Helper()

	var body customFieldDefinitionListBody
	decodeBody(t, response, &body)
	return body
}

func assertCustomFieldDefinition(t *testing.T, definition customFieldDefinitionResponse, tenantID string, inventoryID string, scope string, key string, fieldType string) {
	t.Helper()

	if definition.TenantID != tenantID || definition.InventoryID != inventoryID || definition.Scope != scope || definition.Key != key || definition.Type != fieldType {
		t.Fatalf("expected custom field definition %s/%s/%s/%s/%s, got %+v", tenantID, inventoryID, scope, key, fieldType, definition)
	}
}

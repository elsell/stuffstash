package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type customFieldDefinitionResponse struct {
	ID                 string   `json:"id"`
	TenantID           string   `json:"tenantId"`
	InventoryID        string   `json:"inventoryId,omitempty"`
	Scope              string   `json:"scope"`
	Key                string   `json:"key"`
	DisplayName        string   `json:"displayName"`
	Type               string   `json:"type"`
	EnumOptions        []string `json:"enumOptions"`
	Applicability      string   `json:"applicability"`
	CustomAssetTypeIDs []string `json:"customAssetTypeIds"`
	LifecycleState     string   `json:"lifecycleState"`
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

func createCustomAssetTypeForTest(t *testing.T, server *http.Server, path string, auth string, key string) customAssetTypeBody {
	t.Helper()

	response := performRequest(server, http.MethodPost, path, auth, map[string]any{
		"key":         key,
		"displayName": key,
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("expected custom asset type status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
	}
	return decodeCustomAssetType(t, response)
}

func customFieldDefinitionFromList(t *testing.T, server *http.Server, tenantID string, inventoryID string, fieldID string, inventoryScoped bool) customFieldDefinitionResponse {
	t.Helper()

	path := "/tenants/" + tenantID + "/custom-field-definitions?limit=20"
	auth := "Bearer dev:tenant-owner"
	if inventoryScoped {
		path = "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-field-definitions?limit=20"
		auth = "Bearer dev:inventory-owner"
	}
	response := performRequest(server, http.MethodGet, path, auth, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("expected custom field list status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}
	body := decodeCustomFieldDefinitionList(t, response)
	for _, definition := range body.Data {
		if definition.ID == fieldID {
			return definition
		}
	}
	t.Fatalf("expected custom field definition %q in list, got %+v", fieldID, body.Data)
	return customFieldDefinitionResponse{}
}

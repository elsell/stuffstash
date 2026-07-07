package httpserver

import (
	"net/http"
	"testing"
)

func TestTagListDefaultsLimitWhenQueryOmitted(t *testing.T) {
	server := NewServer(":0", newSeededTestApp(t, seededState{}))

	tenantCreate := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:owner", map[string]any{"name": "Home"})
	requireStatus(t, tenantCreate, http.StatusCreated)
	tenantID := decodeTenant(t, tenantCreate).Data.ID

	inventoryCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:owner", map[string]any{"name": "Household"})
	requireStatus(t, inventoryCreate, http.StatusCreated)
	inventoryID := decodeScenarioInventory(t, inventoryCreate).Data.ID
	inventoryPath := "/tenants/" + tenantID + "/inventories/" + inventoryID

	tagCreate := performRequest(server, http.MethodPost, inventoryPath+"/tags", "Bearer dev:owner", map[string]any{"displayName": "Workshop"})
	requireStatus(t, tagCreate, http.StatusCreated)

	list := performRequest(server, http.MethodGet, inventoryPath+"/tags", "Bearer dev:owner", nil)
	requireStatus(t, list, http.StatusOK)
	body := decodeScenarioTagList(t, list)
	if len(body.Data) != 1 || body.Data[0].Key != "workshop" {
		t.Fatalf("expected listed tag without explicit limit, got %+v", body.Data)
	}
	if body.Meta.Pagination == nil || body.Meta.Pagination.Limit != 50 {
		t.Fatalf("expected default pagination metadata, got %+v", body.Meta.Pagination)
	}
}

func TestTagListUsesDefaultLimitWhenExplicitZero(t *testing.T) {
	server := NewServer(":0", newSeededTestApp(t, seededState{}))

	tenantCreate := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:owner", map[string]any{"name": "Home"})
	requireStatus(t, tenantCreate, http.StatusCreated)
	tenantID := decodeTenant(t, tenantCreate).Data.ID

	inventoryCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:owner", map[string]any{"name": "Household"})
	requireStatus(t, inventoryCreate, http.StatusCreated)
	inventoryID := decodeScenarioInventory(t, inventoryCreate).Data.ID

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/tags?limit=0", "Bearer dev:owner", nil)
	requireStatus(t, list, http.StatusOK)
	body := decodeScenarioTagList(t, list)
	if body.Meta.Pagination == nil || body.Meta.Pagination.Limit != 50 {
		t.Fatalf("expected explicit zero limit to use default pagination metadata, got %+v", body.Meta.Pagination)
	}
}

func TestTagListRejectsNegativeLimit(t *testing.T) {
	server := NewServer(":0", newSeededTestApp(t, seededState{}))

	tenantCreate := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:owner", map[string]any{"name": "Home"})
	requireStatus(t, tenantCreate, http.StatusCreated)
	tenantID := decodeTenant(t, tenantCreate).Data.ID

	inventoryCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:owner", map[string]any{"name": "Household"})
	requireStatus(t, inventoryCreate, http.StatusCreated)
	inventoryID := decodeScenarioInventory(t, inventoryCreate).Data.ID

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/tags?limit=-1", "Bearer dev:owner", nil)
	if list.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected negative limit status %d, got %d with body %s", http.StatusUnprocessableEntity, list.Code, list.Body.String())
	}
}

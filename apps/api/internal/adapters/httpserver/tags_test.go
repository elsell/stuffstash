package httpserver

import (
	"net/http"
	"testing"
)

func TestTagEndpointsAuthorizeAndScopeInventory(t *testing.T) {
	server := NewServer(":0", newSeededTestApp(t, seededState{}))

	tenantCreate := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:owner", map[string]any{"name": "Home"})
	requireStatus(t, tenantCreate, http.StatusCreated)
	tenantID := decodeTenant(t, tenantCreate).Data.ID
	tenantPath := "/tenants/" + tenantID

	inventoryCreate := performRequest(server, http.MethodPost, tenantPath+"/inventories", "Bearer dev:owner", map[string]any{"name": "Household"})
	requireStatus(t, inventoryCreate, http.StatusCreated)
	inventoryID := decodeScenarioInventory(t, inventoryCreate).Data.ID
	inventoryPath := tenantPath + "/inventories/" + inventoryID

	otherInventoryCreate := performRequest(server, http.MethodPost, tenantPath+"/inventories", "Bearer dev:owner", map[string]any{"name": "Workshop"})
	requireStatus(t, otherInventoryCreate, http.StatusCreated)
	otherInventoryID := decodeScenarioInventory(t, otherInventoryCreate).Data.ID
	otherInventoryPath := tenantPath + "/inventories/" + otherInventoryID

	otherTenantCreate := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:other-owner", map[string]any{"name": "Cabin"})
	requireStatus(t, otherTenantCreate, http.StatusCreated)
	otherTenantID := decodeTenant(t, otherTenantCreate).Data.ID
	otherTenantPath := "/tenants/" + otherTenantID
	otherTenantInventoryCreate := performRequest(server, http.MethodPost, otherTenantPath+"/inventories", "Bearer dev:other-owner", map[string]any{"name": "Cabin Gear"})
	requireStatus(t, otherTenantInventoryCreate, http.StatusCreated)
	otherTenantInventoryID := decodeScenarioInventory(t, otherTenantInventoryCreate).Data.ID
	otherTenantInventoryPath := otherTenantPath + "/inventories/" + otherTenantInventoryID
	otherTenantTagCreate := performRequest(server, http.MethodPost, otherTenantInventoryPath+"/tags", "Bearer dev:other-owner", map[string]any{"displayName": "Private"})
	requireStatus(t, otherTenantTagCreate, http.StatusCreated)
	otherTenantTagID := decodeScenarioTag(t, otherTenantTagCreate).Data.ID

	grantViewer := performRequest(server, http.MethodPost, inventoryPath+"/access-grants", "Bearer dev:owner", map[string]any{
		"principalId":  "viewer",
		"relationship": "viewer",
	})
	requireStatus(t, grantViewer, http.StatusCreated)

	tagCreate := performRequest(server, http.MethodPost, inventoryPath+"/tags", "Bearer dev:owner", map[string]any{"displayName": "Workshop", "color": "#2f80ed"})
	requireStatus(t, tagCreate, http.StatusCreated)
	tag := decodeScenarioTag(t, tagCreate).Data
	tagPath := inventoryPath + "/tags/" + tag.ID

	ownerOnlyTagCreate := performRequest(server, http.MethodPost, inventoryPath+"/tags", "Bearer dev:owner", map[string]any{"displayName": "Owner Only"})
	requireStatus(t, ownerOnlyTagCreate, http.StatusCreated)
	ownerOnlyTagID := decodeScenarioTag(t, ownerOnlyTagCreate).Data.ID
	ownerOnlyTagPath := inventoryPath + "/tags/" + ownerOnlyTagID

	for _, item := range []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "list", method: http.MethodGet, path: inventoryPath + "/tags"},
		{name: "create", method: http.MethodPost, path: inventoryPath + "/tags", body: map[string]any{"displayName": "Blocked"}},
		{name: "update", method: http.MethodPatch, path: tagPath, body: map[string]any{"displayName": "Blocked"}},
		{name: "delete", method: http.MethodDelete, path: tagPath},
	} {
		t.Run("unauthenticated_"+item.name, func(t *testing.T) {
			response := performRequest(server, item.method, item.path, "", item.body)
			requireStatus(t, response, http.StatusUnauthorized)
			assertSafeError(t, response, "authentication_required", "Authentication required.")
		})
	}

	ownerUpdate := performRequest(server, http.MethodPatch, ownerOnlyTagPath, "Bearer dev:owner", map[string]any{"displayName": "Owner Edited", "color": "#00aa88"})
	requireStatus(t, ownerUpdate, http.StatusOK)
	if updated := decodeScenarioTag(t, ownerUpdate).Data; updated.DisplayName != "Owner Edited" || updated.Color != "#00AA88" {
		t.Fatalf("expected owner to update tag, got %+v", updated)
	}

	ownerDelete := performRequest(server, http.MethodDelete, ownerOnlyTagPath, "Bearer dev:owner", nil)
	requireStatus(t, ownerDelete, http.StatusOK)
	if archived := decodeScenarioTag(t, ownerDelete).Data; archived.LifecycleState != "archived" {
		t.Fatalf("expected owner delete to archive tag, got %+v", archived)
	}

	viewerList := performRequest(server, http.MethodGet, inventoryPath+"/tags", "Bearer dev:viewer", nil)
	requireStatus(t, viewerList, http.StatusOK)
	viewerTags := decodeScenarioTagList(t, viewerList)
	if len(viewerTags.Data) != 1 || viewerTags.Data[0].ID != tag.ID || viewerTags.Data[0].Color != "#2F80ED" {
		t.Fatalf("expected viewer to list scoped tag with normalized color, got %+v", viewerTags.Data)
	}

	viewerCreate := performRequest(server, http.MethodPost, inventoryPath+"/tags", "Bearer dev:viewer", map[string]any{"displayName": "Blocked"})
	requireStatus(t, viewerCreate, http.StatusForbidden)
	assertSafeError(t, viewerCreate, "forbidden", "Forbidden.")

	viewerUpdate := performRequest(server, http.MethodPatch, tagPath, "Bearer dev:viewer", map[string]any{"displayName": "Blocked"})
	requireStatus(t, viewerUpdate, http.StatusForbidden)
	assertSafeError(t, viewerUpdate, "forbidden", "Forbidden.")

	viewerDelete := performRequest(server, http.MethodDelete, tagPath, "Bearer dev:viewer", nil)
	requireStatus(t, viewerDelete, http.StatusForbidden)
	assertSafeError(t, viewerDelete, "forbidden", "Forbidden.")

	crossInventoryList := performRequest(server, http.MethodGet, otherInventoryPath+"/tags", "Bearer dev:owner", nil)
	requireStatus(t, crossInventoryList, http.StatusOK)
	if tags := decodeScenarioTagList(t, crossInventoryList).Data; len(tags) != 0 {
		t.Fatalf("expected other inventory list not to leak tags, got %+v", tags)
	}

	crossInventoryUpdate := performRequest(server, http.MethodPatch, otherInventoryPath+"/tags/"+tag.ID, "Bearer dev:owner", map[string]any{"displayName": "Blocked"})
	requireStatus(t, crossInventoryUpdate, http.StatusNotFound)
	assertSafeError(t, crossInventoryUpdate, "resource_not_found", "Resource not found.")

	crossInventoryDelete := performRequest(server, http.MethodDelete, otherInventoryPath+"/tags/"+tag.ID, "Bearer dev:owner", nil)
	requireStatus(t, crossInventoryDelete, http.StatusNotFound)
	assertSafeError(t, crossInventoryDelete, "resource_not_found", "Resource not found.")

	crossTenantList := performRequest(server, http.MethodGet, otherTenantInventoryPath+"/tags", "Bearer dev:owner", nil)
	requireStatus(t, crossTenantList, http.StatusForbidden)
	assertSafeError(t, crossTenantList, "forbidden", "Forbidden.")

	crossTenantUpdate := performRequest(server, http.MethodPatch, otherTenantInventoryPath+"/tags/"+otherTenantTagID, "Bearer dev:owner", map[string]any{"displayName": "Blocked"})
	requireStatus(t, crossTenantUpdate, http.StatusForbidden)
	assertSafeError(t, crossTenantUpdate, "forbidden", "Forbidden.")

	crossTenantDelete := performRequest(server, http.MethodDelete, otherTenantInventoryPath+"/tags/"+otherTenantTagID, "Bearer dev:owner", nil)
	requireStatus(t, crossTenantDelete, http.StatusForbidden)
	assertSafeError(t, crossTenantDelete, "forbidden", "Forbidden.")

	crossTenantScopedTag := performRequest(server, http.MethodPatch, otherTenantInventoryPath+"/tags/"+tag.ID, "Bearer dev:other-owner", map[string]any{"displayName": "Blocked"})
	requireStatus(t, crossTenantScopedTag, http.StatusNotFound)
	assertSafeError(t, crossTenantScopedTag, "resource_not_found", "Resource not found.")

	crossTenantList = performRequest(server, http.MethodGet, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/tags", "Bearer dev:owner", nil)
	requireStatus(t, crossTenantList, http.StatusNotFound)
	assertSafeError(t, crossTenantList, "resource_not_found", "Resource not found.")
}

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

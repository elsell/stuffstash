package httpserver

import (
	"net/http"
	"testing"
)

func TestUndoableOperationEndpointsUndoAndRedoAssetCreation(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const siblingInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAZ"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
			{id: otherTenantID, name: "Cabin", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
			{id: siblingInventoryID, tenantID: tenantID, name: "Medicine", owner: "owner"},
			{id: otherInventoryID, tenantID: otherTenantID, name: "Cabin Tools", owner: "owner"},
		},
		ids: []string{
			"asset-one", "op-asset-one", "audit-asset-one",
			"audit-viewer-grant", "viewer-grant-event", "viewer-claim",
			"audit-undo", "audit-redo",
		},
	}))

	create := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("expected asset create status %d, got %d with body %s", http.StatusCreated, create.Code, create.Body.String())
	}
	created := decodeAsset(t, create)

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]string{
		"principalId":  "viewer",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}

	auditResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?limit=10", "Bearer dev:owner", nil)
	if auditResponse.Code != http.StatusOK {
		t.Fatalf("expected audit status %d, got %d with body %s", http.StatusOK, auditResponse.Code, auditResponse.Body.String())
	}
	operationID := operationIDForTarget(t, decodeAuditRecordList(t, auditResponse).Data, created.Data.ID)

	missingAuth := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/undoable-operations/"+operationID+"/undo", "", nil)
	if missingAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing auth undo status %d, got %d with body %s", http.StatusUnauthorized, missingAuth.Code, missingAuth.Body.String())
	}
	assertSafeError(t, missingAuth, "authentication_required", "Authentication required.")

	viewerUndo := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/undoable-operations/"+operationID+"/undo", "Bearer dev:viewer", nil)
	if viewerUndo.Code != http.StatusForbidden {
		t.Fatalf("expected viewer undo status %d, got %d with body %s", http.StatusForbidden, viewerUndo.Code, viewerUndo.Body.String())
	}
	assertSafeError(t, viewerUndo, "forbidden", "Forbidden.")

	missingOperation := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/undoable-operations/missing-operation/undo", "Bearer dev:owner", nil)
	if missingOperation.Code != http.StatusNotFound {
		t.Fatalf("expected missing operation undo status %d, got %d with body %s", http.StatusNotFound, missingOperation.Code, missingOperation.Body.String())
	}
	assertSafeError(t, missingOperation, "resource_not_found", "Resource not found.")

	wrongInventory := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+siblingInventoryID+"/undoable-operations/"+operationID+"/undo", "Bearer dev:owner", nil)
	if wrongInventory.Code != http.StatusNotFound {
		t.Fatalf("expected wrong inventory undo status %d, got %d with body %s", http.StatusNotFound, wrongInventory.Code, wrongInventory.Body.String())
	}
	assertSafeError(t, wrongInventory, "resource_not_found", "Resource not found.")

	wrongTenant := performRequest(server, http.MethodPost, "/tenants/"+otherTenantID+"/inventories/"+otherInventoryID+"/undoable-operations/"+operationID+"/undo", "Bearer dev:owner", nil)
	if wrongTenant.Code != http.StatusNotFound {
		t.Fatalf("expected wrong tenant undo status %d, got %d with body %s", http.StatusNotFound, wrongTenant.Code, wrongTenant.Body.String())
	}
	assertSafeError(t, wrongTenant, "resource_not_found", "Resource not found.")

	undo := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/undoable-operations/"+operationID+"/undo", "Bearer dev:owner", nil)
	if undo.Code != http.StatusOK {
		t.Fatalf("expected undo status %d, got %d with body %s", http.StatusOK, undo.Code, undo.Body.String())
	}
	undone := decodeAsset(t, undo)
	if undone.Data.ID != created.Data.ID || undone.Data.LifecycleState != "archived" {
		t.Fatalf("expected undo to archive created asset, got %+v", undone.Data)
	}

	viewerRedo := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/undoable-operations/"+operationID+"/redo", "Bearer dev:viewer", nil)
	if viewerRedo.Code != http.StatusForbidden {
		t.Fatalf("expected viewer redo status %d, got %d with body %s", http.StatusForbidden, viewerRedo.Code, viewerRedo.Body.String())
	}
	assertSafeError(t, viewerRedo, "forbidden", "Forbidden.")

	redo := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/undoable-operations/"+operationID+"/redo", "Bearer dev:owner", nil)
	if redo.Code != http.StatusOK {
		t.Fatalf("expected redo status %d, got %d with body %s", http.StatusOK, redo.Code, redo.Body.String())
	}
	redone := decodeAsset(t, redo)
	if redone.Data.ID != created.Data.ID || redone.Data.LifecycleState != "active" {
		t.Fatalf("expected redo to restore created asset, got %+v", redone.Data)
	}
}

func TestRedoEndpointRejectsStaleAssetState(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{
			"asset-one", "op-asset-one", "audit-asset-one",
			"audit-list", "audit-undo",
			"op-restore", "audit-restore",
			"audit-redo",
		},
	}))

	create := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("expected asset create status %d, got %d with body %s", http.StatusCreated, create.Code, create.Body.String())
	}
	created := decodeAsset(t, create)
	auditResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?limit=10", "Bearer dev:owner", nil)
	if auditResponse.Code != http.StatusOK {
		t.Fatalf("expected audit status %d, got %d with body %s", http.StatusOK, auditResponse.Code, auditResponse.Body.String())
	}
	operationID := operationIDForTarget(t, decodeAuditRecordList(t, auditResponse).Data, created.Data.ID)

	undo := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/undoable-operations/"+operationID+"/undo", "Bearer dev:owner", nil)
	if undo.Code != http.StatusOK {
		t.Fatalf("expected undo status %d, got %d with body %s", http.StatusOK, undo.Code, undo.Body.String())
	}
	restore := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+created.Data.ID+"/restore", "Bearer dev:owner", nil)
	if restore.Code != http.StatusOK {
		t.Fatalf("expected restore status %d, got %d with body %s", http.StatusOK, restore.Code, restore.Body.String())
	}
	redo := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/undoable-operations/"+operationID+"/redo", "Bearer dev:owner", nil)
	if redo.Code != http.StatusBadRequest {
		t.Fatalf("expected stale redo status %d, got %d with body %s", http.StatusBadRequest, redo.Code, redo.Body.String())
	}
	assertSafeError(t, redo, "invalid_request", "Invalid request.")
}

func operationIDForTarget(t *testing.T, records []auditRecordResponse, targetID string) string {
	t.Helper()

	for _, record := range records {
		if record.Action == "asset.created" && record.TargetID == targetID && record.Metadata["operation_id"] != "" {
			return record.Metadata["operation_id"]
		}
	}
	t.Fatalf("expected asset.created audit record for %s to include operation_id, got %+v", targetID, records)
	return ""
}

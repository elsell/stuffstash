package httpserver

import (
	"net/http"
	"testing"
)

func TestSupportedAssetMutationResponsesExposeOnlyTheirOperationID(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "owner"}},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Household", owner: "owner"}},
		ids: []string{
			"asset-one", "operation-create", "audit-create",
			"operation-update", "audit-update",
			"operation-archive", "audit-archive",
			"operation-restore", "audit-restore",
			"audit-detail", "audit-list",
		},
	}))

	assetPath := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets"
	created := decodeAsset(t, performRequest(server, http.MethodPost, assetPath, "Bearer dev:owner", map[string]string{
		"kind": "item", "title": "Drill",
	}))
	if created.Data.UndoableOperationID != "operation-create" {
		t.Fatalf("expected create operation ID, got %+v", created.Data)
	}
	itemPath := assetPath + "/" + created.Data.ID
	updated := decodeAsset(t, performRequest(server, http.MethodPatch, itemPath, "Bearer dev:owner", map[string]string{"title": "Cordless drill"}))
	if updated.Data.UndoableOperationID != "operation-update" {
		t.Fatalf("expected update operation ID, got %+v", updated.Data)
	}
	archived := decodeAsset(t, performRequest(server, http.MethodPatch, itemPath+"/archive", "Bearer dev:owner", nil))
	if archived.Data.UndoableOperationID != "operation-archive" {
		t.Fatalf("expected archive operation ID, got %+v", archived.Data)
	}
	restored := decodeAsset(t, performRequest(server, http.MethodPatch, itemPath+"/restore", "Bearer dev:owner", nil))
	if restored.Data.UndoableOperationID != "operation-restore" {
		t.Fatalf("expected restore operation ID, got %+v", restored.Data)
	}
	detail := decodeAsset(t, performRequest(server, http.MethodGet, itemPath, "Bearer dev:owner", nil))
	if detail.Data.UndoableOperationID != "" {
		t.Fatalf("expected detail to omit operation ID, got %+v", detail.Data)
	}
	listed := decodeAssetList(t, performRequest(server, http.MethodGet, assetPath+"?limit=10", "Bearer dev:owner", nil))
	if len(listed.Data) != 1 || listed.Data[0].UndoableOperationID != "" {
		t.Fatalf("expected list to omit operation IDs, got %+v", listed.Data)
	}
}

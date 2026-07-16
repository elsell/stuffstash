package httpserver

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
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

func TestUndoReturnEndpointRejectsReopeningHistoricalLocationCheckout(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "owner"}},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "House", owner: "owner"}},
		ids:         []string{"audit-rejected-undo"},
	}, store, authorizer)
	location, open, _ := seedHistoricalLocationCheckout(t, store, tenantID, inventoryID, "garage", "")
	returned := open
	returned.State = asset.CheckoutStateReturned
	returned.ReturnedAt = open.CheckedOutAt.Add(time.Hour)
	returned.ReturnedByPrincipal = "owner"
	returned.UpdatedAt = returned.ReturnedAt
	operation := checkoutHTTPTestOperation("return-operation", tenantID, inventoryID, location.ID, audit.ActionAssetReturned, ports.UndoableOperationAvailable, &open, &returned)
	if err := store.ReturnAsset(context.Background(), open, returned, checkoutHTTPTestAudit("audit-return", tenantID, inventoryID, location.ID, audit.ActionAssetReturned), &operation); err != nil {
		t.Fatalf("seed historical location return: %v", err)
	}
	server := NewServer(":0", application)

	undo := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/undoable-operations/"+operation.ID+"/undo", "Bearer dev:owner", nil)
	requireStatus(t, undo, http.StatusBadRequest)
	assertSafeError(t, undo, "invalid_request", "Invalid request.")
	history := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+location.ID.String()+"/checkouts?limit=10", "Bearer dev:owner", nil)
	requireStatus(t, history, http.StatusOK)
	entries := decodeAssetCheckoutList(t, history).Data
	if len(entries) != 1 || entries[0].State != "returned" {
		t.Fatalf("expected rejected undo to preserve returned checkout, got %+v", entries)
	}
}

func TestRedoCheckoutEndpointRejectsReopeningHistoricalLocationCheckout(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "owner"}},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "House", owner: "owner"}},
		ids:         []string{"audit-undo", "audit-rejected-redo"},
	}, store, authorizer)
	location, _, operation := seedHistoricalLocationCheckout(t, store, tenantID, inventoryID, "garage", "checkout-operation")
	server := NewServer(":0", application)
	operationPath := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/undoable-operations/" + operation.ID

	undo := performRequest(server, http.MethodPost, operationPath+"/undo", "Bearer dev:owner", nil)
	requireStatus(t, undo, http.StatusOK)
	redo := performRequest(server, http.MethodPost, operationPath+"/redo", "Bearer dev:owner", nil)
	requireStatus(t, redo, http.StatusBadRequest)
	assertSafeError(t, redo, "invalid_request", "Invalid request.")
	history := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+location.ID.String()+"/checkouts?limit=10", "Bearer dev:owner", nil)
	requireStatus(t, history, http.StatusOK)
	entries := decodeAssetCheckoutList(t, history).Data
	if len(entries) != 1 || entries[0].State != "undone" {
		t.Fatalf("expected rejected redo to preserve undone checkout, got %+v", entries)
	}
}

func seedHistoricalLocationCheckout(t *testing.T, store *memory.Store, tenantID string, inventoryID string, assetID string, checkoutOperationID string) (asset.Asset, asset.Checkout, ports.UndoableOperation) {
	t.Helper()
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	location := asset.Asset{
		ID:             asset.ID(assetID),
		TenantID:       asset.TenantID(tenantID),
		InventoryID:    asset.InventoryID(inventoryID),
		Kind:           asset.KindLocation,
		Title:          asset.Title("Garage"),
		LifecycleState: asset.LifecycleStateActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := store.CreateAsset(context.Background(), location, checkoutHTTPTestAudit("audit-location", tenantID, inventoryID, location.ID, audit.ActionAssetCreated), nil); err != nil {
		t.Fatalf("seed historical location: %v", err)
	}
	checkout := asset.Checkout{
		ID:                    asset.CheckoutID("checkout-" + assetID),
		TenantID:              asset.TenantID(tenantID),
		InventoryID:           asset.InventoryID(inventoryID),
		AssetID:               location.ID,
		State:                 asset.CheckoutStateOpen,
		CheckedOutAt:          now,
		CheckedOutByPrincipal: "owner",
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	var operation ports.UndoableOperation
	var operationPointer *ports.UndoableOperation
	if checkoutOperationID != "" {
		operation = checkoutHTTPTestOperation(checkoutOperationID, tenantID, inventoryID, location.ID, audit.ActionAssetCheckedOut, ports.UndoableOperationAvailable, nil, &checkout)
		operationPointer = &operation
	}
	if err := store.CheckOutAsset(context.Background(), checkout, checkoutHTTPTestAudit("audit-checkout", tenantID, inventoryID, location.ID, audit.ActionAssetCheckedOut), operationPointer); err != nil {
		t.Fatalf("seed historical location checkout: %v", err)
	}
	return location, checkout, operation
}

func checkoutHTTPTestOperation(id string, tenantID string, inventoryID string, assetID asset.ID, action audit.Action, status ports.UndoableOperationStatus, before *asset.Checkout, after *asset.Checkout) ports.UndoableOperation {
	return ports.UndoableOperation{
		ID:             id,
		TenantID:       tenant.ID(tenantID),
		InventoryID:    inventory.InventoryID(inventoryID),
		PrincipalID:    identity.PrincipalID("owner"),
		Source:         audit.SourceAPI,
		TargetType:     audit.TargetAsset,
		TargetID:       assetID.String(),
		OriginalAction: action,
		Status:         status,
		CreatedAt:      time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC),
		BeforeCheckout: before,
		AfterCheckout:  after,
	}
}

func checkoutHTTPTestAudit(id string, tenantID string, inventoryID string, assetID asset.ID, action audit.Action) audit.Record {
	return audit.Record{
		ID:          audit.ID(id),
		TenantID:    audit.TenantID(tenantID),
		InventoryID: audit.InventoryID(inventoryID),
		PrincipalID: audit.PrincipalID("owner"),
		Action:      action,
		Source:      audit.SourceAPI,
		TargetType:  audit.TargetAsset,
		TargetID:    assetID.String(),
		OccurredAt:  time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC),
		Metadata:    map[string]string{},
	}
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

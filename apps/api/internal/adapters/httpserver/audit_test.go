package httpserver

import (
	"net/http"
	"testing"
)

func TestAuditRecordEndpointsEnforceScopeAndPagination(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const siblingInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAZ"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
			{id: siblingInventoryID, tenantID: tenantID, name: "Medicine", owner: "owner"},
			{id: otherInventoryID, tenantID: otherTenantID, name: "Cabin Tools", owner: "other-owner"},
		},
		ids: []string{
			"asset-one", "op-asset-one", "audit-asset-one",
			"asset-two", "op-asset-two", "audit-asset-two",
			"sibling-asset", "op-sibling-asset", "audit-sibling-asset",
			"other-tenant-asset", "op-other-tenant-asset", "audit-other-tenant-asset",
			"audit-viewer-grant", "viewer-grant-event", "viewer-grant-claim",
		},
	}))

	firstAsset := performRequestWithHeaders(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner:owner@example.test", map[string]string{"X-Request-ID": "request-audit-one"}, map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if firstAsset.Code != http.StatusCreated {
		t.Fatalf("expected first asset status %d, got %d with body %s", http.StatusCreated, firstAsset.Code, firstAsset.Body.String())
	}
	updateFirstAsset := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/asset-one", "Bearer dev:owner", map[string]any{
		"title": "Drill Kit",
	})
	if updateFirstAsset.Code != http.StatusOK {
		t.Fatalf("expected first asset update status %d, got %d with body %s", http.StatusOK, updateFirstAsset.Code, updateFirstAsset.Body.String())
	}
	secondAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Hammer",
	})
	if secondAsset.Code != http.StatusCreated {
		t.Fatalf("expected second asset status %d, got %d with body %s", http.StatusCreated, secondAsset.Code, secondAsset.Body.String())
	}
	siblingAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+siblingInventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Bandages",
	})
	if siblingAsset.Code != http.StatusCreated {
		t.Fatalf("expected sibling asset status %d, got %d with body %s", http.StatusCreated, siblingAsset.Code, siblingAsset.Body.String())
	}
	otherTenantAsset := performRequest(server, http.MethodPost, "/tenants/"+otherTenantID+"/inventories/"+otherInventoryID+"/assets", "Bearer dev:other-owner", map[string]any{
		"kind":  "item",
		"title": "Saw",
	})
	if otherTenantAsset.Code != http.StatusCreated {
		t.Fatalf("expected other tenant asset status %d, got %d with body %s", http.StatusCreated, otherTenantAsset.Code, otherTenantAsset.Body.String())
	}

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}

	firstPageResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?limit=1", "Bearer dev:viewer-user", nil)
	if firstPageResponse.Code != http.StatusOK {
		t.Fatalf("expected first audit page status %d, got %d with body %s", http.StatusOK, firstPageResponse.Code, firstPageResponse.Body.String())
	}
	firstPage := decodeAuditRecordList(t, firstPageResponse)
	if len(firstPage.Data) != 1 || firstPage.Meta.Pagination == nil || !firstPage.Meta.Pagination.HasMore || firstPage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected paginated first audit page, got %+v", firstPage)
	}
	if firstPage.Data[0].Action != "asset.created" || firstPage.Data[0].Source != "api" || firstPage.Data[0].TargetType != "asset" {
		t.Fatalf("unexpected first audit record: %+v", firstPage.Data[0])
	}
	if firstPage.Data[0].RequestID != "request-audit-one" {
		t.Fatalf("expected request ID on audit record, got %+v", firstPage.Data[0])
	}
	if firstPage.Data[0].Principal == nil || firstPage.Data[0].Principal.Email != "owner@example.test" {
		t.Fatalf("expected resolved principal on audit record, got %+v", firstPage.Data[0])
	}

	assetAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/asset-one/audit-records?limit=1", "Bearer dev:viewer-user", nil)
	if assetAudit.Code != http.StatusOK {
		t.Fatalf("expected asset audit status %d, got %d with body %s", http.StatusOK, assetAudit.Code, assetAudit.Body.String())
	}
	assetAuditBody := decodeAuditRecordList(t, assetAudit)
	if len(assetAuditBody.Data) != 1 || assetAuditBody.Data[0].TargetID != "asset-one" {
		t.Fatalf("expected asset-scoped audit records, got %+v", assetAuditBody.Data)
	}
	if assetAuditBody.Meta.Pagination == nil || !assetAuditBody.Meta.Pagination.HasMore {
		t.Fatalf("expected asset audit page to report more history, got %+v", assetAuditBody.Meta.Pagination)
	}
	if assetAuditBody.Data[0].Principal == nil || assetAuditBody.Data[0].Principal.Email != "owner@example.test" {
		t.Fatalf("expected resolved principal on asset audit record, got %+v", assetAuditBody.Data[0])
	}

	secondPageResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?limit=1&cursor="+*firstPage.Meta.Pagination.NextCursor, "Bearer dev:viewer-user", nil)
	if secondPageResponse.Code != http.StatusOK {
		t.Fatalf("expected second audit page status %d, got %d with body %s", http.StatusOK, secondPageResponse.Code, secondPageResponse.Body.String())
	}
	secondPage := decodeAuditRecordList(t, secondPageResponse)
	if len(secondPage.Data) != 1 {
		t.Fatalf("expected one record on second audit page, got %+v", secondPage)
	}

	firstInventoryAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?limit=50", "Bearer dev:owner", nil)
	if firstInventoryAudit.Code != http.StatusOK {
		t.Fatalf("expected first inventory audit status %d, got %d with body %s", http.StatusOK, firstInventoryAudit.Code, firstInventoryAudit.Body.String())
	}
	firstInventoryAuditBody := decodeAuditRecordList(t, firstInventoryAudit)
	if auditRecordsContainTarget(firstInventoryAuditBody.Data, "sibling-asset") || auditRecordsContainTarget(firstInventoryAuditBody.Data, "other-tenant-asset") {
		t.Fatalf("expected first inventory audit to exclude sibling and other tenant records, got %+v", firstInventoryAuditBody.Data)
	}

	tenantAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/audit-records?limit=50", "Bearer dev:owner", nil)
	if tenantAudit.Code != http.StatusOK {
		t.Fatalf("expected tenant audit status %d, got %d with body %s", http.StatusOK, tenantAudit.Code, tenantAudit.Body.String())
	}
	tenantAuditBody := decodeAuditRecordList(t, tenantAudit)
	if len(tenantAuditBody.Data) < 3 {
		t.Fatalf("expected tenant audit to include state changes, got %+v", tenantAuditBody.Data)
	}
	if auditRecordsContainTarget(tenantAuditBody.Data, "other-tenant-asset") {
		t.Fatalf("expected tenant audit to exclude other tenant records, got %+v", tenantAuditBody.Data)
	}

	viewerTenantAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/audit-records", "Bearer dev:viewer-user", nil)
	if viewerTenantAudit.Code != http.StatusForbidden {
		t.Fatalf("expected viewer tenant audit status %d, got %d with body %s", http.StatusForbidden, viewerTenantAudit.Code, viewerTenantAudit.Body.String())
	}
	assertSafeError(t, viewerTenantAudit, "forbidden", "Forbidden.")

	intruderInventoryAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records", "Bearer dev:intruder", nil)
	if intruderInventoryAudit.Code != http.StatusForbidden {
		t.Fatalf("expected intruder inventory audit status %d, got %d with body %s", http.StatusForbidden, intruderInventoryAudit.Code, intruderInventoryAudit.Body.String())
	}
	assertSafeError(t, intruderInventoryAudit, "forbidden", "Forbidden.")

	crossTenantAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/audit-records", "Bearer dev:other-owner", nil)
	if crossTenantAudit.Code != http.StatusForbidden {
		t.Fatalf("expected cross-tenant audit status %d, got %d with body %s", http.StatusForbidden, crossTenantAudit.Code, crossTenantAudit.Body.String())
	}
	assertSafeError(t, crossTenantAudit, "forbidden", "Forbidden.")

	wrongScopeCursor := paginationCursor(map[string]any{
		"v":          1,
		"collection": "audit_records",
		"scope":      tenantID + ":" + siblingInventoryID,
		"lastId":     firstPage.Data[0].ID,
	})
	wrongScopeAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?cursor="+wrongScopeCursor, "Bearer dev:owner", nil)
	if wrongScopeAudit.Code != http.StatusBadRequest {
		t.Fatalf("expected wrong-scope cursor status %d, got %d with body %s", http.StatusBadRequest, wrongScopeAudit.Code, wrongScopeAudit.Body.String())
	}
	assertSafeError(t, wrongScopeAudit, "invalid_request", "Invalid request.")
}

package httpserver

import (
	"net/http"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
)

func TestInventorySharingEnforcesViewerEditorAndShareBoundaries(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const hiddenInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
			{id: hiddenInventoryID, tenantID: tenantID, name: "Hidden", owner: "other-inventory-owner"},
		},
		ids: []string{
			"audit-viewer-grant", "viewer-grant-event", "viewer-grant-claim",
			"audit-duplicate-viewer-grant", "duplicate-viewer-grant-event", "duplicate-viewer-grant-claim",
			"audit-editor-grant", "editor-grant-event", "editor-grant-claim",
			"editor-created-asset", "audit-editor-created-asset",
			"audit-viewer-revoke", "viewer-revoke-event", "viewer-revoke-claim",
			"audit-missing-viewer-revoke", "missing-viewer-revoke-event", "missing-viewer-revoke-claim",
			"audit-editor-revoke", "editor-revoke-event", "editor-revoke-claim",
		},
	}))

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}
	assertInventoryAccessGrant(t, decodeInventoryAccessGrant(t, viewerGrant).Data, tenantID, inventoryID, "viewer-user", "viewer")

	duplicateViewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if duplicateViewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected duplicate viewer grant status %d, got %d with body %s", http.StatusCreated, duplicateViewerGrant.Code, duplicateViewerGrant.Body.String())
	}
	assertInventoryAccessGrant(t, decodeInventoryAccessGrant(t, duplicateViewerGrant).Data, tenantID, inventoryID, "viewer-user", "viewer")

	editorGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "editor-user",
		"relationship": "editor",
	})
	if editorGrant.Code != http.StatusCreated {
		t.Fatalf("expected editor grant status %d, got %d with body %s", http.StatusCreated, editorGrant.Code, editorGrant.Body.String())
	}
	assertInventoryAccessGrant(t, decodeInventoryAccessGrant(t, editorGrant).Data, tenantID, inventoryID, "editor-user", "editor")

	invalidRelationship := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "bad-grant-user",
		"relationship": "owner",
	})
	if invalidRelationship.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected invalid relationship status %d, got %d with body %s", http.StatusUnprocessableEntity, invalidRelationship.Code, invalidRelationship.Body.String())
	}
	assertErrorCode(t, invalidRelationship, "invalid_request")

	wrongTenantGrant := performRequest(server, http.MethodPost, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "wrong-tenant-user",
		"relationship": "viewer",
	})
	if wrongTenantGrant.Code != http.StatusNotFound {
		t.Fatalf("expected wrong tenant grant status %d, got %d with body %s", http.StatusNotFound, wrongTenantGrant.Code, wrongTenantGrant.Body.String())
	}
	assertSafeError(t, wrongTenantGrant, "resource_not_found", "Resource not found.")

	wrongInventoryGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/01ARZ3NDEKTSV4RRFFQ69G5FB0/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "wrong-inventory-user",
		"relationship": "viewer",
	})
	if wrongInventoryGrant.Code != http.StatusNotFound {
		t.Fatalf("expected wrong inventory grant status %d, got %d with body %s", http.StatusNotFound, wrongInventoryGrant.Code, wrongInventoryGrant.Body.String())
	}
	assertSafeError(t, wrongInventoryGrant, "resource_not_found", "Resource not found.")

	unauthenticatedRevoke := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants/viewer-user/viewer", "", nil)
	if unauthenticatedRevoke.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated revoke status %d, got %d with body %s", http.StatusUnauthorized, unauthenticatedRevoke.Code, unauthenticatedRevoke.Body.String())
	}
	assertSafeError(t, unauthenticatedRevoke, "authentication_required", "Authentication required.")

	malformedAuthRevoke := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants/viewer-user/viewer", "Bearer not-local-dev", nil)
	if malformedAuthRevoke.Code != http.StatusUnauthorized {
		t.Fatalf("expected malformed-auth revoke status %d, got %d with body %s", http.StatusUnauthorized, malformedAuthRevoke.Code, malformedAuthRevoke.Body.String())
	}
	assertSafeError(t, malformedAuthRevoke, "authentication_required", "Authentication required.")

	invalidRelationshipRevoke := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants/viewer-user/owner", "Bearer dev:inventory-owner", nil)
	if invalidRelationshipRevoke.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected invalid relationship revoke status %d, got %d with body %s", http.StatusUnprocessableEntity, invalidRelationshipRevoke.Code, invalidRelationshipRevoke.Body.String())
	}
	assertErrorCode(t, invalidRelationshipRevoke, "invalid_request")

	wrongTenantRevoke := performRequest(server, http.MethodDelete, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/access-grants/viewer-user/viewer", "Bearer dev:inventory-owner", nil)
	if wrongTenantRevoke.Code != http.StatusNotFound {
		t.Fatalf("expected wrong tenant revoke status %d, got %d with body %s", http.StatusNotFound, wrongTenantRevoke.Code, wrongTenantRevoke.Body.String())
	}
	assertSafeError(t, wrongTenantRevoke, "resource_not_found", "Resource not found.")

	wrongInventoryRevoke := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/01ARZ3NDEKTSV4RRFFQ69G5FB0/access-grants/viewer-user/viewer", "Bearer dev:inventory-owner", nil)
	if wrongInventoryRevoke.Code != http.StatusNotFound {
		t.Fatalf("expected wrong inventory revoke status %d, got %d with body %s", http.StatusNotFound, wrongInventoryRevoke.Code, wrongInventoryRevoke.Body.String())
	}
	assertSafeError(t, wrongInventoryRevoke, "resource_not_found", "Resource not found.")

	firstGrantPage := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?limit=1", "Bearer dev:inventory-owner", nil)
	if firstGrantPage.Code != http.StatusOK {
		t.Fatalf("expected grant list status %d, got %d with body %s", http.StatusOK, firstGrantPage.Code, firstGrantPage.Body.String())
	}
	firstPage := decodeInventoryAccessGrantList(t, firstGrantPage)
	if len(firstPage.Data) != 1 || firstPage.Meta.Pagination == nil || firstPage.Meta.Pagination.Limit != 1 || !firstPage.Meta.Pagination.HasMore || firstPage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected first grant page metadata, got %+v", firstPage)
	}
	assertInventoryAccessGrant(t, firstPage.Data[0], tenantID, inventoryID, "editor-user", "editor")

	secondGrantPage := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?limit=1&cursor="+*firstPage.Meta.Pagination.NextCursor, "Bearer dev:inventory-owner", nil)
	if secondGrantPage.Code != http.StatusOK {
		t.Fatalf("expected second grant list status %d, got %d with body %s", http.StatusOK, secondGrantPage.Code, secondGrantPage.Body.String())
	}
	secondPage := decodeInventoryAccessGrantList(t, secondGrantPage)
	if len(secondPage.Data) != 1 || secondPage.Meta.Pagination == nil || secondPage.Meta.Pagination.HasMore || secondPage.Meta.Pagination.NextCursor != nil {
		t.Fatalf("expected final grant page metadata, got %+v", secondPage)
	}
	assertInventoryAccessGrant(t, secondPage.Data[0], tenantID, inventoryID, "viewer-user", "viewer")

	allGrants := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?limit=50", "Bearer dev:inventory-owner", nil)
	if allGrants.Code != http.StatusOK {
		t.Fatalf("expected all grant list status %d, got %d with body %s", http.StatusOK, allGrants.Code, allGrants.Body.String())
	}
	allGrantBody := decodeInventoryAccessGrantList(t, allGrants)
	if len(allGrantBody.Data) != 2 {
		t.Fatalf("expected duplicate grant not to add a third grant, got %+v", allGrantBody.Data)
	}

	badCursor := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?cursor=%25%25%25", "Bearer dev:inventory-owner", nil)
	if badCursor.Code != http.StatusBadRequest {
		t.Fatalf("expected bad cursor status %d, got %d with body %s", http.StatusBadRequest, badCursor.Code, badCursor.Body.String())
	}
	assertSafeError(t, badCursor, "invalid_request", "Invalid request.")

	wrongScopeCursor := paginationCursor(map[string]any{
		"v":          1,
		"collection": "inventory_access_grants",
		"scope":      tenantID + ":" + hiddenInventoryID,
		"lastId":     "viewer-user:viewer",
	})
	wrongCursorList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?cursor="+wrongScopeCursor, "Bearer dev:inventory-owner", nil)
	if wrongCursorList.Code != http.StatusBadRequest {
		t.Fatalf("expected wrong-scope cursor status %d, got %d with body %s", http.StatusBadRequest, wrongCursorList.Code, wrongCursorList.Body.String())
	}
	assertSafeError(t, wrongCursorList, "invalid_request", "Invalid request.")

	viewerListAssets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", nil)
	if viewerListAssets.Code != http.StatusOK {
		t.Fatalf("expected viewer list assets status %d, got %d with body %s", http.StatusOK, viewerListAssets.Code, viewerListAssets.Body.String())
	}

	viewerListInventories := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories?limit=50", "Bearer dev:viewer-user", nil)
	if viewerListInventories.Code != http.StatusOK {
		t.Fatalf("expected viewer inventory list status %d, got %d with body %s", http.StatusOK, viewerListInventories.Code, viewerListInventories.Body.String())
	}
	assertInventories(t, decodeInventoryList(t, viewerListInventories),
		expectedInventory{id: inventoryID, tenantID: tenantID, name: "Tools"},
	)

	viewerCreateAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", map[string]string{
		"kind":  "item",
		"title": "Drill",
	})
	if viewerCreateAsset.Code != http.StatusForbidden {
		t.Fatalf("expected viewer create asset status %d, got %d with body %s", http.StatusForbidden, viewerCreateAsset.Code, viewerCreateAsset.Body.String())
	}
	assertSafeError(t, viewerCreateAsset, "forbidden", "Forbidden.")

	editorCreateAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:editor-user", map[string]string{
		"kind":  "item",
		"title": "Drill",
	})
	if editorCreateAsset.Code != http.StatusCreated {
		t.Fatalf("expected editor create asset status %d, got %d with body %s", http.StatusCreated, editorCreateAsset.Code, editorCreateAsset.Body.String())
	}

	editorListInventories := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories?limit=50", "Bearer dev:editor-user", nil)
	if editorListInventories.Code != http.StatusOK {
		t.Fatalf("expected editor inventory list status %d, got %d with body %s", http.StatusOK, editorListInventories.Code, editorListInventories.Body.String())
	}
	assertInventories(t, decodeInventoryList(t, editorListInventories),
		expectedInventory{id: inventoryID, tenantID: tenantID, name: "Tools"},
	)

	for _, item := range []struct {
		name          string
		principal     string
		expectedGrant string
	}{
		{name: "viewer", principal: "viewer-user", expectedGrant: "another-viewer"},
		{name: "editor", principal: "editor-user", expectedGrant: "another-editor"},
		{name: "unrelated", principal: "intruder", expectedGrant: "another-intruder"},
	} {
		t.Run(item.name+" cannot share", func(t *testing.T) {
			response := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:"+item.principal, map[string]string{
				"principalId":  item.expectedGrant,
				"relationship": "viewer",
			})
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected share status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "forbidden", "Forbidden.")
		})

		t.Run(item.name+" cannot list grants", func(t *testing.T) {
			response := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:"+item.principal, nil)
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected grant list status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "forbidden", "Forbidden.")
		})

		t.Run(item.name+" cannot revoke grants", func(t *testing.T) {
			response := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants/viewer-user/viewer", "Bearer dev:"+item.principal, nil)
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected grant revoke status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "forbidden", "Forbidden.")
		})
	}

	revokeViewer := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants/viewer-user/viewer", "Bearer dev:inventory-owner", nil)
	if revokeViewer.Code != http.StatusNoContent {
		t.Fatalf("expected revoke status %d, got %d with body %s", http.StatusNoContent, revokeViewer.Code, revokeViewer.Body.String())
	}

	revokeViewerAgain := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants/viewer-user/viewer", "Bearer dev:inventory-owner", nil)
	if revokeViewerAgain.Code != http.StatusNoContent {
		t.Fatalf("expected idempotent revoke status %d, got %d with body %s", http.StatusNoContent, revokeViewerAgain.Code, revokeViewerAgain.Body.String())
	}

	remainingGrants := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?limit=50", "Bearer dev:inventory-owner", nil)
	if remainingGrants.Code != http.StatusOK {
		t.Fatalf("expected remaining grant list status %d, got %d with body %s", http.StatusOK, remainingGrants.Code, remainingGrants.Body.String())
	}
	remainingGrantBody := decodeInventoryAccessGrantList(t, remainingGrants)
	if len(remainingGrantBody.Data) != 1 || remainingGrantBody.Data[0].PrincipalID != "editor-user" {
		t.Fatalf("expected only editor grant after viewer revoke, got %+v", remainingGrantBody.Data)
	}

	revokedViewerListAssets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", nil)
	if revokedViewerListAssets.Code != http.StatusForbidden {
		t.Fatalf("expected revoked viewer list status %d, got %d with body %s", http.StatusForbidden, revokedViewerListAssets.Code, revokedViewerListAssets.Body.String())
	}
	assertSafeError(t, revokedViewerListAssets, "forbidden", "Forbidden.")

	revokeEditor := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants/editor-user/editor", "Bearer dev:inventory-owner", nil)
	if revokeEditor.Code != http.StatusNoContent {
		t.Fatalf("expected editor revoke status %d, got %d with body %s", http.StatusNoContent, revokeEditor.Code, revokeEditor.Body.String())
	}

	noRemainingGrants := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?limit=50", "Bearer dev:inventory-owner", nil)
	if noRemainingGrants.Code != http.StatusOK {
		t.Fatalf("expected empty grant list status %d, got %d with body %s", http.StatusOK, noRemainingGrants.Code, noRemainingGrants.Body.String())
	}
	noRemainingGrantBody := decodeInventoryAccessGrantList(t, noRemainingGrants)
	if len(noRemainingGrantBody.Data) != 0 {
		t.Fatalf("expected no grants after editor revoke, got %+v", noRemainingGrantBody.Data)
	}

	revokedEditorListAssets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:editor-user", nil)
	if revokedEditorListAssets.Code != http.StatusForbidden {
		t.Fatalf("expected revoked editor list status %d, got %d with body %s", http.StatusForbidden, revokedEditorListAssets.Code, revokedEditorListAssets.Body.String())
	}
	assertSafeError(t, revokedEditorListAssets, "forbidden", "Forbidden.")

	revokedEditorCreateAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:editor-user", map[string]string{
		"kind":  "item",
		"title": "Saw",
	})
	if revokedEditorCreateAsset.Code != http.StatusForbidden {
		t.Fatalf("expected revoked editor create status %d, got %d with body %s", http.StatusForbidden, revokedEditorCreateAsset.Code, revokedEditorCreateAsset.Body.String())
	}
	assertSafeError(t, revokedEditorCreateAsset, "forbidden", "Forbidden.")
}

func TestInventoryAccessRevocationFailsClosedWhenAuthorizerRevokeFails(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	authorizer := memory.NewAuthorizer()
	server := NewServer(":0", newSeededTestAppWithAuthorizer(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
		},
		ids: []string{
			"audit-viewer-grant", "viewer-grant-event", "viewer-grant-claim",
			"audit-viewer-revoke", "viewer-revoke-event", "viewer-revoke-claim",
		},
	}, failingRevokeAuthorizer{delegate: authorizer}))

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}

	revokeViewer := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants/viewer-user/viewer", "Bearer dev:inventory-owner", nil)
	if revokeViewer.Code != http.StatusInternalServerError {
		t.Fatalf("expected revoke failure status %d, got %d with body %s", http.StatusInternalServerError, revokeViewer.Code, revokeViewer.Body.String())
	}
	assertSafeError(t, revokeViewer, "internal_error", "Internal server error.")
}

func TestInventoryAccessRevocationIgnoresUnrelatedOutboxFailures(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const poisonTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	store := memory.NewStore()
	enqueuePoisonTenantOwnerOutboxEvent(t, store, "older-poison-event", poisonTenantID)
	authorizer := memory.NewAuthorizer()
	server := NewServer(":0", newSeededTestAppWithStoreAndAuthorizer(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
		},
		ids: []string{
			"audit-viewer-grant", "viewer-grant-event", "viewer-grant-claim",
			"audit-viewer-revoke", "viewer-revoke-event", "viewer-revoke-claim",
		},
	}, store, failingTenantGrantAuthorizer{delegate: authorizer}))

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}

	revokeViewer := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants/viewer-user/viewer", "Bearer dev:inventory-owner", nil)
	if revokeViewer.Code != http.StatusNoContent {
		t.Fatalf("expected revoke status %d despite unrelated outbox failure, got %d with body %s", http.StatusNoContent, revokeViewer.Code, revokeViewer.Body.String())
	}

	revokedViewerListAssets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", nil)
	if revokedViewerListAssets.Code != http.StatusForbidden {
		t.Fatalf("expected revoked viewer list status %d, got %d with body %s", http.StatusForbidden, revokedViewerListAssets.Code, revokedViewerListAssets.Body.String())
	}
	assertSafeError(t, revokedViewerListAssets, "forbidden", "Forbidden.")
}

func TestInventoryAccessInvitationsCreateAcceptAndRevoke(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FB0"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
			{id: otherInventoryID, tenantID: tenantID, name: "Supplies", owner: "inventory-owner"},
		},
		ids: []string{
			"invite-viewer", "audit-invite-viewer",
			"audit-wrong-token-accept", "wrong-token-accept-event",
			"audit-wrong-email-accept", "wrong-email-accept-event",
			"audit-wrong-tenant-accept", "wrong-tenant-accept-event",
			"audit-wrong-inventory-accept", "wrong-inventory-accept-event",
			"audit-accept-viewer", "accept-viewer-event", "accept-viewer-claim",
			"audit-already-accepted", "already-accepted-event",
			"audit-editor-grant", "editor-grant-event", "editor-grant-claim",
			"invite-editor", "audit-invite-editor", "audit-revoke-editor",
			"audit-revoked-accept", "revoked-accept-event",
		},
	}))

	invitationResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:inventory-owner", map[string]string{
		"email":        "Viewer@Example.COM",
		"relationship": "viewer",
	})
	if invitationResponse.Code != http.StatusCreated {
		t.Fatalf("expected invitation create status %d, got %d with body %s", http.StatusCreated, invitationResponse.Code, invitationResponse.Body.String())
	}
	invitation := decodeInventoryAccessInvitation(t, invitationResponse).Data
	if invitation.Email != "viewer@example.com" || invitation.Status != "pending" || invitation.Relationship != "viewer" {
		t.Fatalf("unexpected invitation response: %+v", invitation)
	}
	if invitation.AcceptanceToken == "" {
		t.Fatalf("expected one-time acceptance token in invitation response")
	}
	if invitation.ExpiresAt == "" {
		t.Fatalf("expected invitation response to include expiresAt")
	}

	missingEmailAccept := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user", map[string]string{
		"acceptanceToken": invitation.AcceptanceToken,
	})
	if missingEmailAccept.Code != http.StatusForbidden {
		t.Fatalf("expected missing email accept status %d, got %d with body %s", http.StatusForbidden, missingEmailAccept.Code, missingEmailAccept.Body.String())
	}
	assertSafeError(t, missingEmailAccept, "forbidden", "Forbidden.")

	missingTokenAccept := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": "",
	})
	if missingTokenAccept.Code != http.StatusForbidden {
		t.Fatalf("expected missing token accept status %d, got %d with body %s", http.StatusForbidden, missingTokenAccept.Code, missingTokenAccept.Body.String())
	}
	assertSafeError(t, missingTokenAccept, "forbidden", "Forbidden.")

	wrongTokenAccept := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": "wrong-token",
	})
	if wrongTokenAccept.Code != http.StatusForbidden {
		t.Fatalf("expected wrong token accept status %d, got %d with body %s", http.StatusForbidden, wrongTokenAccept.Code, wrongTokenAccept.Body.String())
	}
	assertSafeError(t, wrongTokenAccept, "forbidden", "Forbidden.")

	wrongEmailAccept := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:wrong@example.com", map[string]string{
		"acceptanceToken": invitation.AcceptanceToken,
	})
	if wrongEmailAccept.Code != http.StatusForbidden {
		t.Fatalf("expected wrong email accept status %d, got %d with body %s", http.StatusForbidden, wrongEmailAccept.Code, wrongEmailAccept.Body.String())
	}
	assertSafeError(t, wrongEmailAccept, "forbidden", "Forbidden.")

	wrongTenantAccept := performRequest(server, http.MethodPost, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": invitation.AcceptanceToken,
	})
	if wrongTenantAccept.Code != http.StatusForbidden {
		t.Fatalf("expected wrong tenant accept status %d, got %d with body %s", http.StatusForbidden, wrongTenantAccept.Code, wrongTenantAccept.Body.String())
	}
	assertSafeError(t, wrongTenantAccept, "forbidden", "Forbidden.")

	wrongInventoryAccept := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+otherInventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": invitation.AcceptanceToken,
	})
	if wrongInventoryAccept.Code != http.StatusForbidden {
		t.Fatalf("expected wrong inventory accept status %d, got %d with body %s", http.StatusForbidden, wrongInventoryAccept.Code, wrongInventoryAccept.Body.String())
	}
	assertSafeError(t, wrongInventoryAccept, "forbidden", "Forbidden.")

	acceptResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": invitation.AcceptanceToken,
	})
	if acceptResponse.Code != http.StatusOK {
		t.Fatalf("expected invitation accept status %d, got %d with body %s", http.StatusOK, acceptResponse.Code, acceptResponse.Body.String())
	}
	accepted := decodeInventoryAccessInvitationAcceptance(t, acceptResponse).Data
	if accepted.Invitation.Status != "accepted" || accepted.Invitation.AcceptedPrincipalID != "viewer-user" {
		t.Fatalf("expected accepted invitation, got %+v", accepted.Invitation)
	}
	assertInventoryAccessGrant(t, accepted.Grant, tenantID, inventoryID, "viewer-user", "viewer")

	acceptAlreadyAccepted := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": invitation.AcceptanceToken,
	})
	if acceptAlreadyAccepted.Code != http.StatusForbidden {
		t.Fatalf("expected already accepted invite status %d, got %d with body %s", http.StatusForbidden, acceptAlreadyAccepted.Code, acceptAlreadyAccepted.Body.String())
	}
	assertSafeError(t, acceptAlreadyAccepted, "forbidden", "Forbidden.")

	viewerListAssets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", nil)
	if viewerListAssets.Code != http.StatusOK {
		t.Fatalf("expected accepted viewer list assets status %d, got %d with body %s", http.StatusOK, viewerListAssets.Code, viewerListAssets.Body.String())
	}
	viewerCreateAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", map[string]string{"kind": "item", "title": "Drill"})
	if viewerCreateAsset.Code != http.StatusForbidden {
		t.Fatalf("expected accepted viewer create asset status %d, got %d with body %s", http.StatusForbidden, viewerCreateAsset.Code, viewerCreateAsset.Body.String())
	}

	editorGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "editor-user",
		"relationship": "editor",
	})
	if editorGrant.Code != http.StatusCreated {
		t.Fatalf("expected editor grant status %d, got %d with body %s", http.StatusCreated, editorGrant.Code, editorGrant.Body.String())
	}

	for _, item := range []struct {
		name string
		auth string
	}{
		{name: "viewer", auth: "Bearer dev:viewer-user"},
		{name: "editor", auth: "Bearer dev:editor-user"},
		{name: "unrelated user", auth: "Bearer dev:unrelated-user"},
	} {
		t.Run(item.name+" cannot create invitations", func(t *testing.T) {
			response := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", item.auth, map[string]string{
				"email":        item.name + "@example.com",
				"relationship": "viewer",
			})
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected invitation create status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
		})
	}

	editorInviteResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:inventory-owner", map[string]string{
		"email":        "editor@example.com",
		"relationship": "editor",
	})
	if editorInviteResponse.Code != http.StatusCreated {
		t.Fatalf("expected editor invitation create status %d, got %d with body %s", http.StatusCreated, editorInviteResponse.Code, editorInviteResponse.Body.String())
	}
	editorInvite := decodeInventoryAccessInvitation(t, editorInviteResponse).Data

	for _, item := range []struct {
		name string
		auth string
	}{
		{name: "viewer", auth: "Bearer dev:viewer-user"},
		{name: "editor", auth: "Bearer dev:editor-user"},
		{name: "unrelated user", auth: "Bearer dev:unrelated-user"},
	} {
		t.Run(item.name+" cannot revoke invitations", func(t *testing.T) {
			response := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+editorInvite.ID, item.auth, nil)
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected invitation revoke status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
		})
	}

	revokeEditorInvite := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+editorInvite.ID, "Bearer dev:inventory-owner", nil)
	if revokeEditorInvite.Code != http.StatusNoContent {
		t.Fatalf("expected invite revoke status %d, got %d with body %s", http.StatusNoContent, revokeEditorInvite.Code, revokeEditorInvite.Body.String())
	}
	acceptRevoked := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+editorInvite.ID+"/accept", "Bearer dev:editor-user:editor@example.com", map[string]string{
		"acceptanceToken": editorInvite.AcceptanceToken,
	})
	if acceptRevoked.Code != http.StatusForbidden {
		t.Fatalf("expected revoked invite accept status %d, got %d with body %s", http.StatusForbidden, acceptRevoked.Code, acceptRevoked.Body.String())
	}
	assertSafeError(t, acceptRevoked, "forbidden", "Forbidden.")

	wrongTenantInvite := performRequest(server, http.MethodPost, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:inventory-owner", map[string]string{
		"email":        "other@example.com",
		"relationship": "viewer",
	})
	if wrongTenantInvite.Code != http.StatusNotFound {
		t.Fatalf("expected wrong tenant invite status %d, got %d with body %s", http.StatusNotFound, wrongTenantInvite.Code, wrongTenantInvite.Body.String())
	}
	assertSafeError(t, wrongTenantInvite, "resource_not_found", "Resource not found.")
}

func TestInventoryAccessInvitationRejectsExpiredToken(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
		},
		ids:           []string{"invite-viewer", "audit-invite-viewer", "audit-expired-accept", "expired-accept-event"},
		invitationTTL: time.Nanosecond,
	}))

	invitationResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:inventory-owner", map[string]string{
		"email":        "viewer@example.com",
		"relationship": "viewer",
	})
	if invitationResponse.Code != http.StatusCreated {
		t.Fatalf("expected invitation create status %d, got %d with body %s", http.StatusCreated, invitationResponse.Code, invitationResponse.Body.String())
	}
	invitation := decodeInventoryAccessInvitation(t, invitationResponse).Data
	time.Sleep(time.Millisecond)

	acceptExpired := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": invitation.AcceptanceToken,
	})
	if acceptExpired.Code != http.StatusForbidden {
		t.Fatalf("expected expired invite accept status %d, got %d with body %s", http.StatusForbidden, acceptExpired.Code, acceptExpired.Body.String())
	}
	assertSafeError(t, acceptExpired, "forbidden", "Forbidden.")
}

func TestStateCreatedDuringAuthorizationGrantFailureStaysProtected(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newTestAppWithAuthorizer(&fakeObserver{}, failingGrantAuthorizer{}, tenantID, "tenant-event", "tenant-claim", "tenant-claim-attempt"))

	createTenant := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:owner", map[string]string{"name": "Home"})
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected create tenant status %d, got %d with body %s", http.StatusCreated, createTenant.Code, createTenant.Body.String())
	}

	createInventory := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:owner", map[string]string{"name": "Tools"})
	if createInventory.Code != http.StatusForbidden {
		t.Fatalf("expected create inventory status %d, got %d with body %s", http.StatusForbidden, createInventory.Code, createInventory.Body.String())
	}
	assertSafeError(t, createInventory, "forbidden", "Forbidden.")

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories", "Bearer dev:owner", nil)
	if list.Code != http.StatusForbidden {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusForbidden, list.Code, list.Body.String())
	}
	assertSafeError(t, list, "forbidden", "Forbidden.")
}

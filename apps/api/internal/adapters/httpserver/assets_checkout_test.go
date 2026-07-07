package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAssetCheckoutEndpoints(t *testing.T) {
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
			"socket-set", "op-socket-set", "audit-socket-set",
			"checkout-socket-set", "op-checkout-socket-set", "audit-checkout-socket-set",
			"op-return-socket-set", "audit-return-socket-set",
		},
	}))
	grantEditor := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]any{"principalId": "editor-user", "relationship": "editor"})
	if grantEditor.Code != http.StatusCreated {
		t.Fatalf("expected editor grant status %d, got %d with body %s", http.StatusCreated, grantEditor.Code, grantEditor.Body.String())
	}
	grantViewer := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]any{"principalId": "viewer-user", "relationship": "viewer"})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, grantViewer.Code, grantViewer.Body.String())
	}

	create := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Socket Set",
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("expected asset create status %d, got %d with body %s", http.StatusCreated, create.Code, create.Body.String())
	}
	created := decodeAsset(t, create)
	assetPath := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + created.Data.ID

	checkout := performRequest(server, http.MethodPost, assetPath+"/checkout", "Bearer dev:owner", map[string]any{"details": "  using at my desk  "})
	if checkout.Code != http.StatusCreated {
		t.Fatalf("expected checkout status %d, got %d with body %s", http.StatusCreated, checkout.Code, checkout.Body.String())
	}
	checkedOut := decodeAssetCheckout(t, checkout)
	if checkedOut.Data.State != "open" || checkedOut.Data.CheckedOutByPrincipalID != "owner" || checkedOut.Data.CheckoutDetails != "using at my desk" {
		t.Fatalf("unexpected checkout response: %+v", checkedOut.Data)
	}

	duplicate := performRequest(server, http.MethodPost, assetPath+"/checkout", "Bearer dev:owner", nil)
	if duplicate.Code != http.StatusBadRequest {
		t.Fatalf("expected duplicate checkout status %d, got %d with body %s", http.StatusBadRequest, duplicate.Code, duplicate.Body.String())
	}

	viewerCheckout := performRequest(server, http.MethodPost, assetPath+"/checkout", "Bearer dev:viewer-user", map[string]any{})
	if viewerCheckout.Code != http.StatusForbidden {
		t.Fatalf("expected viewer checkout status %d, got %d with body %s", http.StatusForbidden, viewerCheckout.Code, viewerCheckout.Body.String())
	}

	detail := performRequest(server, http.MethodGet, assetPath, "Bearer dev:viewer-user", nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("expected detail status %d, got %d with body %s", http.StatusOK, detail.Code, detail.Body.String())
	}
	detailBody := decodeAsset(t, detail)
	if detailBody.Data.CurrentCheckout == nil || detailBody.Data.CurrentCheckout.ID != checkedOut.Data.ID {
		t.Fatalf("expected detail current checkout, got %+v", detailBody.Data.CurrentCheckout)
	}

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?limit=10", "Bearer dev:viewer-user", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, list.Code, list.Body.String())
	}
	listBody := decodeAssetList(t, list)
	if len(listBody.Data) != 1 || listBody.Data[0].CurrentCheckout == nil || listBody.Data[0].CurrentCheckout.ID != checkedOut.Data.ID {
		t.Fatalf("expected list current checkout projection, got %+v", listBody.Data)
	}

	checkedOutList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/checked-out-assets?limit=10", "Bearer dev:viewer-user", nil)
	if checkedOutList.Code != http.StatusOK {
		t.Fatalf("expected checked-out list status %d, got %d with body %s", http.StatusOK, checkedOutList.Code, checkedOutList.Body.String())
	}
	checkedOutListBody := decodeCheckedOutAssetList(t, checkedOutList)
	if len(checkedOutListBody.Data) != 1 || checkedOutListBody.Data[0].Asset.ID != created.Data.ID || checkedOutListBody.Data[0].Checkout.ID != checkedOut.Data.ID {
		t.Fatalf("unexpected checked-out list: %+v", checkedOutListBody.Data)
	}

	history := performRequest(server, http.MethodGet, assetPath+"/checkouts?limit=10", "Bearer dev:viewer-user", nil)
	if history.Code != http.StatusOK {
		t.Fatalf("expected checkout history status %d, got %d with body %s", http.StatusOK, history.Code, history.Body.String())
	}
	historyBody := decodeAssetCheckoutList(t, history)
	if len(historyBody.Data) != 1 || historyBody.Data[0].CheckoutDetails != "using at my desk" {
		t.Fatalf("unexpected checkout history: %+v", historyBody.Data)
	}

	deleteOpen := performRequest(server, http.MethodDelete, assetPath, "Bearer dev:owner", nil)
	if deleteOpen.Code != http.StatusForbidden {
		t.Fatalf("expected delete open checkout status %d, got %d with body %s", http.StatusForbidden, deleteOpen.Code, deleteOpen.Body.String())
	}

	returned := performRequest(server, http.MethodPost, assetPath+"/return", "Bearer dev:editor-user", map[string]any{"details": "back in drawer"})
	if returned.Code != http.StatusOK {
		t.Fatalf("expected return status %d, got %d with body %s", http.StatusOK, returned.Code, returned.Body.String())
	}
	returnedBody := decodeAssetCheckout(t, returned)
	if returnedBody.Data.State != "returned" || returnedBody.Data.ReturnedByPrincipalID != "editor-user" || returnedBody.Data.ReturnDetails != "back in drawer" {
		t.Fatalf("unexpected return response: %+v", returnedBody.Data)
	}

	emptyCheckedOutList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/checked-out-assets?limit=10", "Bearer dev:owner", nil)
	if emptyCheckedOutList.Code != http.StatusOK {
		t.Fatalf("expected empty checked-out list status %d, got %d with body %s", http.StatusOK, emptyCheckedOutList.Code, emptyCheckedOutList.Body.String())
	}
	if len(decodeCheckedOutAssetList(t, emptyCheckedOutList).Data) != 0 {
		t.Fatalf("expected no checked-out assets after return")
	}
}

func TestAssetCheckoutEndpointsRejectUnauthorizedAndCrossScopeAccess(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FB1"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FB2"
	const sameTenantOtherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FB3"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
			{id: sameTenantOtherInventoryID, tenantID: tenantID, name: "Garden", owner: "owner"},
			{id: otherInventoryID, tenantID: otherTenantID, name: "Cabin Gear", owner: "other-owner"},
		},
		ids: []string{
			"checkout-security-asset", "op-checkout-security-asset", "audit-checkout-security-asset",
			"checkout-security-record", "op-checkout-security-record", "audit-checkout-security-record",
			"checkout-smuggle-asset", "op-checkout-smuggle-asset", "audit-checkout-smuggle-asset",
			"checkout-smuggle-record", "op-checkout-smuggle-record", "audit-checkout-smuggle-record",
			"op-return-smuggle-record", "audit-return-smuggle-record",
		},
	}))
	requireStatus(t, performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]any{"principalId": "viewer-user", "relationship": "viewer"}), http.StatusCreated)

	create := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Checkout Security Asset",
	})
	requireStatus(t, create, http.StatusCreated)
	created := decodeAsset(t, create)
	assetPath := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + created.Data.ID
	requireStatus(t, performRequest(server, http.MethodPost, assetPath+"/checkout", "Bearer dev:owner", map[string]any{"details": "borrowed"}), http.StatusCreated)

	mutationCases := []struct {
		name          string
		method        string
		path          string
		authorization string
		body          any
		status        int
		code          string
		message       string
	}{
		{name: "checkout missing auth", method: http.MethodPost, path: assetPath + "/checkout", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "checkout malformed auth", method: http.MethodPost, path: assetPath + "/checkout", authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "return missing auth", method: http.MethodPost, path: assetPath + "/return", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "return malformed auth", method: http.MethodPost, path: assetPath + "/return", authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "viewer return", method: http.MethodPost, path: assetPath + "/return", authorization: "Bearer dev:viewer-user", status: http.StatusForbidden, code: "forbidden", message: "Forbidden."},
		{name: "intruder checkout", method: http.MethodPost, path: assetPath + "/checkout", authorization: "Bearer dev:intruder", status: http.StatusForbidden, code: "forbidden", message: "Forbidden."},
		{name: "cross tenant checkout", method: http.MethodPost, path: "/tenants/" + otherTenantID + "/inventories/" + otherInventoryID + "/assets/" + created.Data.ID + "/checkout", authorization: "Bearer dev:other-owner", status: http.StatusNotFound, code: "resource_not_found", message: "Resource not found."},
		{name: "wrong inventory checkout", method: http.MethodPost, path: "/tenants/" + tenantID + "/inventories/" + sameTenantOtherInventoryID + "/assets/" + created.Data.ID + "/checkout", authorization: "Bearer dev:owner", status: http.StatusNotFound, code: "resource_not_found", message: "Resource not found."},
		{name: "wrong inventory return", method: http.MethodPost, path: "/tenants/" + tenantID + "/inventories/" + sameTenantOtherInventoryID + "/assets/" + created.Data.ID + "/return", authorization: "Bearer dev:owner", status: http.StatusBadRequest, code: "invalid_request", message: "Invalid request."},
	}
	for _, tc := range mutationCases {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.body
			if body == nil {
				body = map[string]any{}
			}
			response := performRequest(server, tc.method, tc.path, tc.authorization, body)
			if response.Code != tc.status {
				t.Fatalf("expected status %d, got %d with body %s", tc.status, response.Code, response.Body.String())
			}
			assertSafeError(t, response, tc.code, tc.message)
		})
	}

	readCases := []struct {
		name          string
		path          string
		authorization string
		status        int
		code          string
		message       string
	}{
		{name: "history missing auth", path: assetPath + "/checkouts?limit=10", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "history malformed auth", path: assetPath + "/checkouts?limit=10", authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "history intruder", path: assetPath + "/checkouts?limit=10", authorization: "Bearer dev:intruder", status: http.StatusForbidden, code: "forbidden", message: "Forbidden."},
		{name: "history cross tenant", path: "/tenants/" + otherTenantID + "/inventories/" + otherInventoryID + "/assets/" + created.Data.ID + "/checkouts?limit=10", authorization: "Bearer dev:other-owner", status: http.StatusNotFound, code: "resource_not_found", message: "Resource not found."},
		{name: "checked-out list missing auth", path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/checked-out-assets?limit=10", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "checked-out list malformed auth", path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/checked-out-assets?limit=10", authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "checked-out list intruder", path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/checked-out-assets?limit=10", authorization: "Bearer dev:intruder", status: http.StatusForbidden, code: "forbidden", message: "Forbidden."},
		{name: "asset detail checkout projection missing auth", path: assetPath, status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "asset detail checkout projection malformed auth", path: assetPath, authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "asset list checkout projection missing auth", path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets?limit=10", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "asset list checkout projection malformed auth", path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets?limit=10", authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "search checkout projection missing auth", path: "/tenants/" + tenantID + "/search/assets?q=Checkout&inventoryId=" + inventoryID + "&checkoutState=checked_out", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "search checkout projection malformed auth", path: "/tenants/" + tenantID + "/search/assets?q=Checkout&inventoryId=" + inventoryID + "&checkoutState=checked_out", authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
	}
	for _, tc := range readCases {
		t.Run(tc.name, func(t *testing.T) {
			response := performRequest(server, http.MethodGet, tc.path, tc.authorization, nil)
			if response.Code != tc.status {
				t.Fatalf("expected status %d, got %d with body %s", tc.status, response.Code, response.Body.String())
			}
			assertSafeError(t, response, tc.code, tc.message)
		})
	}

	otherTenantList := performRequest(server, http.MethodGet, "/tenants/"+otherTenantID+"/inventories/"+otherInventoryID+"/checked-out-assets?limit=10", "Bearer dev:other-owner", nil)
	requireStatus(t, otherTenantList, http.StatusOK)
	if len(decodeCheckedOutAssetList(t, otherTenantList).Data) != 0 {
		t.Fatalf("expected other tenant checked-out list to be empty")
	}
	wrongInventoryList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+sameTenantOtherInventoryID+"/checked-out-assets?limit=10", "Bearer dev:owner", nil)
	requireStatus(t, wrongInventoryList, http.StatusOK)
	if len(decodeCheckedOutAssetList(t, wrongInventoryList).Data) != 0 {
		t.Fatalf("expected wrong inventory checked-out list to be empty")
	}

	smuggleCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Smuggle Target",
	})
	requireStatus(t, smuggleCreate, http.StatusCreated)
	smuggleAsset := decodeAsset(t, smuggleCreate)
	smugglePath := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + smuggleAsset.Data.ID
	smuggledCheckout := performRequest(server, http.MethodPost, smugglePath+"/checkout", "Bearer dev:owner", map[string]any{
		"details":                 "using at desk",
		"checkedOutByPrincipalId": "attacker",
		"returnedByPrincipalId":   "attacker",
		"principalId":             "attacker",
		"role":                    "owner",
		"permissions":             []string{"inventory.edit_asset"},
	})
	if smuggledCheckout.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected smuggled checkout status %d, got %d with body %s", http.StatusUnprocessableEntity, smuggledCheckout.Code, smuggledCheckout.Body.String())
	}
	assertValidationError(t, smuggledCheckout)
	smuggleDetail := performRequest(server, http.MethodGet, smugglePath, "Bearer dev:owner", nil)
	requireStatus(t, smuggleDetail, http.StatusOK)
	if decodeAsset(t, smuggleDetail).Data.CurrentCheckout != nil {
		t.Fatalf("expected rejected smuggled checkout not to mutate asset")
	}
	requireStatus(t, performRequest(server, http.MethodPost, smugglePath+"/checkout", "Bearer dev:owner", map[string]any{"details": "using at desk"}), http.StatusCreated)
	smuggledReturn := performRequest(server, http.MethodPost, smugglePath+"/return", "Bearer dev:owner", map[string]any{
		"details":               "back",
		"returnedByPrincipalId": "attacker",
		"principalId":           "attacker",
		"role":                  "owner",
		"permissions":           []string{"inventory.edit_asset"},
	})
	if smuggledReturn.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected smuggled return status %d, got %d with body %s", http.StatusUnprocessableEntity, smuggledReturn.Code, smuggledReturn.Body.String())
	}
	assertValidationError(t, smuggledReturn)
	smuggleDetail = performRequest(server, http.MethodGet, smugglePath, "Bearer dev:owner", nil)
	requireStatus(t, smuggleDetail, http.StatusOK)
	if decodeAsset(t, smuggleDetail).Data.CurrentCheckout == nil {
		t.Fatalf("expected rejected smuggled return not to close checkout")
	}
}

func assertValidationError(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	var body errorResponse
	decodeBody(t, response, &body)
	if body.Error.Code != "invalid_request" || body.Error.Message != "validation failed" {
		t.Fatalf("expected validation error, got %+v", body.Error)
	}
}

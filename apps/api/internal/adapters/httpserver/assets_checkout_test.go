package httpserver

import (
	"net/http"
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

	checkedOutList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/checked-out?limit=10", "Bearer dev:viewer-user", nil)
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

	emptyCheckedOutList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/checked-out?limit=10", "Bearer dev:owner", nil)
	if emptyCheckedOutList.Code != http.StatusOK {
		t.Fatalf("expected empty checked-out list status %d, got %d with body %s", http.StatusOK, emptyCheckedOutList.Code, emptyCheckedOutList.Body.String())
	}
	if len(decodeCheckedOutAssetList(t, emptyCheckedOutList).Data) != 0 {
		t.Fatalf("expected no checked-out assets after return")
	}
}

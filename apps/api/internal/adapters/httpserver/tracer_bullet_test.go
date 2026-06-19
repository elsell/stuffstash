package httpserver

import (
	"bytes"
	"net/http"
	"testing"
)

func TestSecureTenantInventoryFlow(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"

	server := NewServer(":0", newTestApp(&fakeObserver{},
		tenantID, "tenant-event", "audit-tenant", "tenant-claim",
		inventoryID, "inventory-event", "audit-inventory", "inventory-claim",
		"location-one", "audit-location", "asset-one", "audit-asset",
	))

	me := performRequest(server, http.MethodGet, "/me", "Bearer dev:user-one", nil)
	if me.Code != http.StatusOK {
		t.Fatalf("expected /me status %d, got %d with body %s", http.StatusOK, me.Code, me.Body.String())
	}

	createTenant := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:user-one", map[string]string{"name": "Home"})
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected create tenant status %d, got %d with body %s", http.StatusCreated, createTenant.Code, createTenant.Body.String())
	}

	var tenantBody struct {
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	decodeBody(t, createTenant, &tenantBody)
	if tenantBody.Data.ID != tenantID {
		t.Fatalf("expected tenant ID %q, got %q", tenantID, tenantBody.Data.ID)
	}

	createInventory := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:user-one", map[string]string{"name": "Tools"})
	if createInventory.Code != http.StatusCreated {
		t.Fatalf("expected create inventory status %d, got %d with body %s", http.StatusCreated, createInventory.Code, createInventory.Body.String())
	}

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories", "Bearer dev:user-one", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, list.Code, list.Body.String())
	}

	var listBody struct {
		Data []struct {
			ID       string `json:"id"`
			TenantID string `json:"tenantId"`
			Name     string `json:"name"`
		} `json:"data"`
	}
	decodeBody(t, list, &listBody)
	if len(listBody.Data) != 1 {
		t.Fatalf("expected 1 inventory, got %d", len(listBody.Data))
	}
	if listBody.Data[0].ID != inventoryID {
		t.Fatalf("expected inventory ID %q, got %q", inventoryID, listBody.Data[0].ID)
	}

	createLocation := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:user-one", map[string]string{
		"kind":  "location",
		"title": "Garage",
	})
	if createLocation.Code != http.StatusCreated {
		t.Fatalf("expected create location status %d, got %d with body %s", http.StatusCreated, createLocation.Code, createLocation.Body.String())
	}
	locationBody := decodeAsset(t, createLocation)
	if locationBody.Data.Kind != "location" || locationBody.Data.LifecycleState != "active" {
		t.Fatalf("unexpected location response: %+v", locationBody.Data)
	}

	createAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:user-one", map[string]string{
		"kind":          "item",
		"title":         "Drill",
		"description":   "Cordless",
		"parentAssetId": locationBody.Data.ID,
	})
	if createAsset.Code != http.StatusCreated {
		t.Fatalf("expected create asset status %d, got %d with body %s", http.StatusCreated, createAsset.Code, createAsset.Body.String())
	}

	assets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:user-one", nil)
	if assets.Code != http.StatusOK {
		t.Fatalf("expected list assets status %d, got %d with body %s", http.StatusOK, assets.Code, assets.Body.String())
	}
	assetList := decodeAssetList(t, assets)
	if len(assetList.Data) != 2 {
		t.Fatalf("expected 2 assets, got %+v", assetList.Data)
	}

	firstAssetPage := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?limit=1", "Bearer dev:user-one", nil)
	if firstAssetPage.Code != http.StatusOK {
		t.Fatalf("expected first asset page status %d, got %d with body %s", http.StatusOK, firstAssetPage.Code, firstAssetPage.Body.String())
	}
	firstPage := decodeAssetList(t, firstAssetPage)
	if len(firstPage.Data) != 1 || firstPage.Meta.Pagination == nil || firstPage.Meta.Pagination.Limit != 1 || !firstPage.Meta.Pagination.HasMore || firstPage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected first paginated asset page, got %+v", firstPage)
	}

	secondAssetPage := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?limit=1&cursor="+*firstPage.Meta.Pagination.NextCursor, "Bearer dev:user-one", nil)
	if secondAssetPage.Code != http.StatusOK {
		t.Fatalf("expected second asset page status %d, got %d with body %s", http.StatusOK, secondAssetPage.Code, secondAssetPage.Body.String())
	}
	if !bytes.Contains(secondAssetPage.Body.Bytes(), []byte(`"nextCursor":null`)) {
		t.Fatalf("expected final asset page to include null nextCursor, got %s", secondAssetPage.Body.String())
	}
	secondPage := decodeAssetList(t, secondAssetPage)
	if len(secondPage.Data) != 1 || secondPage.Meta.Pagination == nil || secondPage.Meta.Pagination.Limit != 1 || secondPage.Meta.Pagination.HasMore || secondPage.Meta.Pagination.NextCursor != nil {
		t.Fatalf("expected final paginated asset page, got %+v", secondPage)
	}
}

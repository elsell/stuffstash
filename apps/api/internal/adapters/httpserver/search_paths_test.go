package httpserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestSearchReturnsAuthorizedAncestorPath(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const visibleInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "owner"}},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"}, {id: visibleInventoryID, tenantID: tenantID, name: "Visible", owner: "owner"}},
		ids:         []string{"garage", "audit-garage", "shelf", "audit-shelf", "drill", "audit-drill"},
	}))
	parent := ""
	for _, item := range []struct{ kind, title string }{{"location", "Garage"}, {"container", "Shelf"}, {"item", "Drill"}} {
		body := map[string]string{"kind": item.kind, "title": item.title}
		if parent != "" {
			body["parentAssetId"] = parent
		}
		response := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", body)
		if response.Code != http.StatusCreated {
			t.Fatalf("create: %d %s", response.Code, response.Body.String())
		}
		parent = decodeAsset(t, response).Data.ID
	}
	result := searchAssets(t, server, tenantID, "Bearer dev:owner", "Drill", "", "", "")
	if len(result.Data) != 1 || len(result.Data[0].AncestorPath) != 2 {
		t.Fatalf("expected complete path, got %+v", result.Data)
	}
	path := result.Data[0].AncestorPath
	if path[0].Title != "Garage" || path[1].Title != "Shelf" || path[1].ID != result.Data[0].Asset.ParentAssetID {
		t.Fatalf("incorrect path: %+v", path)
	}
	root := searchAssets(t, server, tenantID, "Bearer dev:owner", "Garage", "", "", "")
	if len(root.Data) != 1 || root.Data[0].AncestorPath == nil || len(root.Data[0].AncestorPath) != 0 {
		t.Fatalf("root must have explicit empty path: %+v", root.Data)
	}
	for _, token := range []string{"", "Bearer dev:stranger"} {
		response := searchAssetsResponse(server, tenantID, inventoryID, token, "Drill", "", "", "", 0, "")
		if response.Code != http.StatusUnauthorized && response.Code != http.StatusForbidden {
			t.Fatalf("unauthorized path read: %d %s", response.Code, response.Body.String())
		}
	}

	visibleParent := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+visibleInventoryID+"/assets", "Bearer dev:owner", map[string]string{"kind": "location", "title": "Office"})
	if visibleParent.Code != http.StatusCreated {
		t.Fatalf("visible parent: %s", visibleParent.Body.String())
	}
	visibleChild := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+visibleInventoryID+"/assets", "Bearer dev:owner", map[string]string{"kind": "item", "title": "Pen", "parentAssetId": decodeAsset(t, visibleParent).Data.ID})
	if visibleChild.Code != http.StatusCreated {
		t.Fatalf("visible child: %s", visibleChild.Body.String())
	}
	grant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+visibleInventoryID+"/access-grants", "Bearer dev:owner", map[string]string{"principalId": "viewer", "relationship": "viewer"})
	if grant.Code != http.StatusCreated {
		t.Fatalf("grant: %d %s", grant.Code, grant.Body.String())
	}
	for _, scope := range []string{"", inventoryID} {
		hidden := searchAssetsResponse(server, tenantID, scope, "Bearer dev:viewer", "Drill", "", "", "", 0, "")
		if hidden.Code != http.StatusOK {
			t.Fatalf("viewer search: %d %s", hidden.Code, hidden.Body.String())
		}
		if len(decodeAssetSearch(t, hidden).Data) != 0 || strings.Contains(hidden.Body.String(), "Garage") || strings.Contains(hidden.Body.String(), "Shelf") {
			t.Fatalf("hidden path leaked: %s", hidden.Body.String())
		}
	}

	visible := searchAssets(t, server, tenantID, "Bearer dev:viewer", "Pen", "", "", "")
	if len(visible.Data) != 1 || len(visible.Data[0].AncestorPath) != 1 || visible.Data[0].AncestorPath[0].Title != "Office" {
		t.Fatalf("viewer lost authorized path: %+v", visible.Data)
	}

}

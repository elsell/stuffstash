package httpserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestAssetKindEditAuthorizationAndAtomicConversion(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{tenants: []seedTenant{{id: tenantID, name: "Home", owner: "owner"}}, inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Main", owner: "owner"}}}))
	base := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets"
	created := performRequest(server, http.MethodPost, base, "Bearer dev:owner", map[string]string{"kind": "location", "title": "Bin 60"})
	if created.Code != 201 {
		t.Fatal(created.Body.String())
	}
	id := decodeAsset(t, created).Data.ID
	wrongScope := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAX/assets/"+id, "Bearer dev:owner", map[string]string{"kind": "container"})
	if wrongScope.Code != 403 && wrongScope.Code != 404 {
		t.Fatalf("cross inventory: %d", wrongScope.Code)
	}
	wrongTenant := performRequest(server, http.MethodPatch, "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAX/inventories/"+inventoryID+"/assets/"+id, "Bearer dev:owner", map[string]string{"kind": "container"})
	if wrongTenant.Code != 403 && wrongTenant.Code != 404 {
		t.Fatalf("cross tenant: %d", wrongTenant.Code)
	}

	path := base + "/" + id
	grant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]string{"principalId": "viewer", "relationship": "viewer"})
	if grant.Code != 201 {
		t.Fatal(grant.Body.String())
	}
	for _, token := range []string{"", "Bearer broken", "Bearer dev:outsider", "Bearer dev:viewer"} {
		denied := performRequest(server, http.MethodPatch, path, token, map[string]string{"kind": "container"})
		if denied.Code != 401 && denied.Code != 403 {
			t.Fatalf("unauthorized edit: %d %s", denied.Code, denied.Body.String())
		}
	}
	changed := performRequest(server, http.MethodPatch, path, "Bearer dev:owner", map[string]string{"kind": "container", "title": "Portable bin"})
	if changed.Code != 200 || decodeAsset(t, changed).Data.Kind != "container" {
		t.Fatalf("conversion: %s", changed.Body.String())
	}

	for _, kind := range []string{"item", "location", "container"} {
		result := performRequest(server, http.MethodPatch, path, "Bearer dev:owner", map[string]string{"kind": kind})
		if result.Code != 200 {
			t.Fatal(result.Body.String())
		}
		body := decodeAsset(t, result)
		if body.Data.Kind != kind || body.Data.ID != id {
			t.Fatalf("empty conversion %s: %s", kind, result.Body.String())
		}
	}
	child := performRequest(server, http.MethodPost, base, "Bearer dev:owner", map[string]string{"kind": "item", "title": "Nails", "parentAssetId": id})
	if child.Code != 201 {
		t.Fatal(child.Body.String())
	}
	childID := decodeAsset(t, child).Data.ID
	for _, kind := range []string{"location", "container"} {
		result := performRequest(server, http.MethodPatch, path, "Bearer dev:owner", map[string]string{"kind": kind})
		if result.Code != 200 || decodeAsset(t, result).Data.Kind != kind {
			t.Fatalf("conversion retaining contents: %s", result.Body.String())
		}
	}
	childAfter := performRequest(server, http.MethodGet, base+"/"+childID, "Bearer dev:owner", nil)
	if decodeAsset(t, childAfter).Data.ParentAssetID != id {
		t.Fatal("lost child placement")
	}

	archived := performRequest(server, http.MethodPatch, base+"/"+childID+"/archive", "Bearer dev:owner", nil)
	if archived.Code != 200 {
		t.Fatal(archived.Body.String())
	}
	rejected := performRequest(server, http.MethodPatch, path, "Bearer dev:owner", map[string]string{"kind": "item", "title": "Must not save"})
	if rejected.Code != 409 || !strings.Contains(rejected.Body.String(), "including archived assets") {
		t.Fatalf("archived contents must prevent conversion: %d %s", rejected.Code, rejected.Body.String())
	}
	current := performRequest(server, http.MethodGet, path, "Bearer dev:owner", nil)
	if decodeAsset(t, current).Data.Title != "Portable bin" {
		t.Fatal("partial edit committed")
	}
	checkedOut := performRequest(server, http.MethodPost, path+"/checkout", "Bearer dev:owner", map[string]string{"details": "Taking the bin"})
	if checkedOut.Code != 200 && checkedOut.Code != 201 {
		t.Fatal(checkedOut.Body.String())
	}
	rejected = performRequest(server, http.MethodPatch, path, "Bearer dev:owner", map[string]string{"kind": "location"})
	if rejected.Code != 409 || !strings.Contains(rejected.Body.String(), "Return this asset") {
		t.Fatalf("open checkout must block location: %d %s", rejected.Code, rejected.Body.String())
	}
	invalid := performRequest(server, http.MethodPatch, path, "Bearer dev:owner", map[string]string{"kind": "unknown"})
	if invalid.Code != 422 {
		t.Fatalf("invalid kind: %d", invalid.Code)
	}

}

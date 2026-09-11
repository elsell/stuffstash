package httpserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAssetExpirationPreservesPrecisionAndClears(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "owner"}},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Home", owner: "owner"}},
		ids:         []string{"01ARZ3NDEKTSV4RRFFQ69G5FAX", "audit-type", "01ARZ3NDEKTSV4RRFFQ69G5FAY", "audit-item"},
	}))
	base := "/tenants/" + tenantID + "/inventories/" + inventoryID
	createType := performRequest(server, http.MethodPost, base+"/custom-asset-types", "Bearer dev:owner", map[string]any{"key": "medicine", "displayName": "Medicine", "expirationEnabled": true})
	if createType.Code != http.StatusCreated {
		t.Fatal(createType.Body.String())
	}
	typeID := decodeCustomAssetType(t, createType).Data.ID
	result := performRequest(server, http.MethodPost, base+"/assets", "Bearer dev:owner", map[string]any{"kind": "item", "title": "Ibuprofen", "customAssetTypeId": typeID, "expiration": map[string]any{"date": "2028-02", "precision": "month"}})
	if result.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", result.Code, result.Body.String())
	}
	var response struct {
		Data struct {
			ID         string `json:"id"`
			Expiration *struct {
				Date      string `json:"date"`
				Precision string `json:"precision"`
			} `json:"expiration"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Expiration == nil || response.Data.Expiration.Date != "2028-02" || response.Data.Expiration.Precision != "month" {
		t.Fatal("month precision lost")
	}
	endpoint := base + "/assets/" + response.Data.ID
	for _, auth := range []string{"", "Bearer malformed", "Bearer dev:outsider"} {
		denied := performRequest(server, http.MethodPatch, endpoint, auth, map[string]any{"expiration": nil})
		if denied.Code != http.StatusUnauthorized && denied.Code != http.StatusForbidden {
			t.Fatalf("unauthorized edit: %d", denied.Code)
		}
	}
	for _, value := range []any{map[string]any{"date": "2027-02-29", "precision": "day"}, map[string]any{"date": "2028-02-29", "precision": "month"}} {
		invalid := performRequest(server, http.MethodPatch, endpoint, "Bearer dev:owner", map[string]any{"expiration": value})
		if invalid.Code != http.StatusBadRequest && invalid.Code != http.StatusUnprocessableEntity {
			t.Fatalf("invalid date accepted: %d", invalid.Code)
		}
	}
	for _, step := range []struct {
		body map[string]any
		date string
	}{
		{map[string]any{"title": "Ibuprofen bottle"}, "2028-02"},
		{map[string]any{"expiration": map[string]any{"date": "2028-02-29", "precision": "day"}}, "2028-02-29"},
		{map[string]any{"expiration": nil}, ""},
	} {
		edited := performRequest(server, http.MethodPatch, endpoint, "Bearer dev:owner", step.body)
		if edited.Code != http.StatusOK {
			t.Fatalf("edit: %d %s", edited.Code, edited.Body.String())
		}
		read := performRequest(server, http.MethodGet, endpoint, "Bearer dev:owner", nil)
		if read.Code != http.StatusOK {
			t.Fatalf("read: %d", read.Code)
		}
		if err := json.Unmarshal(read.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if step.date == "" {
			if response.Data.Expiration != nil {
				t.Fatal("expiration not cleared")
			}
		} else if response.Data.Expiration == nil || response.Data.Expiration.Date != step.date {
			t.Fatal("expiration not retained")
		}
	}
	untyped := performRequest(server, http.MethodPost, base+"/assets", "Bearer dev:owner", map[string]any{"kind": "item", "title": "Existing medicine"})
	if untyped.Code != http.StatusCreated {
		t.Fatal(untyped.Body.String())
	}
	if err := json.Unmarshal(untyped.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	assigned := performRequest(server, http.MethodPatch, base+"/assets/"+response.Data.ID, "Bearer dev:owner", map[string]any{"customAssetTypeId": typeID, "expiration": map[string]any{"date": "2029-04", "precision": "month"}})
	if assigned.Code != http.StatusOK {
		t.Fatalf("initial type assignment: %d %s", assigned.Code, assigned.Body.String())
	}

	var assignment struct {
		Data struct {
			OperationID string `json:"undoableOperationId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(assigned.Body.Bytes(), &assignment); err != nil {
		t.Fatal(err)
	}
	if assignment.Data.OperationID == "" {
		t.Fatal("assignment not undoable")
	}
	undoBase := base + "/undoable-operations/" + assignment.Data.OperationID
	undone := performRequest(server, http.MethodPost, undoBase+"/undo", "Bearer dev:owner", nil)
	if undone.Code != http.StatusOK {
		t.Fatalf("assignment undo: %d %s", undone.Code, undone.Body.String())
	}
	archived := performRequest(server, http.MethodPatch, base+"/custom-asset-types/"+typeID+"/archive", "Bearer dev:owner", nil)
	if archived.Code != http.StatusOK {
		t.Fatal(archived.Body.String())
	}
	redo := performRequest(server, http.MethodPost, undoBase+"/redo", "Bearer dev:owner", nil)
	if redo.Code != http.StatusBadRequest {
		t.Fatalf("redo assigned archived type: %d %s", redo.Code, redo.Body.String())
	}

}

func TestExpirationRejectsUnavailableTypesAndCrossTenantEdits(t *testing.T) {
	const tenantA = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const tenantB = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const inventoryA = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const inventoryB = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants:     []seedTenant{{id: tenantA, name: "A", owner: "a"}, {id: tenantB, name: "B", owner: "b"}},
		inventories: []seedInventory{{id: inventoryA, tenantID: tenantA, name: "A", owner: "a"}, {id: inventoryB, tenantID: tenantB, name: "B", owner: "b"}},
		ids:         []string{"01ARZ3NDEKTSV4RRFFQ69G5FAZ", "audit-type"},
	}))
	baseA := "/tenants/" + tenantA + "/inventories/" + inventoryA
	baseB := "/tenants/" + tenantB + "/inventories/" + inventoryB
	created := performRequest(server, http.MethodPost, baseA+"/custom-asset-types", "Bearer dev:a", map[string]any{"key": "medicine", "displayName": "Medicine"})
	if created.Code != http.StatusCreated {
		t.Fatal(created.Body.String())
	}
	typeID := decodeCustomAssetType(t, created).Data.ID
	date := map[string]any{"date": "2028-02", "precision": "month"}
	for _, typeValue := range []string{"", typeID} {
		denied := performRequest(server, http.MethodPost, baseA+"/assets", "Bearer dev:a", map[string]any{"kind": "item", "title": "Bottle", "customAssetTypeId": typeValue, "expiration": date})
		if denied.Code != http.StatusBadRequest {
			t.Fatalf("expiration on missing/disabled type: %d", denied.Code)
		}
	}
	enabled := performRequest(server, http.MethodPatch, baseA+"/custom-asset-types/"+typeID, "Bearer dev:a", map[string]any{"expirationEnabled": true})
	if enabled.Code != http.StatusOK {
		t.Fatal(enabled.Body.String())
	}
	crossType := performRequest(server, http.MethodPost, baseB+"/assets", "Bearer dev:b", map[string]any{"kind": "item", "title": "Bottle", "customAssetTypeId": typeID, "expiration": date})
	if crossType.Code != http.StatusNotFound && crossType.Code != http.StatusForbidden {
		t.Fatalf("cross tenant type accepted: %d", crossType.Code)
	}
	item := performRequest(server, http.MethodPost, baseA+"/assets", "Bearer dev:a", map[string]any{"kind": "item", "title": "Bottle", "customAssetTypeId": typeID, "expiration": date})
	if item.Code != http.StatusCreated {
		t.Fatal(item.Body.String())
	}
	var response struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(item.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	denied := performRequest(server, http.MethodPatch, baseB+"/assets/"+response.Data.ID, "Bearer dev:b", map[string]any{"expiration": nil})
	if denied.Code != http.StatusNotFound && denied.Code != http.StatusForbidden {
		t.Fatalf("cross tenant asset edit: %d", denied.Code)
	}
}

func TestExpirationResponseSchemaAllowsUnsetDate(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "unused-id"))
	result := performRequest(server, http.MethodGet, "/openapi.json", "", nil)
	var schema struct {
		Components struct {
			Schemas map[string]struct {
				Properties map[string]struct {
					AnyOf []struct {
						Type string `json:"type"`
					} `json:"anyOf"`
				} `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(result.Body.Bytes(), &schema); err != nil {
		t.Fatal(err)
	}
	for _, option := range schema.Components.Schemas["AssetResponse"].Properties["expiration"].AnyOf {
		if option.Type == "null" {
			return
		}
	}
	t.Fatal("expiration response schema rejects actual null response for unset date")
}

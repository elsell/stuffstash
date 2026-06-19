package httpserver

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestAssetSearchFindsMetadata(t *testing.T) {
	fixture := newAssetSearchFixture(t)

	attachmentSearch := searchAssets(t, fixture.server, fixture.tenantID, "Bearer dev:owner", "warranty", "", "", "")
	if len(attachmentSearch.Data) != 1 || attachmentSearch.Data[0].Asset.Title != "Cordless Drill" || attachmentSearch.Data[0].Matches[0].Field != "attachment_file_name" {
		t.Fatalf("expected attachment metadata search to find drill, got %+v", attachmentSearch.Data)
	}

	customFieldSearch := searchAssets(t, fixture.server, fixture.tenantID, "Bearer dev:owner", "2027", "", "", "")
	if len(customFieldSearch.Data) != 1 || customFieldSearch.Data[0].Asset.Title != "Aspirin" || customFieldSearch.Data[0].Matches[0].Field != "custom_field" {
		t.Fatalf("expected custom field search to find aspirin, got %+v", customFieldSearch.Data)
	}

	exactTypeSearch := searchAssets(t, fixture.server, fixture.tenantID, "Bearer dev:owner", "Medicine", "exact", fixture.medicineTypeID, "")
	if len(exactTypeSearch.Data) != 1 || exactTypeSearch.Data[0].Asset.Title != "Aspirin" || exactTypeSearch.Data[0].Type != "asset" {
		t.Fatalf("expected exact custom asset type search to find aspirin, got %+v", exactTypeSearch.Data)
	}
}

func TestAssetSearchPaginatesAndRejectsWrongCursorScope(t *testing.T) {
	fixture := newAssetSearchFixture(t)

	firstPage := searchAssetsWithLimit(t, fixture.server, fixture.tenantID, "Bearer dev:owner", "i", "", "", "", 1, "")
	if len(firstPage.Data) != 1 || firstPage.Meta.Pagination == nil || !firstPage.Meta.Pagination.HasMore || firstPage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected paginated search first page, got %+v", firstPage)
	}
	secondPage := searchAssetsWithLimit(t, fixture.server, fixture.tenantID, "Bearer dev:owner", "i", "", "", "", 1, *firstPage.Meta.Pagination.NextCursor)
	if len(secondPage.Data) != 1 || secondPage.Meta.Pagination == nil {
		t.Fatalf("expected paginated search second page, got %+v", secondPage)
	}
	if secondPage.Data[0].Asset.ID == firstPage.Data[0].Asset.ID {
		t.Fatalf("expected second page to advance, got first=%+v second=%+v", firstPage.Data, secondPage.Data)
	}

	wrongScopeCursor := searchAssetsResponse(fixture.server, fixture.tenantID, "Bearer dev:owner", "Aspirin", "", "", "", 0, *firstPage.Meta.Pagination.NextCursor)
	if wrongScopeCursor.Code != http.StatusBadRequest {
		t.Fatalf("expected wrong-scope cursor status %d, got %d with body %s", http.StatusBadRequest, wrongScopeCursor.Code, wrongScopeCursor.Body.String())
	}
	assertSafeError(t, wrongScopeCursor, "invalid_request", "Invalid request.")
}

func TestAssetSearchFiltersByAuthorization(t *testing.T) {
	fixture := newAssetSearchFixture(t)

	grantViewer := performRequest(fixture.server, http.MethodPost, "/tenants/"+fixture.tenantID+"/inventories/"+fixture.medicineInventoryID+"/access-grants", "Bearer dev:owner", map[string]any{
		"principalId":  "viewer",
		"relationship": "viewer",
	})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, grantViewer.Code, grantViewer.Body.String())
	}
	viewerDrillSearch := searchAssets(t, fixture.server, fixture.tenantID, "Bearer dev:viewer", "warranty", "", "", "")
	if len(viewerDrillSearch.Data) != 0 {
		t.Fatalf("expected viewer not to see hidden inventory result, got %+v", viewerDrillSearch.Data)
	}
	viewerAspirinSearch := searchAssets(t, fixture.server, fixture.tenantID, "Bearer dev:viewer", "Aspirin", "", "", "")
	if len(viewerAspirinSearch.Data) != 1 || viewerAspirinSearch.Data[0].Asset.Title != "Aspirin" {
		t.Fatalf("expected viewer to see granted inventory result, got %+v", viewerAspirinSearch.Data)
	}

	crossTenantSearch := searchAssetsResponse(fixture.server, fixture.otherTenantID, "Bearer dev:owner", "Cordless", "", "", "", 0, "")
	if crossTenantSearch.Code != http.StatusForbidden {
		t.Fatalf("expected cross-tenant search status %d, got %d with body %s", http.StatusForbidden, crossTenantSearch.Code, crossTenantSearch.Body.String())
	}
	assertSafeError(t, crossTenantSearch, "forbidden", "Forbidden.")
}

func TestAssetSearchFiltersLifecycle(t *testing.T) {
	fixture := newAssetSearchFixture(t)

	archiveDrill := performRequest(fixture.server, http.MethodPatch, "/tenants/"+fixture.tenantID+"/inventories/"+fixture.toolsInventoryID+"/assets/"+fixture.drillAssetID+"/archive", "Bearer dev:owner", nil)
	if archiveDrill.Code != http.StatusOK {
		t.Fatalf("expected archive status %d, got %d with body %s", http.StatusOK, archiveDrill.Code, archiveDrill.Body.String())
	}
	activeSearch := searchAssets(t, fixture.server, fixture.tenantID, "Bearer dev:owner", "Cordless", "", "", "")
	if assetSearchContainsTitle(activeSearch.Data, "Cordless Drill") {
		t.Fatalf("expected archived drill hidden from default active search, got %+v", activeSearch.Data)
	}
	archivedSearch := searchAssets(t, fixture.server, fixture.tenantID, "Bearer dev:owner", "Cordless", "", "", "archived")
	if len(archivedSearch.Data) != 1 || archivedSearch.Data[0].Asset.Title != "Cordless Drill" {
		t.Fatalf("expected archived search to find drill, got %+v", archivedSearch.Data)
	}
}

type assetSearchFixture struct {
	server              *http.Server
	tenantID            string
	toolsInventoryID    string
	medicineInventoryID string
	otherTenantID       string
	medicineTypeID      string
	drillAssetID        string
}

func newAssetSearchFixture(t *testing.T) assetSearchFixture {
	t.Helper()

	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const toolsInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const medicineInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAZ"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
			{id: otherTenantID, name: "Other", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: toolsInventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
			{id: medicineInventoryID, tenantID: tenantID, name: "Medicine", owner: "owner"},
			{id: otherInventoryID, tenantID: otherTenantID, name: "Other", owner: "other-owner"},
		},
		ids: []string{
			"medicine-type", "audit-medicine-type",
			"serial-field", "audit-serial-field",
			"expires-field", "audit-expires-field",
			"drill-asset", "audit-drill-asset",
			"aspirin-asset", "audit-aspirin-asset",
			"other-asset", "audit-other-asset",
			"drill-attachment", "audit-drill-attachment",
			"viewer-grant-event", "audit-viewer-grant", "viewer-claim",
			"archive-drill-audit",
		},
	}))

	createMedicineType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+medicineInventoryID+"/custom-asset-types", "Bearer dev:owner", map[string]any{
		"key":         "medicine",
		"displayName": "Medicine",
		"description": "Medication and pharmacy items.",
	})
	if createMedicineType.Code != http.StatusCreated {
		t.Fatalf("expected medicine type status %d, got %d with body %s", http.StatusCreated, createMedicineType.Code, createMedicineType.Body.String())
	}
	medicineType := decodeCustomAssetType(t, createMedicineType)

	createSerialField := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/custom-field-definitions", "Bearer dev:owner", map[string]any{
		"key":         "serial",
		"displayName": "Serial",
		"type":        "text",
	})
	if createSerialField.Code != http.StatusCreated {
		t.Fatalf("expected serial field status %d, got %d with body %s", http.StatusCreated, createSerialField.Code, createSerialField.Body.String())
	}
	createExpiresField := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+medicineInventoryID+"/custom-field-definitions", "Bearer dev:owner", map[string]any{
		"key":                "expires-on",
		"displayName":        "Expires On",
		"type":               "date",
		"applicability":      "custom_asset_types",
		"customAssetTypeIds": []string{medicineType.Data.ID},
	})
	if createExpiresField.Code != http.StatusCreated {
		t.Fatalf("expected expiration field status %d, got %d with body %s", http.StatusCreated, createExpiresField.Code, createExpiresField.Body.String())
	}

	createDrill := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+toolsInventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":        "item",
		"title":       "Cordless Drill",
		"description": "Driver kit",
		"customFields": map[string]any{
			"serial": "bag-42",
		},
	})
	if createDrill.Code != http.StatusCreated {
		t.Fatalf("expected drill status %d, got %d with body %s", http.StatusCreated, createDrill.Code, createDrill.Body.String())
	}
	drill := decodeAsset(t, createDrill)

	createAspirin := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+medicineInventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":              "item",
		"title":             "Aspirin",
		"description":       "Pain relief tablets",
		"customAssetTypeId": medicineType.Data.ID,
		"customFields": map[string]any{
			"serial":     "med-101",
			"expires-on": "2027-01-01",
		},
	})
	if createAspirin.Code != http.StatusCreated {
		t.Fatalf("expected aspirin status %d, got %d with body %s", http.StatusCreated, createAspirin.Code, createAspirin.Body.String())
	}

	createOther := performRequest(server, http.MethodPost, "/tenants/"+otherTenantID+"/inventories/"+otherInventoryID+"/assets", "Bearer dev:other-owner", map[string]any{
		"kind":  "item",
		"title": "Cordless Drill",
	})
	if createOther.Code != http.StatusCreated {
		t.Fatalf("expected other asset status %d, got %d with body %s", http.StatusCreated, createOther.Code, createOther.Body.String())
	}

	createAttachment := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+toolsInventoryID+"/assets/"+drill.Data.ID+"/attachments", "Bearer dev:owner", map[string]any{
		"fileName":      "warranty-card.png",
		"contentType":   "image/png",
		"contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent()),
	})
	if createAttachment.Code != http.StatusCreated {
		t.Fatalf("expected attachment status %d, got %d with body %s", http.StatusCreated, createAttachment.Code, createAttachment.Body.String())
	}

	return assetSearchFixture{
		server:              server,
		tenantID:            tenantID,
		toolsInventoryID:    toolsInventoryID,
		medicineInventoryID: medicineInventoryID,
		otherTenantID:       otherTenantID,
		medicineTypeID:      medicineType.Data.ID,
		drillAssetID:        drill.Data.ID,
	}
}

func TestAssetSearchRejectsMissingAndInvalidInput(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
	}))

	cases := []struct {
		name          string
		path          string
		authorization string
		status        int
		code          string
	}{
		{name: "missing auth", path: "/tenants/" + tenantID + "/search/assets?q=drill", status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "missing query", path: "/tenants/" + tenantID + "/search/assets", authorization: "Bearer dev:owner", status: http.StatusUnprocessableEntity, code: ""},
		{name: "invalid mode", path: "/tenants/" + tenantID + "/search/assets?q=drill&mode=wide", authorization: "Bearer dev:owner", status: http.StatusUnprocessableEntity, code: "invalid_request"},
		{name: "invalid lifecycle", path: "/tenants/" + tenantID + "/search/assets?q=drill&lifecycleState=deleted", authorization: "Bearer dev:owner", status: http.StatusUnprocessableEntity, code: "invalid_request"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := performRequest(server, http.MethodGet, tc.path, tc.authorization, nil)
			if response.Code != tc.status {
				t.Fatalf("expected status %d, got %d with body %s", tc.status, response.Code, response.Body.String())
			}
			if tc.code != "" {
				assertErrorCode(t, response, tc.code)
			}
		})
	}
}

func searchAssets(t *testing.T, server *http.Server, tenantID string, authorization string, query string, mode string, customAssetTypeID string, lifecycleState string) searchAssetListBody {
	t.Helper()
	return searchAssetsWithLimit(t, server, tenantID, authorization, query, mode, customAssetTypeID, lifecycleState, 0, "")
}

func searchAssetsWithLimit(t *testing.T, server *http.Server, tenantID string, authorization string, query string, mode string, customAssetTypeID string, lifecycleState string, limit int, cursor string) searchAssetListBody {
	t.Helper()
	response := searchAssetsResponse(server, tenantID, authorization, query, mode, customAssetTypeID, lifecycleState, limit, cursor)
	if response.Code != http.StatusOK {
		t.Fatalf("expected search status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}
	return decodeAssetSearch(t, response)
}

func searchAssetsResponse(server *http.Server, tenantID string, authorization string, query string, mode string, customAssetTypeID string, lifecycleState string, limit int, cursor string) *httptest.ResponseRecorder {
	values := url.Values{}
	if query != "" {
		values.Set("q", query)
	}
	if mode != "" {
		values.Set("mode", mode)
	}
	if customAssetTypeID != "" {
		values.Set("customAssetTypeId", customAssetTypeID)
	}
	if lifecycleState != "" {
		values.Set("lifecycleState", lifecycleState)
	}
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		values.Set("cursor", cursor)
	}
	path := "/tenants/" + tenantID + "/search/assets"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	return performRequest(server, http.MethodGet, path, authorization, nil)
}

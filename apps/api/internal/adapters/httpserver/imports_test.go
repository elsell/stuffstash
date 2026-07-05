package httpserver

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLegacyHomeboxImportAuthorizationAndSourceSafety(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"viewer-grant-event", "audit-viewer-grant", "viewer-claim"},
	}))
	path := "/tenants/" + tenantID + "/inventories/" + inventoryID
	grantViewer := performRequest(server, http.MethodPost, path+"/access-grants", "Bearer dev:owner", map[string]any{
		"principalId":  "viewer",
		"relationship": "viewer",
	})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, grantViewer.Code, grantViewer.Body.String())
	}
	csvBody := map[string]any{
		"sourceType":    "legacy_homebox_csv",
		"fileName":      "homebox.csv",
		"contentBase64": base64.StdEncoding.EncodeToString([]byte("HB.location,HB.asset_id,HB.name\nGarage,HB-1,Drill\n")),
	}

	viewerPreview := performRequest(server, http.MethodPost, path+"/imports/legacy-homebox/preview", "Bearer dev:viewer", csvBody)
	if viewerPreview.Code != http.StatusForbidden {
		t.Fatalf("expected viewer preview status %d, got %d with body %s", http.StatusForbidden, viewerPreview.Code, viewerPreview.Body.String())
	}
	assertSafeError(t, viewerPreview, "forbidden", "Forbidden.")

	blockedLiveSource := performRequest(server, http.MethodPost, path+"/imports/legacy-homebox/preview", "Bearer dev:owner", map[string]any{
		"sourceType": "legacy_homebox",
		"baseUrl":    "http://127.0.0.1:7744",
		"username":   "owner@example.com",
		"password":   "secret",
	})
	if blockedLiveSource.Code != http.StatusBadRequest {
		t.Fatalf("expected blocked live source status %d, got %d with body %s", http.StatusBadRequest, blockedLiveSource.Code, blockedLiveSource.Body.String())
	}
	assertImportSourceError(t, blockedLiveSource, "Homebox URL resolves to a blocked address")
}

func assertImportSourceError(t *testing.T, response *httptest.ResponseRecorder, expectedDetail string) {
	t.Helper()

	var body errorResponse
	decodeBody(t, response, &body)
	if body.Error.Code != "invalid_request" {
		t.Fatalf("expected error code invalid_request, got %q", body.Error.Code)
	}
	if body.Error.Message != "Invalid request." {
		t.Fatalf("expected generic error message, got %q", body.Error.Message)
	}
	if len(body.Error.Details) != 1 {
		t.Fatalf("expected one safe import source detail, got %+v", body.Error.Details)
	}
	detail, ok := body.Error.Details[0].(map[string]any)
	if !ok || detail["message"] != expectedDetail {
		t.Fatalf("expected safe import source detail %q, got %+v", expectedDetail, body.Error.Details)
	}
}

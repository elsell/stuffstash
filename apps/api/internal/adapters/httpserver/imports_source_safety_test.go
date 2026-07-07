package httpserver

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	appcore "github.com/stuffstash/stuff-stash/internal/app"
)

func TestDurableImportAuthorizationAndSourceSafety(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"viewer-grant-event", "audit-viewer-grant", "viewer-claim", "job-start-route", "audit-start-route"},
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

	viewerPreview := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:viewer", csvBody)
	if viewerPreview.Code != http.StatusForbidden {
		t.Fatalf("expected viewer preview status %d, got %d with body %s", http.StatusForbidden, viewerPreview.Code, viewerPreview.Body.String())
	}
	assertSafeError(t, viewerPreview, "forbidden", "Forbidden.")

	blockedLiveSource := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", map[string]any{
		"sourceType": "legacy_homebox",
		"baseUrl":    "http://127.0.0.1:7744",
		"username":   "owner@example.com",
		"password":   "secret",
	})
	if blockedLiveSource.Code != http.StatusBadRequest {
		t.Fatalf("expected blocked live source status %d, got %d with body %s", http.StatusBadRequest, blockedLiveSource.Code, blockedLiveSource.Body.String())
	}
	assertImportSourceError(t, blockedLiveSource, "Homebox URL resolves to a blocked address")

	credentialURL := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", map[string]any{
		"sourceType": "legacy_homebox",
		"baseUrl":    "https://user:secret@homebox.example.test?token=secret",
		"username":   "owner@example.com",
		"password":   "secret",
	})
	if credentialURL.Code != http.StatusBadRequest {
		t.Fatalf("expected credential URL status %d, got %d with body %s", http.StatusBadRequest, credentialURL.Code, credentialURL.Body.String())
	}
	if body := credentialURL.Body.String(); containsAny(body, "user:secret", "token=secret") {
		t.Fatalf("credential URL error leaked source URL: %s", body)
	}
	assertImportSourceError(t, credentialURL, "Homebox URL must not include credentials")

	invalidCSV := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", map[string]any{
		"sourceType":    "legacy_homebox_csv",
		"fileName":      "homebox.csv",
		"contentBase64": "not base64",
	})
	if invalidCSV.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid CSV status %d, got %d with body %s", http.StatusBadRequest, invalidCSV.Code, invalidCSV.Body.String())
	}
	assertImportSourceError(t, invalidCSV, "CSV import file could not be decoded. Choose a valid exported CSV file and try again.")

	oversizedCSVContent := strings.Repeat("a", appcore.MaxImportCSVBytes+1)
	oversizedCSV := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", map[string]any{
		"sourceType":    "legacy_homebox_csv",
		"fileName":      "homebox.csv",
		"contentBase64": base64.StdEncoding.EncodeToString([]byte(oversizedCSVContent)),
	})
	if oversizedCSV.Code != http.StatusBadRequest {
		t.Fatalf("expected oversized CSV status %d, got %d with body %s", http.StatusBadRequest, oversizedCSV.Code, oversizedCSV.Body.String())
	}
	assertImportSourceError(t, oversizedCSV, "CSV import file is too large. Choose a CSV up to 10 MB.")

	startPreview := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", csvBody)
	if startPreview.Code != http.StatusOK {
		t.Fatalf("expected start-route preview status %d, got %d with body %s", http.StatusOK, startPreview.Code, startPreview.Body.String())
	}
	var previewed importJobResponseEnvelope
	decodeBody(t, startPreview, &previewed)
	invalidStartCSV := performRequest(server, http.MethodPost, path+"/imports/jobs/"+previewed.Data.ID+"/start", "Bearer dev:owner", map[string]any{
		"sourceType":    "legacy_homebox_csv",
		"fileName":      "homebox.csv",
		"contentBase64": "not base64",
	})
	if invalidStartCSV.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid start CSV status %d, got %d with body %s", http.StatusBadRequest, invalidStartCSV.Code, invalidStartCSV.Body.String())
	}
	assertImportSourceError(t, invalidStartCSV, "CSV import file could not be decoded. Choose a valid exported CSV file and try again.")

	oversizedStartCSV := performRequest(server, http.MethodPost, path+"/imports/jobs/"+previewed.Data.ID+"/start", "Bearer dev:owner", map[string]any{
		"sourceType":    "legacy_homebox_csv",
		"fileName":      "homebox.csv",
		"contentBase64": base64.StdEncoding.EncodeToString([]byte(oversizedCSVContent)),
	})
	if oversizedStartCSV.Code != http.StatusBadRequest {
		t.Fatalf("expected oversized start CSV status %d, got %d with body %s", http.StatusBadRequest, oversizedStartCSV.Code, oversizedStartCSV.Body.String())
	}
	assertImportSourceError(t, oversizedStartCSV, "CSV import file is too large. Choose a CSV up to 10 MB.")
}

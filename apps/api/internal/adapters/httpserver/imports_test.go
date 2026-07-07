package httpserver

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
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
}

func TestDurableImportJobEndpointsRejectAdversarialCallers(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
			{id: otherTenantID, name: "Other", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
			{id: otherInventoryID, tenantID: otherTenantID, name: "Other Tools", owner: "other-owner"},
		},
		ids: []string{"viewer-grant-event", "audit-viewer-grant", "viewer-claim", "job-secure", "job-delete"},
	}))
	path := "/tenants/" + tenantID + "/inventories/" + inventoryID
	otherPath := "/tenants/" + otherTenantID + "/inventories/" + otherInventoryID
	source := map[string]any{
		"sourceType":    "legacy_homebox_csv",
		"fileName":      "homebox.csv",
		"contentBase64": base64.StdEncoding.EncodeToString([]byte("HB.location,HB.asset_id,HB.name\nGarage,HB-1,Drill\n")),
	}
	grantViewer := performRequest(server, http.MethodPost, path+"/access-grants", "Bearer dev:owner", map[string]any{
		"principalId":  "viewer",
		"relationship": "viewer",
	})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, grantViewer.Code, grantViewer.Body.String())
	}
	create := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", source)
	if create.Code != http.StatusOK {
		t.Fatalf("expected preview status %d, got %d with body %s", http.StatusOK, create.Code, create.Body.String())
	}
	var created importJobResponseEnvelope
	decodeBody(t, create, &created)
	deletablePreview := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", source)
	if deletablePreview.Code != http.StatusOK {
		t.Fatalf("expected deletable preview status %d, got %d with body %s", http.StatusOK, deletablePreview.Code, deletablePreview.Body.String())
	}
	var deletable importJobResponseEnvelope
	decodeBody(t, deletablePreview, &deletable)
	cancelDeletable := performRequest(server, http.MethodPost, path+"/imports/jobs/"+deletable.Data.ID+"/cancel", "Bearer dev:owner", map[string]any{
		"mode": "keep_partial_progress",
	})
	if cancelDeletable.Code != http.StatusOK {
		t.Fatalf("expected cancel status %d, got %d with body %s", http.StatusOK, cancelDeletable.Code, cancelDeletable.Body.String())
	}

	routes := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "list", method: http.MethodGet, path: path + "/imports/jobs"},
		{name: "preview", method: http.MethodPost, path: path + "/imports/jobs/preview", body: source},
		{name: "detail", method: http.MethodGet, path: path + "/imports/jobs/" + created.Data.ID},
		{name: "start", method: http.MethodPost, path: path + "/imports/jobs/" + created.Data.ID + "/start", body: source},
		{name: "cancel", method: http.MethodPost, path: path + "/imports/jobs/" + created.Data.ID + "/cancel", body: map[string]any{"mode": "discard_partial_progress"}},
		{name: "delete", method: http.MethodDelete, path: path + "/imports/jobs/" + deletable.Data.ID},
	}
	for _, route := range routes {
		t.Run(route.name+"/missing_auth", func(t *testing.T) {
			response := performRequest(server, route.method, route.path, "", route.body)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected missing auth status %d, got %d with body %s", http.StatusUnauthorized, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "authentication_required", "Authentication required.")
		})
		t.Run(route.name+"/malformed_auth", func(t *testing.T) {
			response := performRequest(server, route.method, route.path, "Bearer nope", route.body)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected malformed auth status %d, got %d with body %s", http.StatusUnauthorized, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "authentication_required", "Authentication required.")
		})
		for _, caller := range []string{"viewer", "intruder"} {
			t.Run(route.name+"/"+caller, func(t *testing.T) {
				response := performRequest(server, route.method, route.path, "Bearer dev:"+caller, route.body)
				if response.Code != http.StatusForbidden {
					t.Fatalf("expected %s status %d, got %d with body %s", caller, http.StatusForbidden, response.Code, response.Body.String())
				}
				assertSafeError(t, response, "forbidden", "Forbidden.")
			})
		}
		t.Run(route.name+"/wrong_tenant", func(t *testing.T) {
			wrongPath := strings.Replace(route.path, "/tenants/"+tenantID+"/", "/tenants/"+otherTenantID+"/", 1)
			response := performRequest(server, route.method, wrongPath, "Bearer dev:owner", route.body)
			assertNotSuccessful(t, response, route.method, wrongPath)
		})
		t.Run(route.name+"/wrong_inventory", func(t *testing.T) {
			wrongPath := strings.Replace(route.path, "/inventories/"+inventoryID, "/inventories/"+otherInventoryID, 1)
			response := performRequest(server, route.method, wrongPath, "Bearer dev:owner", route.body)
			assertNotSuccessful(t, response, route.method, wrongPath)
		})
	}

	otherList := performRequest(server, http.MethodGet, otherPath+"/imports/jobs", "Bearer dev:other-owner", nil)
	if otherList.Code != http.StatusOK {
		t.Fatalf("expected other scope list status %d, got %d with body %s", http.StatusOK, otherList.Code, otherList.Body.String())
	}
	var listedOther importJobListResponseEnvelope
	decodeBody(t, otherList, &listedOther)
	for _, job := range listedOther.Data.Jobs {
		if job.ID == created.Data.ID || job.ID == deletable.Data.ID {
			t.Fatalf("valid other scope list exposed original import job: %+v", listedOther.Data.Jobs)
		}
	}
	for _, route := range []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "detail", method: http.MethodGet, path: otherPath + "/imports/jobs/" + created.Data.ID},
		{name: "start", method: http.MethodPost, path: otherPath + "/imports/jobs/" + created.Data.ID + "/start", body: source},
		{name: "cancel", method: http.MethodPost, path: otherPath + "/imports/jobs/" + created.Data.ID + "/cancel", body: map[string]any{"mode": "discard_partial_progress"}},
		{name: "delete", method: http.MethodDelete, path: otherPath + "/imports/jobs/" + deletable.Data.ID},
	} {
		t.Run(route.name+"/valid_other_scope", func(t *testing.T) {
			response := performRequest(server, route.method, route.path, "Bearer dev:other-owner", route.body)
			assertNotSuccessful(t, response, route.method, route.path)
		})
	}
	originalDetail := performRequest(server, http.MethodGet, path+"/imports/jobs/"+created.Data.ID, "Bearer dev:owner", nil)
	if originalDetail.Code != http.StatusOK {
		t.Fatalf("expected original job detail status %d after adversarial access, got %d with body %s", http.StatusOK, originalDetail.Code, originalDetail.Body.String())
	}
	var original importJobResponseEnvelope
	decodeBody(t, originalDetail, &original)
	if original.Data.Status != "previewed" {
		t.Fatalf("expected adversarial cross-scope start/cancel attempts not to mutate original job, got %+v", original.Data)
	}
}

func TestDurableImportJobLifecycleHTTP(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"job-one", "field-one", "asset-one", "job-two"},
	}))
	path := "/tenants/" + tenantID + "/inventories/" + inventoryID
	source := map[string]any{
		"sourceType":    "legacy_homebox_csv",
		"fileName":      "homebox.csv",
		"contentBase64": base64.StdEncoding.EncodeToString([]byte("HB.location,HB.asset_id,HB.name\nGarage,HB-1,Drill\n")),
		"password":      "must-not-echo",
	}

	create := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", source)
	if create.Code != http.StatusOK {
		t.Fatalf("expected create preview status %d, got %d with body %s", http.StatusOK, create.Code, create.Body.String())
	}
	if body := create.Body.String(); containsAny(body, "must-not-echo", "contentBase64") {
		t.Fatalf("durable job preview leaked source input: %s", body)
	}
	var created importJobResponseEnvelope
	decodeBody(t, create, &created)
	if created.Data.ID != "job-one" || created.Data.Status != "previewed" {
		t.Fatalf("unexpected created import job: %+v", created.Data)
	}
	if created.Data.ActorID != "owner" {
		t.Fatalf("expected import job actor owner, got %q", created.Data.ActorID)
	}
	if created.Data.Counts.Assets != 1 || created.Data.Source.Fingerprint == "" {
		t.Fatalf("unexpected preview response: %+v", created.Data)
	}
	if len(created.Data.Preview.Locations) != 1 || created.Data.Preview.Locations[0].Title != "Garage" {
		t.Fatalf("expected bounded preview location samples, got %+v", created.Data.Preview.Locations)
	}
	if len(created.Data.Preview.Assets) != 1 || created.Data.Preview.Assets[0].Title != "Drill" {
		t.Fatalf("expected bounded preview asset samples, got %+v", created.Data.Preview.Assets)
	}
	if len(created.Data.Preview.Fields) == 0 || created.Data.Preview.Fields[0].Key == "" {
		t.Fatalf("expected preview field samples, got %+v", created.Data.Preview.Fields)
	}

	list := performRequest(server, http.MethodGet, path+"/imports/jobs", "Bearer dev:owner", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, list.Code, list.Body.String())
	}
	var listed importJobListResponseEnvelope
	decodeBody(t, list, &listed)
	if len(listed.Data.Jobs) != 1 || listed.Data.Jobs[0].ID != "job-one" {
		t.Fatalf("expected job in import history, got %+v", listed.Data.Jobs)
	}

	start := performRequest(server, http.MethodPost, path+"/imports/jobs/job-one/start", "Bearer dev:owner", source)
	if start.Code != http.StatusOK {
		t.Fatalf("expected start status %d, got %d with body %s", http.StatusOK, start.Code, start.Body.String())
	}
	var started importJobResponseEnvelope
	decodeBody(t, start, &started)
	if started.Data.Status != "running" || started.Data.Progress.Phase != "reading_source" {
		t.Fatalf("unexpected started job: %+v", started.Data)
	}
	if len(started.Data.ProgressHistory) != 2 || started.Data.ProgressHistory[0].Phase != "ready" || started.Data.ProgressHistory[1].Phase != "reading_source" {
		t.Fatalf("expected started job progress history, got %+v", started.Data.ProgressHistory)
	}
	completed := waitForImportJobStatus(t, server, path+"/imports/jobs/job-one", "succeeded")
	if completed.Data.Counts.AssetsCreated != 1 || completed.Data.Counts.FieldsCreated == 0 {
		t.Fatalf("expected import to create planned records, got %+v", completed.Data.Counts)
	}
	if len(completed.Data.Resources) != 2 || completed.Data.Resources[0].ResourceType != "asset" || completed.Data.Resources[0].ResourceID == "" {
		t.Fatalf("expected completed import to expose safe imported resources, got %+v", completed.Data.Resources)
	}
	var drillResourceFound bool
	for _, resource := range completed.Data.Resources {
		if resource.DisplayName == "Drill" {
			drillResourceFound = true
			break
		}
	}
	if !drillResourceFound {
		t.Fatalf("expected imported resources to include safe display names, got %+v", completed.Data.Resources)
	}
	if len(completed.Data.ProgressHistory) < 3 || completed.Data.ProgressHistory[len(completed.Data.ProgressHistory)-1].Phase != "terminal" {
		t.Fatalf("expected completed import progress history through terminal, got %+v", completed.Data.ProgressHistory)
	}
	if body := encodeJSON(t, completed); containsAny(body, "must-not-echo", "contentBase64") {
		t.Fatalf("import resource summary leaked source input: %s", body)
	}

	cancelPreview := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", source)
	if cancelPreview.Code != http.StatusOK {
		t.Fatalf("expected cancel preview status %d, got %d with body %s", http.StatusOK, cancelPreview.Code, cancelPreview.Body.String())
	}
	var previewedForCancel importJobResponseEnvelope
	decodeBody(t, cancelPreview, &previewedForCancel)

	cancel := performRequest(server, http.MethodPost, path+"/imports/jobs/"+previewedForCancel.Data.ID+"/cancel", "Bearer dev:owner", map[string]any{
		"mode": "discard_partial_progress",
	})
	if cancel.Code != http.StatusOK {
		t.Fatalf("expected cancel status %d, got %d with body %s", http.StatusOK, cancel.Code, cancel.Body.String())
	}
	var cancelled importJobResponseEnvelope
	decodeBody(t, cancel, &cancelled)
	if cancelled.Data.Status != "cancelled_discarded" || cancelled.Data.CancellationMode != "discard_partial_progress" {
		t.Fatalf("unexpected cancelled job: %+v", cancelled.Data)
	}
	remove := performRequest(server, http.MethodDelete, path+"/imports/jobs/"+previewedForCancel.Data.ID, "Bearer dev:owner", nil)
	if remove.Code != http.StatusNoContent {
		t.Fatalf("expected remove-from-history status %d, got %d with body %s", http.StatusNoContent, remove.Code, remove.Body.String())
	}
	if remove.Body.Len() != 0 {
		t.Fatalf("expected remove-from-history to return no body, got %s", remove.Body.String())
	}
	listAfterRemove := performRequest(server, http.MethodGet, path+"/imports/jobs", "Bearer dev:owner", nil)
	if listAfterRemove.Code != http.StatusOK {
		t.Fatalf("expected list after remove status %d, got %d with body %s", http.StatusOK, listAfterRemove.Code, listAfterRemove.Body.String())
	}
	var listedAfterRemove importJobListResponseEnvelope
	decodeBody(t, listAfterRemove, &listedAfterRemove)
	for _, job := range listedAfterRemove.Data.Jobs {
		if job.ID == previewedForCancel.Data.ID {
			t.Fatalf("removed import job remained visible in history: %+v", listedAfterRemove.Data.Jobs)
		}
	}
}

func TestDurableImportJobStartReportsSourceChangedPrecondition(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestAppWithImportSource(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"job-stale"},
	}, &stalePreviewImportSource{}))
	path := "/tenants/" + tenantID + "/inventories/" + inventoryID
	source := map[string]any{
		"sourceType":    "legacy_homebox",
		"baseUrl":       "https://homebox.example.test",
		"username":      "owner@example.test",
		"password":      "secret",
		"includeImages": false,
	}

	create := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", source)
	if create.Code != http.StatusOK {
		t.Fatalf("expected preview status %d, got %d with body %s", http.StatusOK, create.Code, create.Body.String())
	}

	start := performRequest(server, http.MethodPost, path+"/imports/jobs/job-stale/start", "Bearer dev:owner", source)
	if start.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected stale source start status %d, got %d with body %s", http.StatusPreconditionFailed, start.Code, start.Body.String())
	}
	assertSafeError(t, start, "precondition_failed", "Import source changed after preview. Preview the source again before starting the import.")
	if body := start.Body.String(); containsAny(body, "secret", "owner@example.test") {
		t.Fatalf("source changed precondition leaked source input: %s", body)
	}
}

func TestDurableImportJobStartRequiresNewPreviewWhenSecurityOptionsChange(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestAppWithImportSource(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{"job-options"},
	}, &attachmentImportSource{}))
	path := "/tenants/" + tenantID + "/inventories/" + inventoryID
	source := map[string]any{
		"sourceType":    "legacy_homebox",
		"baseUrl":       "https://homebox.example.test",
		"username":      "owner@example.test",
		"password":      "secret",
		"includeImages": true,
	}

	create := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", source)
	if create.Code != http.StatusOK {
		t.Fatalf("expected preview status %d, got %d with body %s", http.StatusOK, create.Code, create.Body.String())
	}

	changed := map[string]any{}
	for key, value := range source {
		changed[key] = value
	}
	changed["allowPrivateNetwork"] = true
	start := performRequest(server, http.MethodPost, path+"/imports/jobs/job-options/start", "Bearer dev:owner", changed)
	if start.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected security option start status %d, got %d with body %s", http.StatusPreconditionFailed, start.Code, start.Body.String())
	}
	assertSafeError(t, start, "precondition_failed", "Import source changed after preview. Preview the source again before starting the import.")

	detail := performRequest(server, http.MethodGet, path+"/imports/jobs/job-options", "Bearer dev:owner", nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("expected detail status %d, got %d with body %s", http.StatusOK, detail.Code, detail.Body.String())
	}
	var body importJobResponseEnvelope
	decodeBody(t, detail, &body)
	if body.Data.Status != "previewed" || body.Data.Source.AllowPrivateNetwork || body.Data.StartedAt != "" {
		t.Fatalf("expected security option mismatch to leave job previewed with original source options, got %+v", body.Data)
	}
}

func TestDurableImportJobImportsAttachmentsThroughHTTPBoundary(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	sourceReader := &attachmentImportSource{}
	server := NewServer(":0", newSeededTestAppWithImportSource(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{
			"job-with-attachment",
			"field-source", "field-asset",
			"location-garage", "op-location-garage", "audit-location-garage",
			"asset-drill", "op-asset-drill", "audit-asset-drill",
			"attachment-drill", "audit-attachment-drill",
		},
	}, sourceReader))
	path := "/tenants/" + tenantID + "/inventories/" + inventoryID
	source := map[string]any{
		"sourceType":    "legacy_homebox",
		"baseUrl":       "https://homebox.example.test",
		"username":      "owner@example.test",
		"password":      "secret",
		"includeImages": true,
	}

	create := performRequest(server, http.MethodPost, path+"/imports/jobs/preview", "Bearer dev:owner", source)
	if create.Code != http.StatusOK {
		t.Fatalf("expected preview status %d, got %d with body %s", http.StatusOK, create.Code, create.Body.String())
	}
	var created importJobResponseEnvelope
	decodeBody(t, create, &created)
	if created.Data.Counts.Attachments != 1 || len(created.Data.Preview.Attachments) != 1 {
		t.Fatalf("expected attachment in preview, got %+v", created.Data)
	}

	start := performRequest(server, http.MethodPost, path+"/imports/jobs/"+created.Data.ID+"/start", "Bearer dev:owner", source)
	if start.Code != http.StatusOK {
		t.Fatalf("expected start status %d, got %d with body %s", http.StatusOK, start.Code, start.Body.String())
	}

	completed := waitForImportJobStatus(t, server, path+"/imports/jobs/"+created.Data.ID, "succeeded")
	if completed.Data.Counts.AttachmentsCreated != 1 || completed.Data.Counts.AssetsCreated != 1 {
		t.Fatalf("expected import to create item and attachment, got %+v", completed.Data.Counts)
	}
	var attachmentResource *struct {
		ResourceType     string `json:"resourceType"`
		ResourceID       string `json:"resourceId"`
		DisplayName      string `json:"displayName"`
		ResourceOwnerID  string `json:"resourceOwnerId"`
		SourceEntityType string `json:"sourceEntityType"`
		SourceEntityID   string `json:"sourceEntityId"`
		CreatedAt        string `json:"createdAt"`
	}
	for index := range completed.Data.Resources {
		if completed.Data.Resources[index].ResourceType == "attachment" {
			attachmentResource = &completed.Data.Resources[index]
			break
		}
	}
	if attachmentResource == nil || attachmentResource.ResourceID == "" || attachmentResource.ResourceOwnerID == "" || attachmentResource.SourceEntityID != "attachment:drill-photo" {
		t.Fatalf("expected safe imported attachment resource, got %+v", completed.Data.Resources)
	}
	download := performRequest(server, http.MethodGet, path+"/assets/"+attachmentResource.ResourceOwnerID+"/attachments/"+attachmentResource.ResourceID+"/content", "Bearer dev:owner", nil)
	if download.Code != http.StatusOK {
		t.Fatalf("expected imported attachment download status %d, got %d with body %s", http.StatusOK, download.Code, download.Body.String())
	}
	if download.Header().Get("Content-Type") != "image/png" || !bytes.Equal(download.Body.Bytes(), pngAttachmentContent()) {
		t.Fatalf("expected imported PNG bytes, content type %q length %d", download.Header().Get("Content-Type"), download.Body.Len())
	}
	detail := performRequest(server, http.MethodGet, path+"/assets/"+attachmentResource.ResourceOwnerID, "Bearer dev:owner", nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("expected imported asset detail status %d, got %d with body %s", http.StatusOK, detail.Code, detail.Body.String())
	}
	detailBody := decodeAsset(t, detail)
	if detailBody.Data.PrimaryPhoto == nil || detailBody.Data.PrimaryPhoto.ID != attachmentResource.ResourceID || detailBody.Data.PrimaryPhoto.ContentType != "image/png" {
		t.Fatalf("expected imported attachment to be asset detail primary photo, got %+v", detailBody.Data.PrimaryPhoto)
	}
	list := performRequest(server, http.MethodGet, path+"/assets", "Bearer dev:owner", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected imported asset list status %d, got %d with body %s", http.StatusOK, list.Code, list.Body.String())
	}
	var listedPrimaryPhoto *assetPrimaryPhoto
	for _, item := range decodeAssetList(t, list).Data {
		if item.ID == attachmentResource.ResourceOwnerID {
			listedPrimaryPhoto = item.PrimaryPhoto
			break
		}
	}
	if listedPrimaryPhoto == nil || listedPrimaryPhoto.ID != attachmentResource.ResourceID {
		t.Fatalf("expected imported attachment to be asset list primary photo, got %+v", listedPrimaryPhoto)
	}
	search := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/search/assets?q=drill&inventoryId="+inventoryID, "Bearer dev:owner", nil)
	if search.Code != http.StatusOK {
		t.Fatalf("expected imported asset search status %d, got %d with body %s", http.StatusOK, search.Code, search.Body.String())
	}
	var searchPrimaryPhoto *assetPrimaryPhoto
	for _, result := range decodeAssetSearch(t, search).Data {
		if result.Asset.ID == attachmentResource.ResourceOwnerID {
			searchPrimaryPhoto = result.Asset.PrimaryPhoto
			break
		}
	}
	if searchPrimaryPhoto == nil || searchPrimaryPhoto.ID != attachmentResource.ResourceID {
		t.Fatalf("expected imported attachment to be search primary photo, got %+v", searchPrimaryPhoto)
	}
	if got := sourceReader.fetchAttachmentBytesCalls(); len(got) != 3 || got[0] || got[1] || !got[2] {
		t.Fatalf("expected preview/start preflight without bytes and worker apply with bytes, got %+v", got)
	}
}

func waitForImportJobStatus(t *testing.T, server *http.Server, path string, status string) importJobResponseEnvelope {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		response := performRequest(server, http.MethodGet, path, "Bearer dev:owner", nil)
		if response.Code != http.StatusOK {
			t.Fatalf("expected job status response %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
		}
		var envelope importJobResponseEnvelope
		decodeBody(t, response, &envelope)
		if envelope.Data.Status == status {
			return envelope
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for import job %s, last job %+v", status, envelope.Data)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func assertNotSuccessful(t *testing.T, response *httptest.ResponseRecorder, method string, path string) {
	t.Helper()
	if response.Code >= http.StatusOK && response.Code < http.StatusMultipleChoices {
		t.Fatalf("expected adversarial %s %s to fail, got %d with body %s", method, path, response.Code, response.Body.String())
	}
	if response.Code != http.StatusForbidden && response.Code != http.StatusNotFound && response.Code != http.StatusBadRequest {
		t.Fatalf("expected adversarial %s %s to fail safely, got %d with body %s", method, path, response.Code, response.Body.String())
	}
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

func encodeJSON(t *testing.T, value any) string {
	t.Helper()
	out, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal test value: %v", err)
	}
	return string(out)
}

type importJobResponseEnvelope struct {
	Data importJobResponse `json:"data"`
}

type importJobListResponseEnvelope struct {
	Data struct {
		Jobs []importJobResponse `json:"jobs"`
	} `json:"data"`
}

type importJobResponse struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	ActorID          string `json:"actorId"`
	CancellationMode string `json:"cancellationMode"`
	StartedAt        string `json:"startedAt"`
	Source           struct {
		Fingerprint         string `json:"fingerprint"`
		AllowPrivateNetwork bool   `json:"allowPrivateNetwork"`
		AllowInsecureTLS    bool   `json:"allowInsecureTLS"`
	} `json:"source"`
	Counts struct {
		Assets             int `json:"assets"`
		Attachments        int `json:"attachments"`
		AssetsCreated      int `json:"assetsCreated"`
		AttachmentsCreated int `json:"attachmentsCreated"`
		FieldsCreated      int `json:"fieldsCreated"`
	} `json:"counts"`
	Preview struct {
		Fields []struct {
			Key string `json:"key"`
		} `json:"fields"`
		Locations []struct {
			Kind  string `json:"kind"`
			Title string `json:"title"`
		} `json:"locations"`
		Assets []struct {
			Kind  string `json:"kind"`
			Title string `json:"title"`
		} `json:"assets"`
		Attachments []struct {
			FileName string `json:"fileName"`
		} `json:"attachments"`
		Messages []struct {
			Code string `json:"code"`
		} `json:"messages"`
	} `json:"preview"`
	Progress struct {
		Phase string `json:"phase"`
	} `json:"progress"`
	ProgressHistory []struct {
		Phase string `json:"phase"`
	} `json:"progressHistory"`
	Resources []struct {
		ResourceType     string `json:"resourceType"`
		ResourceID       string `json:"resourceId"`
		DisplayName      string `json:"displayName"`
		ResourceOwnerID  string `json:"resourceOwnerId"`
		SourceEntityType string `json:"sourceEntityType"`
		SourceEntityID   string `json:"sourceEntityId"`
		CreatedAt        string `json:"createdAt"`
	} `json:"resources"`
}

type attachmentImportSource struct {
	mu                     sync.Mutex
	fetchAttachmentBytes   []bool
	unexpectedRequestError error
}

type stalePreviewImportSource struct {
	mu    sync.Mutex
	calls int
}

func (s *stalePreviewImportSource) ReadImportPlan(_ context.Context, request ports.ImportSourceRequest) (importplan.Plan, error) {
	if request.SourceType != importplan.SourceLegacyHomebox ||
		request.BaseURL != "https://homebox.example.test" ||
		request.Username != "owner@example.test" ||
		request.Password != "secret" ||
		request.IncludeImages ||
		request.AllowPrivateNetwork ||
		request.AllowInsecureTLS ||
		request.FileName != "" ||
		len(request.Content) != 0 ||
		request.FetchAttachmentBytes {
		return importplan.Plan{}, errors.New("unexpected stale preview import source request")
	}
	s.mu.Lock()
	s.calls++
	call := s.calls
	s.mu.Unlock()
	sourceID := "asset:drill"
	title := "Drill"
	if call > 1 {
		sourceID = "asset:changed"
		title = "Changed drill"
	}
	return importplan.Plan{
		Source: importplan.SourceSummary{
			Type:        importplan.SourceLegacyHomebox,
			Name:        "Homebox",
			BaseURL:     request.BaseURL,
			ImageImport: "disabled",
		},
		Fields: []importplan.FieldDefinition{
			{Key: "homebox-source-id", DisplayName: "Homebox source ID", Type: "text"},
		},
		Assets: []importplan.Asset{
			{SourceID: sourceID, Kind: "item", Title: title, CustomFields: map[string]any{"homebox-source-id": sourceID}},
		},
	}, nil
}

func (s *attachmentImportSource) ReadImportPlan(_ context.Context, request ports.ImportSourceRequest) (importplan.Plan, error) {
	s.mu.Lock()
	s.fetchAttachmentBytes = append(s.fetchAttachmentBytes, request.FetchAttachmentBytes)
	s.mu.Unlock()
	if request.SourceType != importplan.SourceLegacyHomebox ||
		request.BaseURL != "https://homebox.example.test" ||
		request.Username != "owner@example.test" ||
		request.Password != "secret" ||
		!request.IncludeImages ||
		request.AllowPrivateNetwork ||
		request.AllowInsecureTLS ||
		request.FileName != "" ||
		len(request.Content) != 0 {
		s.unexpectedRequestError = errors.New("unexpected import source request")
		return importplan.Plan{}, s.unexpectedRequestError
	}
	attachment := importplan.Attachment{
		SourceID:      "attachment:drill-photo",
		AssetSourceID: "asset:drill",
		FileName:      "drill.png",
		ContentType:   "image/png",
		Primary:       true,
	}
	if request.FetchAttachmentBytes {
		attachment.Content = pngAttachmentContent()
		attachment.SizeBytes = len(attachment.Content)
	}
	return importplan.Plan{
		Source: importplan.SourceSummary{
			Type:        importplan.SourceLegacyHomebox,
			Name:        "Homebox",
			BaseURL:     request.BaseURL,
			ImageImport: "enabled",
		},
		Fields: []importplan.FieldDefinition{
			{Key: "homebox-source-id", DisplayName: "Homebox source ID", Type: "text"},
			{Key: "homebox-asset-id", DisplayName: "Homebox asset ID", Type: "text"},
		},
		Assets: []importplan.Asset{
			{SourceID: "location:garage", Kind: "location", Title: "Garage", CustomFields: map[string]any{"homebox-source-id": "location:garage"}},
			{SourceID: "asset:drill", Kind: "item", Title: "Drill", ParentSourceID: "location:garage", CustomFields: map[string]any{"homebox-source-id": "asset:drill", "homebox-asset-id": "HB-1"}},
		},
		Attachments: []importplan.Attachment{attachment},
	}, nil
}

func (s *attachmentImportSource) fetchAttachmentBytesCalls() []bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]bool{}, s.fetchAttachmentBytes...)
}

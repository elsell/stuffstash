package httpserver

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

type tenantResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LifecycleState string `json:"lifecycleState"`
}

type tenantBody struct {
	Data tenantResponse `json:"data"`
	Meta responseMeta   `json:"meta"`
}

func TestResourceLifecycleEndpointsCoverCurrentSurface(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{},
		"tenant-lifecycle", "audit-tenant-create", "outbox-tenant",
		"audit-tenant-view", "audit-tenant-update", "audit-tenant-archive", "audit-tenant-restore",
		"inventory-lifecycle", "audit-inventory-create", "outbox-inventory",
		"audit-inventory-view", "audit-inventory-update", "audit-inventory-archive", "audit-inventory-restore",
		"asset-lifecycle", "audit-asset-create", "audit-asset-view",
		"attachment-lifecycle", "audit-attachment-create", "audit-attachment-view", "audit-attachment-archive", "audit-attachment-restore",
		"type-lifecycle", "audit-type-create", "audit-type-view", "audit-type-archive", "audit-type-restore",
		"field-lifecycle", "audit-field-create", "audit-field-view", "audit-field-archive", "audit-field-restore",
		"audit-grant", "outbox-grant", "audit-grant-view",
		"invite-lifecycle", "audit-invite-create", "audit-invite-view", "audit-invite-cancel", "audit-invite-delete",
		"audit-attachment-delete", "audit-type-delete", "audit-field-delete", "audit-asset-delete", "audit-inventory-delete", "audit-tenant-delete",
	))

	tenantCreate := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:owner", map[string]any{"name": "Home"})
	requireStatus(t, tenantCreate, http.StatusCreated)
	var createdTenant tenantBody
	decodeBody(t, tenantCreate, &createdTenant)
	tenantID := createdTenant.Data.ID

	tenantGet := performRequest(server, http.MethodGet, "/tenants/"+tenantID, "Bearer dev:owner", nil)
	requireStatus(t, tenantGet, http.StatusOK)
	var tenantView tenantBody
	decodeBody(t, tenantGet, &tenantView)
	if tenantView.Data.LifecycleState != "active" {
		t.Fatalf("expected active tenant, got %+v", tenantView.Data)
	}
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID, "Bearer dev:owner", map[string]any{"name": "Updated Home"}), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/archive", "Bearer dev:owner", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/restore", "Bearer dev:owner", nil), http.StatusOK)

	inventoryCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:owner", map[string]any{"name": "Tools"})
	requireStatus(t, inventoryCreate, http.StatusCreated)
	var createdInventory inventoryBody
	decodeBody(t, inventoryCreate, &createdInventory)
	inventoryID := createdInventory.Data.ID
	inventoryGet := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID, "Bearer dev:owner", nil)
	requireStatus(t, inventoryGet, http.StatusOK)
	var inventoryView inventoryBody
	decodeBody(t, inventoryGet, &inventoryView)
	if inventoryView.Data.LifecycleState != "active" {
		t.Fatalf("expected active inventory, got %+v", inventoryView.Data)
	}
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID, "Bearer dev:owner", map[string]any{"name": "Updated Tools"}), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/archive", "Bearer dev:owner", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/restore", "Bearer dev:owner", nil), http.StatusOK)

	assetCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{"kind": "item", "title": "Drill"})
	requireStatus(t, assetCreate, http.StatusCreated)
	assetID := decodeAsset(t, assetCreate).Data.ID
	requireStatus(t, performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+assetID, "Bearer dev:owner", nil), http.StatusOK)

	attachmentCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+assetID+"/attachments", "Bearer dev:owner", map[string]any{
		"fileName":      "receipt.png",
		"contentType":   "image/png",
		"contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent()),
	})
	requireStatus(t, attachmentCreate, http.StatusCreated)
	attachmentID := decodeAttachment(t, attachmentCreate).Data.ID
	requireStatus(t, performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+assetID+"/attachments/"+attachmentID, "Bearer dev:owner", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+assetID+"/attachments/"+attachmentID+"/archive", "Bearer dev:owner", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+assetID+"/attachments/"+attachmentID+"/restore", "Bearer dev:owner", nil), http.StatusOK)

	typeCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:owner", map[string]any{"key": "medicine", "displayName": "Medicine"})
	requireStatus(t, typeCreate, http.StatusCreated)
	typeID := decodeCustomAssetType(t, typeCreate).Data.ID
	requireStatus(t, performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+typeID, "Bearer dev:owner", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+typeID+"/archive", "Bearer dev:owner", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+typeID+"/restore", "Bearer dev:owner", nil), http.StatusOK)

	fieldCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions", "Bearer dev:owner", map[string]any{"key": "expires", "displayName": "Expires", "type": "date"})
	requireStatus(t, fieldCreate, http.StatusCreated)
	fieldID := decodeCustomFieldDefinition(t, fieldCreate).Data.ID
	requireStatus(t, performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions/"+fieldID, "Bearer dev:owner", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions/"+fieldID+"/archive", "Bearer dev:owner", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions/"+fieldID+"/restore", "Bearer dev:owner", nil), http.StatusOK)

	grantCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]any{"principalId": "viewer", "relationship": "viewer"})
	requireStatus(t, grantCreate, http.StatusCreated)
	requireStatus(t, performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants/viewer/viewer", "Bearer dev:owner", nil), http.StatusOK)

	inviteCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:owner", map[string]any{"email": "invitee@example.com", "relationship": "viewer"})
	requireStatus(t, inviteCreate, http.StatusCreated)
	invitationID := decodeInventoryAccessInvitation(t, inviteCreate).Data.ID
	requireStatus(t, performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitationID, "Bearer dev:owner", nil), http.StatusOK)
	requireStatus(t, performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitationID+"/cancel", "Bearer dev:owner", nil), http.StatusNoContent)
	requireStatus(t, performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitationID, "Bearer dev:owner", nil), http.StatusNoContent)

	requireStatus(t, performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+assetID+"/attachments/"+attachmentID, "Bearer dev:owner", nil), http.StatusNoContent)
	requireStatus(t, performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+typeID, "Bearer dev:owner", nil), http.StatusNoContent)
	requireStatus(t, performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions/"+fieldID, "Bearer dev:owner", nil), http.StatusNoContent)
	requireStatus(t, performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+assetID, "Bearer dev:owner", nil), http.StatusNoContent)
	requireStatus(t, performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID, "Bearer dev:owner", nil), http.StatusNoContent)
	requireStatus(t, performRequest(server, http.MethodDelete, "/tenants/"+tenantID, "Bearer dev:owner", nil), http.StatusNoContent)
}

func TestResourceLifecycleEndpointsRejectUnauthorizedCallers(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{},
		"tenant-security", "audit-tenant-create", "outbox-tenant",
		"audit-inventory-create", "outbox-inventory",
		"asset-security", "audit-asset-create",
		"attachment-security", "audit-attachment-create",
		"type-security", "audit-type-create",
		"field-security", "audit-field-create",
		"audit-grant", "outbox-grant",
		"invite-security", "audit-invite-create",
	))

	tenantCreate := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:owner", map[string]any{"name": "Home"})
	requireStatus(t, tenantCreate, http.StatusCreated)
	tenantID := decodeTenant(t, tenantCreate).Data.ID

	inventoryCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:owner", map[string]any{"name": "Tools"})
	requireStatus(t, inventoryCreate, http.StatusCreated)
	var createdInventory inventoryBody
	decodeBody(t, inventoryCreate, &createdInventory)
	inventoryID := createdInventory.Data.ID

	assetCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{"kind": "item", "title": "Drill"})
	requireStatus(t, assetCreate, http.StatusCreated)
	assetID := decodeAsset(t, assetCreate).Data.ID

	attachmentCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+assetID+"/attachments", "Bearer dev:owner", map[string]any{
		"fileName":      "receipt.png",
		"contentType":   "image/png",
		"contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent()),
	})
	requireStatus(t, attachmentCreate, http.StatusCreated)
	attachmentID := decodeAttachment(t, attachmentCreate).Data.ID

	typeCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:owner", map[string]any{"key": "medicine", "displayName": "Medicine"})
	requireStatus(t, typeCreate, http.StatusCreated)
	typeID := decodeCustomAssetType(t, typeCreate).Data.ID

	fieldCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions", "Bearer dev:owner", map[string]any{"key": "expires", "displayName": "Expires", "type": "date"})
	requireStatus(t, fieldCreate, http.StatusCreated)
	fieldID := decodeCustomFieldDefinition(t, fieldCreate).Data.ID

	grantCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]any{"principalId": "viewer", "relationship": "viewer"})
	requireStatus(t, grantCreate, http.StatusCreated)

	inviteCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:owner", map[string]any{"email": "invitee@example.com", "relationship": "viewer"})
	requireStatus(t, inviteCreate, http.StatusCreated)
	invitationID := decodeInventoryAccessInvitation(t, inviteCreate).Data.ID

	requests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "tenant detail", method: http.MethodGet, path: "/tenants/" + tenantID},
		{name: "tenant update", method: http.MethodPatch, path: "/tenants/" + tenantID, body: map[string]any{"name": "Renamed"}},
		{name: "tenant archive", method: http.MethodPatch, path: "/tenants/" + tenantID + "/archive"},
		{name: "tenant restore", method: http.MethodPatch, path: "/tenants/" + tenantID + "/restore"},
		{name: "tenant delete", method: http.MethodDelete, path: "/tenants/" + tenantID},
		{name: "inventory detail", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID},
		{name: "inventory update", method: http.MethodPatch, path: "/tenants/" + tenantID + "/inventories/" + inventoryID, body: map[string]any{"name": "Renamed"}},
		{name: "inventory archive", method: http.MethodPatch, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/archive"},
		{name: "inventory restore", method: http.MethodPatch, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/restore"},
		{name: "inventory delete", method: http.MethodDelete, path: "/tenants/" + tenantID + "/inventories/" + inventoryID},
		{name: "asset detail", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + assetID},
		{name: "asset delete", method: http.MethodDelete, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + assetID},
		{name: "attachment detail", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + assetID + "/attachments/" + attachmentID},
		{name: "attachment archive", method: http.MethodPatch, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + assetID + "/attachments/" + attachmentID + "/archive"},
		{name: "attachment restore", method: http.MethodPatch, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + assetID + "/attachments/" + attachmentID + "/restore"},
		{name: "attachment delete", method: http.MethodDelete, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/assets/" + assetID + "/attachments/" + attachmentID},
		{name: "custom asset type detail", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-asset-types/" + typeID},
		{name: "custom asset type archive", method: http.MethodPatch, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-asset-types/" + typeID + "/archive"},
		{name: "custom asset type restore", method: http.MethodPatch, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-asset-types/" + typeID + "/restore"},
		{name: "custom asset type delete", method: http.MethodDelete, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-asset-types/" + typeID},
		{name: "custom field detail", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-field-definitions/" + fieldID},
		{name: "custom field archive", method: http.MethodPatch, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-field-definitions/" + fieldID + "/archive"},
		{name: "custom field restore", method: http.MethodPatch, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-field-definitions/" + fieldID + "/restore"},
		{name: "custom field delete", method: http.MethodDelete, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-field-definitions/" + fieldID},
		{name: "access grant detail", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/access-grants/viewer/viewer"},
		{name: "invitation detail", method: http.MethodGet, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/access-invitations/" + invitationID},
		{name: "invitation cancel", method: http.MethodPatch, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/access-invitations/" + invitationID + "/cancel"},
		{name: "invitation delete", method: http.MethodDelete, path: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/access-invitations/" + invitationID},
	}

	authCases := []struct {
		name          string
		authorization string
		status        int
		code          string
		message       string
	}{
		{name: "missing auth", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "malformed auth", authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "intruder", authorization: "Bearer dev:intruder", status: http.StatusForbidden, code: "forbidden", message: "Forbidden."},
	}

	for _, request := range requests {
		for _, authCase := range authCases {
			t.Run(request.name+" "+authCase.name, func(t *testing.T) {
				response := performRequest(server, request.method, request.path, authCase.authorization, request.body)
				requireStatus(t, response, authCase.status)
				assertSafeError(t, response, authCase.code, authCase.message)
			})
		}
	}
}

func decodeTenant(t *testing.T, response *httptest.ResponseRecorder) tenantBody {
	t.Helper()

	var body tenantBody
	decodeBody(t, response, &body)
	return body
}

func requireStatus(t *testing.T, response *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if response.Code != expected {
		t.Fatalf("expected status %d, got %d with body %s", expected, response.Code, response.Body.String())
	}
}

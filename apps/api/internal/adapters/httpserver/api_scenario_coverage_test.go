package httpserver

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
)

type executedScenarioCoverage struct {
	name      string
	operation map[string]struct{}
}

func TestOpenAPIOperationsHaveRepresentativeScenarioCoverage(t *testing.T) {
	openAPIOperations := generatedOpenAPIOperations(t)

	realUse := realUseScenarioOperations(t)
	assertScenarioCoversOpenAPI(t, realUse, openAPIOperations)

	adversarial := adversarialScenarioOperations(t)
	assertScenarioCoversOpenAPI(t, adversarial, openAPIOperations)
}

func realUseScenarioOperations(t *testing.T) executedScenarioCoverage {
	t.Helper()

	coverage := newExecutedScenarioCoverage("real use")
	directUploads := &httpFakeDirectAttachmentUploader{}
	imageProcessor := &httpFakeImageProcessor{thumbnailContent: []byte("thumbnail")}
	server := NewServer(":0", newSeededMediaTestApp(t, seededState{}, directUploads, imageProcessor))

	coverage.request(t, server, http.MethodGet, "/me", "/me", "Bearer dev:owner", nil, http.StatusOK)

	tenantCreate := coverage.request(t, server, http.MethodPost, "/tenants", "/tenants", "Bearer dev:owner", map[string]any{"name": "Home"}, http.StatusCreated)
	tenantID := decodeTenant(t, tenantCreate).Data.ID
	tenantPath := "/tenants/" + tenantID
	coverage.request(t, server, http.MethodGet, "/me/tenants", "/me/tenants?limit=10", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}", tenantPath, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}", tenantPath, "Bearer dev:owner", map[string]any{"name": "Updated Home"}, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/archive", tenantPath+"/archive", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/restore", tenantPath+"/restore", "Bearer dev:owner", nil, http.StatusOK)

	providerProfile := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/provider-profiles", tenantPath+"/provider-profiles", "Bearer dev:owner", map[string]any{
		"capability":         "language_inference",
		"providerKind":       "gemini",
		"displayName":        "Google Gemini",
		"modelName":          "gemini-2.5-flash-lite",
		"runtimeOptions":     map[string]any{"temperature": 0.1},
		"capabilityMetadata": map[string]any{"toolCalls": true},
	}, http.StatusCreated)
	providerProfileID := decodeProviderProfile(t, providerProfile).Data.ID
	providerProfilePath := tenantPath + "/provider-profiles/" + providerProfileID
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/provider-profiles", tenantPath+"/provider-profiles", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/provider-profiles/{providerProfileId}", providerProfilePath, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPut, "/tenants/{tenantId}/provider-profiles/{providerProfileId}/credential", providerProfilePath+"/credential", "Bearer dev:owner", map[string]any{"purpose": "api_key", "credential": "scenario-secret"}, http.StatusOK)
	coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/provider-profiles/{providerProfileId}/enable", providerProfilePath+"/enable", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/provider-profiles/{providerProfileId}/disable", providerProfilePath+"/disable", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/provider-profiles/{providerProfileId}/archive", providerProfilePath+"/archive", "Bearer dev:owner", nil, http.StatusOK)

	inventoryCreate := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories", tenantPath+"/inventories", "Bearer dev:owner", map[string]any{"name": "Tools"}, http.StatusCreated)
	inventoryID := decodeScenarioInventory(t, inventoryCreate).Data.ID
	inventoryPath := tenantPath + "/inventories/" + inventoryID
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories", tenantPath+"/inventories?limit=10", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}", inventoryPath, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}", inventoryPath, "Bearer dev:owner", map[string]any{"name": "Updated Tools"}, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/archive", inventoryPath+"/archive", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/restore", inventoryPath+"/restore", "Bearer dev:owner", nil, http.StatusOK)

	tenantType := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/custom-asset-types", tenantPath+"/custom-asset-types", "Bearer dev:owner", map[string]any{"key": "medicine", "displayName": "Medicine"}, http.StatusCreated)
	tenantTypeID := decodeCustomAssetType(t, tenantType).Data.ID
	tenantTypePath := tenantPath + "/custom-asset-types/" + tenantTypeID
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/custom-asset-types", tenantPath+"/custom-asset-types?limit=10", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}", tenantTypePath, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}", tenantTypePath, "Bearer dev:owner", map[string]any{"displayName": "Medicine Supplies"}, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}/archive", tenantTypePath+"/archive", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}/restore", tenantTypePath+"/restore", "Bearer dev:owner", nil, http.StatusOK)

	inventoryType := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types", inventoryPath+"/custom-asset-types", "Bearer dev:owner", map[string]any{"key": "tool", "displayName": "Tool"}, http.StatusCreated)
	inventoryTypeID := decodeCustomAssetType(t, inventoryType).Data.ID
	inventoryTypePath := inventoryPath + "/custom-asset-types/" + inventoryTypeID
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types", inventoryPath+"/custom-asset-types?limit=10", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}", inventoryTypePath, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}", inventoryTypePath, "Bearer dev:owner", map[string]any{"displayName": "Shop Tool"}, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/archive", inventoryTypePath+"/archive", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/restore", inventoryTypePath+"/restore", "Bearer dev:owner", nil, http.StatusOK)

	tenantField := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/custom-field-definitions", tenantPath+"/custom-field-definitions", "Bearer dev:owner", map[string]any{"key": "serial", "displayName": "Serial", "type": "text"}, http.StatusCreated)
	tenantFieldID := decodeCustomFieldDefinition(t, tenantField).Data.ID
	tenantFieldPath := tenantPath + "/custom-field-definitions/" + tenantFieldID
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/custom-field-definitions", tenantPath+"/custom-field-definitions?limit=10", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/custom-field-definitions/{definitionId}", tenantFieldPath, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/custom-field-definitions/{definitionId}", tenantFieldPath, "Bearer dev:owner", map[string]any{"displayName": "Serial Number"}, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/custom-field-definitions/{definitionId}/archive", tenantFieldPath+"/archive", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/custom-field-definitions/{definitionId}/restore", tenantFieldPath+"/restore", "Bearer dev:owner", nil, http.StatusOK)

	inventoryField := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions", inventoryPath+"/custom-field-definitions", "Bearer dev:owner", map[string]any{"key": "expires", "displayName": "Expires", "type": "date"}, http.StatusCreated)
	inventoryFieldID := decodeCustomFieldDefinition(t, inventoryField).Data.ID
	inventoryFieldPath := inventoryPath + "/custom-field-definitions/" + inventoryFieldID
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions", inventoryPath+"/custom-field-definitions?limit=10", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", inventoryFieldPath, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", inventoryFieldPath, "Bearer dev:owner", map[string]any{"displayName": "Expiration Date"}, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}/archive", inventoryFieldPath+"/archive", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}/restore", inventoryFieldPath+"/restore", "Bearer dev:owner", nil, http.StatusOK)

	assetCreate := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/assets", inventoryPath+"/assets", "Bearer dev:owner", map[string]any{
		"kind":        "item",
		"title":       "Cordless Drill",
		"description": "Garage shelf",
	}, http.StatusCreated)
	assetID := decodeAsset(t, assetCreate).Data.ID
	assetPath := inventoryPath + "/assets/" + assetID
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/assets", inventoryPath+"/assets?limit=10", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", assetPath, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", assetPath, "Bearer dev:owner", map[string]any{"title": "Impact Driver"}, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/archive", assetPath+"/archive", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/restore", assetPath+"/restore", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/search/assets", tenantPath+"/search/assets?q=Impact&limit=10", "Bearer dev:owner", nil, http.StatusOK)

	attachmentCreate := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments", assetPath+"/attachments", "Bearer dev:owner", map[string]any{
		"fileName":      "receipt.png",
		"contentType":   "image/png",
		"contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent()),
	}, http.StatusCreated)
	attachmentID := decodeAttachment(t, attachmentCreate).Data.ID
	attachmentPath := assetPath + "/attachments/" + attachmentID
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments", assetPath+"/attachments?limit=10", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}", attachmentPath, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/content", attachmentPath+"/content", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/thumbnail", attachmentPath+"/thumbnail", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/archive", attachmentPath+"/archive", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/restore", attachmentPath+"/restore", "Bearer dev:owner", nil, http.StatusOK)

	directUpload := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads", assetPath+"/attachments/direct-uploads", "Bearer dev:owner", map[string]any{
		"fileName":    "manual.png",
		"contentType": "image/png",
		"sizeBytes":   len(pngAttachmentContent()),
	}, http.StatusCreated)
	uploadID := decodeDirectUpload(t, directUpload).Data.UploadID
	directComplete := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads/{uploadId}/complete", assetPath+"/attachments/direct-uploads/"+uploadID+"/complete", "Bearer dev:owner", nil, http.StatusCreated)
	directAttachmentID := decodeAttachment(t, directComplete).Data.ID
	directAttachmentPath := assetPath + "/attachments/" + directAttachmentID

	coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants", inventoryPath+"/access-grants", "Bearer dev:owner", map[string]any{"principalId": "viewer", "relationship": "viewer"}, http.StatusCreated)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants", inventoryPath+"/access-grants?limit=10", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}", inventoryPath+"/access-grants/viewer/viewer", "Bearer dev:owner", nil, http.StatusOK)

	inviteAcceptCreate := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations", inventoryPath+"/access-invitations", "Bearer dev:owner", map[string]any{"email": "invitee@example.com", "relationship": "viewer"}, http.StatusCreated)
	inviteAccept := decodeInventoryAccessInvitation(t, inviteAcceptCreate).Data
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations", inventoryPath+"/access-invitations?limit=10", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}", inventoryPath+"/access-invitations/"+inviteAccept.ID, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/expiration", inventoryPath+"/access-invitations/"+inviteAccept.ID+"/expiration", "Bearer dev:owner", map[string]any{"expiresAt": "2999-01-01T00:00:00Z"}, http.StatusOK)
	coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/accept", inventoryPath+"/access-invitations/"+inviteAccept.ID+"/accept", "Bearer dev:invitee:invitee@example.com", map[string]any{"acceptanceToken": inviteAccept.AcceptanceToken}, http.StatusOK)

	inviteDeleteCreate := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations", inventoryPath+"/access-invitations", "Bearer dev:owner", map[string]any{"email": "cancel@example.com", "relationship": "viewer"}, http.StatusCreated)
	inviteDeleteID := decodeInventoryAccessInvitation(t, inviteDeleteCreate).Data.ID
	coverage.request(t, server, http.MethodPatch, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/cancel", inventoryPath+"/access-invitations/"+inviteDeleteID+"/cancel", "Bearer dev:owner", nil, http.StatusNoContent)

	undoAssetCreate := coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/assets", inventoryPath+"/assets", "Bearer dev:owner", map[string]any{"kind": "item", "title": "Undo Target"}, http.StatusCreated)
	undoAssetID := decodeAsset(t, undoAssetCreate).Data.ID
	auditInventory := coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/audit-records", inventoryPath+"/audit-records?limit=50", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/audit-records", tenantPath+"/audit-records?limit=50", "Bearer dev:owner", nil, http.StatusOK)
	operationID := operationIDForTarget(t, decodeAuditRecordList(t, auditInventory).Data, undoAssetID)
	coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/undoable-operations/{operationId}/undo", inventoryPath+"/undoable-operations/"+operationID+"/undo", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPost, "/tenants/{tenantId}/inventories/{inventoryId}/undoable-operations/{operationId}/redo", inventoryPath+"/undoable-operations/"+operationID+"/redo", "Bearer dev:owner", nil, http.StatusOK)

	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}", inventoryPath+"/access-grants/viewer/viewer", "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}", inventoryPath+"/access-grants/invitee/viewer", "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}", inventoryPath+"/access-invitations/"+inviteDeleteID, "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}", attachmentPath, "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}", directAttachmentPath, "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", inventoryFieldPath, "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/custom-field-definitions/{definitionId}", tenantFieldPath, "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}", inventoryTypePath, "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}", tenantTypePath, "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", assetPath, "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", inventoryPath+"/assets/"+undoAssetID, "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}/inventories/{inventoryId}", inventoryPath, "Bearer dev:owner", nil, http.StatusNoContent)
	coverage.request(t, server, http.MethodDelete, "/tenants/{tenantId}", tenantPath, "Bearer dev:owner", nil, http.StatusNoContent)

	return coverage
}

func adversarialScenarioOperations(t *testing.T) executedScenarioCoverage {
	t.Helper()

	setup := realUseAdversarialFixture(t)
	coverage := newExecutedScenarioCoverage("adversarial")
	for _, request := range setup.requests {
		coverage.request(t, setup.server, request.method, request.template, request.path, "", request.body, http.StatusUnauthorized)
	}
	return coverage
}

type adversarialFixture struct {
	server   *http.Server
	requests []scenarioRequest
}

type scenarioRequest struct {
	method   string
	template string
	path     string
	body     any
}

func realUseAdversarialFixture(t *testing.T) adversarialFixture {
	t.Helper()

	directUploads := &httpFakeDirectAttachmentUploader{}
	server := NewServer(":0", newSeededMediaTestApp(t, seededState{}, directUploads, &httpFakeImageProcessor{thumbnailContent: []byte("thumbnail")}))

	tenantCreate := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:owner", map[string]any{"name": "Home"})
	requireStatus(t, tenantCreate, http.StatusCreated)
	tenantID := decodeTenant(t, tenantCreate).Data.ID
	tenantPath := "/tenants/" + tenantID

	inventoryCreate := performRequest(server, http.MethodPost, tenantPath+"/inventories", "Bearer dev:owner", map[string]any{"name": "Tools"})
	requireStatus(t, inventoryCreate, http.StatusCreated)
	inventoryID := decodeScenarioInventory(t, inventoryCreate).Data.ID
	inventoryPath := tenantPath + "/inventories/" + inventoryID

	assetCreate := performRequest(server, http.MethodPost, inventoryPath+"/assets", "Bearer dev:owner", map[string]any{"kind": "item", "title": "Drill"})
	requireStatus(t, assetCreate, http.StatusCreated)
	assetID := decodeAsset(t, assetCreate).Data.ID
	assetPath := inventoryPath + "/assets/" + assetID

	attachmentCreate := performRequest(server, http.MethodPost, assetPath+"/attachments", "Bearer dev:owner", map[string]any{
		"fileName":      "receipt.png",
		"contentType":   "image/png",
		"contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent()),
	})
	requireStatus(t, attachmentCreate, http.StatusCreated)
	attachmentID := decodeAttachment(t, attachmentCreate).Data.ID
	attachmentPath := assetPath + "/attachments/" + attachmentID

	directUpload := performRequest(server, http.MethodPost, assetPath+"/attachments/direct-uploads", "Bearer dev:owner", map[string]any{
		"fileName":    "manual.png",
		"contentType": "image/png",
		"sizeBytes":   len(pngAttachmentContent()),
	})
	requireStatus(t, directUpload, http.StatusCreated)
	uploadID := decodeDirectUpload(t, directUpload).Data.UploadID

	tenantType := performRequest(server, http.MethodPost, tenantPath+"/custom-asset-types", "Bearer dev:owner", map[string]any{"key": "medicine", "displayName": "Medicine"})
	requireStatus(t, tenantType, http.StatusCreated)
	tenantTypePath := tenantPath + "/custom-asset-types/" + decodeCustomAssetType(t, tenantType).Data.ID
	inventoryType := performRequest(server, http.MethodPost, inventoryPath+"/custom-asset-types", "Bearer dev:owner", map[string]any{"key": "tool", "displayName": "Tool"})
	requireStatus(t, inventoryType, http.StatusCreated)
	inventoryTypePath := inventoryPath + "/custom-asset-types/" + decodeCustomAssetType(t, inventoryType).Data.ID

	tenantField := performRequest(server, http.MethodPost, tenantPath+"/custom-field-definitions", "Bearer dev:owner", map[string]any{"key": "serial", "displayName": "Serial", "type": "text"})
	requireStatus(t, tenantField, http.StatusCreated)
	tenantFieldPath := tenantPath + "/custom-field-definitions/" + decodeCustomFieldDefinition(t, tenantField).Data.ID
	inventoryField := performRequest(server, http.MethodPost, inventoryPath+"/custom-field-definitions", "Bearer dev:owner", map[string]any{"key": "expires", "displayName": "Expires", "type": "date"})
	requireStatus(t, inventoryField, http.StatusCreated)
	inventoryFieldPath := inventoryPath + "/custom-field-definitions/" + decodeCustomFieldDefinition(t, inventoryField).Data.ID

	grant := performRequest(server, http.MethodPost, inventoryPath+"/access-grants", "Bearer dev:owner", map[string]any{"principalId": "viewer", "relationship": "viewer"})
	requireStatus(t, grant, http.StatusCreated)
	invite := performRequest(server, http.MethodPost, inventoryPath+"/access-invitations", "Bearer dev:owner", map[string]any{"email": "invitee@example.com", "relationship": "viewer"})
	requireStatus(t, invite, http.StatusCreated)
	inviteBody := decodeInventoryAccessInvitation(t, invite).Data
	invitePath := inventoryPath + "/access-invitations/" + inviteBody.ID

	providerProfile := performRequest(server, http.MethodPost, tenantPath+"/provider-profiles", "Bearer dev:owner", map[string]any{
		"capability":   "language_inference",
		"providerKind": "gemini",
		"displayName":  "Google Gemini",
	})
	requireStatus(t, providerProfile, http.StatusCreated)
	providerProfilePath := tenantPath + "/provider-profiles/" + decodeProviderProfile(t, providerProfile).Data.ID

	auditResponse := performRequest(server, http.MethodGet, inventoryPath+"/audit-records?limit=50", "Bearer dev:owner", nil)
	requireStatus(t, auditResponse, http.StatusOK)
	operationID := operationIDForTarget(t, decodeAuditRecordList(t, auditResponse).Data, assetID)

	return adversarialFixture{server: server, requests: []scenarioRequest{
		{method: http.MethodGet, template: "/me", path: "/me"},
		{method: http.MethodGet, template: "/me/tenants", path: "/me/tenants?limit=10"},
		{method: http.MethodPost, template: "/tenants", path: "/tenants", body: map[string]any{"name": "Blocked"}},
		{method: http.MethodGet, template: "/tenants/{tenantId}", path: tenantPath},
		{method: http.MethodPatch, template: "/tenants/{tenantId}", path: tenantPath, body: map[string]any{"name": "Blocked"}},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/archive", path: tenantPath + "/archive"},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/restore", path: tenantPath + "/restore"},
		{method: http.MethodDelete, template: "/tenants/{tenantId}", path: tenantPath},
		{method: http.MethodPost, template: "/tenants/{tenantId}/provider-profiles", path: tenantPath + "/provider-profiles", body: map[string]any{"capability": "language_inference", "providerKind": "gemini", "displayName": "Blocked"}},
		{method: http.MethodGet, template: "/tenants/{tenantId}/provider-profiles", path: tenantPath + "/provider-profiles"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/provider-profiles/{providerProfileId}", path: providerProfilePath},
		{method: http.MethodPut, template: "/tenants/{tenantId}/provider-profiles/{providerProfileId}/credential", path: providerProfilePath + "/credential", body: map[string]any{"purpose": "api_key", "credential": "blocked"}},
		{method: http.MethodPost, template: "/tenants/{tenantId}/provider-profiles/{providerProfileId}/enable", path: providerProfilePath + "/enable"},
		{method: http.MethodPost, template: "/tenants/{tenantId}/provider-profiles/{providerProfileId}/disable", path: providerProfilePath + "/disable"},
		{method: http.MethodPost, template: "/tenants/{tenantId}/provider-profiles/{providerProfileId}/archive", path: providerProfilePath + "/archive"},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories", path: tenantPath + "/inventories", body: map[string]any{"name": "Blocked"}},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories", path: tenantPath + "/inventories?limit=10"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}", path: inventoryPath},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}", path: inventoryPath, body: map[string]any{"name": "Blocked"}},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/archive", path: inventoryPath + "/archive"},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/restore", path: inventoryPath + "/restore"},
		{method: http.MethodDelete, template: "/tenants/{tenantId}/inventories/{inventoryId}", path: inventoryPath},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets", path: inventoryPath + "/assets", body: map[string]any{"kind": "item", "title": "Blocked"}},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets", path: inventoryPath + "/assets?limit=10"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", path: assetPath},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", path: assetPath, body: map[string]any{"title": "Blocked"}},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/archive", path: assetPath + "/archive"},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/restore", path: assetPath + "/restore"},
		{method: http.MethodDelete, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", path: assetPath},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments", path: assetPath + "/attachments", body: map[string]any{"fileName": "blocked.png", "contentType": "image/png", "contentBase64": base64.StdEncoding.EncodeToString(pngAttachmentContent())}},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads", path: assetPath + "/attachments/direct-uploads", body: map[string]any{"fileName": "blocked.png", "contentType": "image/png", "sizeBytes": 1}},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads/{uploadId}/complete", path: assetPath + "/attachments/direct-uploads/" + uploadID + "/complete"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments", path: assetPath + "/attachments?limit=10"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}", path: attachmentPath},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/content", path: attachmentPath + "/content"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/thumbnail", path: attachmentPath + "/thumbnail"},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/archive", path: attachmentPath + "/archive"},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/restore", path: attachmentPath + "/restore"},
		{method: http.MethodDelete, template: "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}", path: attachmentPath},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories/{inventoryId}/access-grants", path: inventoryPath + "/access-grants", body: map[string]any{"principalId": "blocked", "relationship": "viewer"}},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/access-grants", path: inventoryPath + "/access-grants?limit=10"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}", path: inventoryPath + "/access-grants/viewer/viewer"},
		{method: http.MethodDelete, template: "/tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}", path: inventoryPath + "/access-grants/viewer/viewer"},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations", path: inventoryPath + "/access-invitations", body: map[string]any{"email": "blocked@example.com", "relationship": "viewer"}},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations", path: inventoryPath + "/access-invitations?limit=10"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}", path: invitePath},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/expiration", path: invitePath + "/expiration", body: map[string]any{"expiresAt": "2999-01-01T00:00:00Z"}},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/cancel", path: invitePath + "/cancel"},
		{method: http.MethodDelete, template: "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}", path: invitePath},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/accept", path: invitePath + "/accept", body: map[string]any{"acceptanceToken": inviteBody.AcceptanceToken}},
		{method: http.MethodPost, template: "/tenants/{tenantId}/custom-asset-types", path: tenantPath + "/custom-asset-types", body: map[string]any{"key": "blocked", "displayName": "Blocked"}},
		{method: http.MethodGet, template: "/tenants/{tenantId}/custom-asset-types", path: tenantPath + "/custom-asset-types?limit=10"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}", path: tenantTypePath},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}", path: tenantTypePath, body: map[string]any{"displayName": "Blocked"}},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}/archive", path: tenantTypePath + "/archive"},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}/restore", path: tenantTypePath + "/restore"},
		{method: http.MethodDelete, template: "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}", path: tenantTypePath},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types", path: inventoryPath + "/custom-asset-types", body: map[string]any{"key": "blocked", "displayName": "Blocked"}},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types", path: inventoryPath + "/custom-asset-types?limit=10"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}", path: inventoryTypePath},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}", path: inventoryTypePath, body: map[string]any{"displayName": "Blocked"}},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/archive", path: inventoryTypePath + "/archive"},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/restore", path: inventoryTypePath + "/restore"},
		{method: http.MethodDelete, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}", path: inventoryTypePath},
		{method: http.MethodPost, template: "/tenants/{tenantId}/custom-field-definitions", path: tenantPath + "/custom-field-definitions", body: map[string]any{"key": "blocked", "displayName": "Blocked", "type": "text"}},
		{method: http.MethodGet, template: "/tenants/{tenantId}/custom-field-definitions", path: tenantPath + "/custom-field-definitions?limit=10"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/custom-field-definitions/{definitionId}", path: tenantFieldPath},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/custom-field-definitions/{definitionId}", path: tenantFieldPath, body: map[string]any{"displayName": "Blocked"}},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/custom-field-definitions/{definitionId}/archive", path: tenantFieldPath + "/archive"},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/custom-field-definitions/{definitionId}/restore", path: tenantFieldPath + "/restore"},
		{method: http.MethodDelete, template: "/tenants/{tenantId}/custom-field-definitions/{definitionId}", path: tenantFieldPath},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions", path: inventoryPath + "/custom-field-definitions", body: map[string]any{"key": "blocked", "displayName": "Blocked", "type": "text"}},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions", path: inventoryPath + "/custom-field-definitions?limit=10"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", path: inventoryFieldPath},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", path: inventoryFieldPath, body: map[string]any{"displayName": "Blocked"}},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}/archive", path: inventoryFieldPath + "/archive"},
		{method: http.MethodPatch, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}/restore", path: inventoryFieldPath + "/restore"},
		{method: http.MethodDelete, template: "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", path: inventoryFieldPath},
		{method: http.MethodGet, template: "/tenants/{tenantId}/audit-records", path: tenantPath + "/audit-records?limit=10"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/inventories/{inventoryId}/audit-records", path: inventoryPath + "/audit-records?limit=10"},
		{method: http.MethodGet, template: "/tenants/{tenantId}/search/assets", path: tenantPath + "/search/assets?q=Drill&limit=10"},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories/{inventoryId}/undoable-operations/{operationId}/undo", path: inventoryPath + "/undoable-operations/" + operationID + "/undo"},
		{method: http.MethodPost, template: "/tenants/{tenantId}/inventories/{inventoryId}/undoable-operations/{operationId}/redo", path: inventoryPath + "/undoable-operations/" + operationID + "/redo"},
	}}
}

func newExecutedScenarioCoverage(name string) executedScenarioCoverage {
	return executedScenarioCoverage{name: name, operation: map[string]struct{}{}}
}

func (c executedScenarioCoverage) request(t *testing.T, server *http.Server, method string, template string, path string, authorization string, body any, expected int) *httptest.ResponseRecorder {
	t.Helper()
	c.operation[operationKey(method, template)] = struct{}{}
	response := performRequest(server, method, path, authorization, body)
	if response.Code != expected {
		t.Fatalf("%s scenario expected %s %s status %d, got %d with body %s", c.name, method, path, expected, response.Code, response.Body.String())
	}
	return response
}

func decodeScenarioInventory(t *testing.T, response *httptest.ResponseRecorder) inventoryBody {
	t.Helper()

	var body inventoryBody
	decodeBody(t, response, &body)
	return body
}

func assertScenarioCoversOpenAPI(t *testing.T, coverage executedScenarioCoverage, openAPIOperations []string) {
	t.Helper()

	openAPIOperationSet := make(map[string]struct{}, len(openAPIOperations))
	for _, operation := range openAPIOperations {
		openAPIOperationSet[operation] = struct{}{}
		if _, ok := coverage.operation[operation]; !ok {
			t.Fatalf("%s scenario did not exercise OpenAPI operation %s", coverage.name, operation)
		}
	}
	for operation := range coverage.operation {
		if _, ok := openAPIOperationSet[operation]; !ok {
			t.Fatalf("%s scenario recorded non-OpenAPI operation %s", coverage.name, operation)
		}
	}
}

func generatedOpenAPIOperations(t *testing.T) []string {
	t.Helper()

	server := NewServer(":0", newTestApp(&fakeObserver{}, "unused-id"))
	response := performRequest(server, http.MethodGet, "/openapi.json", "", nil)
	requireStatus(t, response, http.StatusOK)

	var body struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	decodeBody(t, response, &body)

	var operations []string
	for path, pathItem := range body.Paths {
		for method := range pathItem {
			if !isOpenAPIMethod(method) {
				continue
			}
			operations = append(operations, operationKey(method, path))
		}
	}
	sort.Strings(operations)
	return operations
}

func operationKey(method string, path string) string {
	return strings.ToUpper(method) + " " + path
}

func isOpenAPIMethod(method string) bool {
	switch strings.ToLower(method) {
	case "get", "post", "put", "patch", "delete":
		return true
	default:
		return false
	}
}

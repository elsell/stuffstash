package httpserver

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func TestInventoryAccessInvitationsCreateAcceptAndRevoke(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FB0"
	application := newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
			{id: otherInventoryID, tenantID: tenantID, name: "Supplies", owner: "inventory-owner"},
		},
		ids: []string{
			"invite-viewer", "audit-invite-viewer",
			"audit-wrong-token-accept", "wrong-token-accept-event",
			"audit-wrong-email-accept", "wrong-email-accept-event",
			"audit-wrong-tenant-accept", "wrong-tenant-accept-event",
			"audit-wrong-inventory-accept", "wrong-inventory-accept-event",
			"audit-accept-viewer", "accept-viewer-event", "accept-viewer-claim",
			"audit-already-accepted", "already-accepted-event",
			"audit-editor-grant", "editor-grant-event", "editor-grant-claim",
			"invite-pending", "audit-invite-pending",
			"audit-wrong-inventory-expiration",
			"audit-list-invitations", "audit-list-invitations-page-two", "audit-list-invitations-all",
			"audit-list-pending", "audit-list-expired-before-update",
			"audit-expire-pending",
			"audit-list-expired-after-update", "audit-list-pending-after-update",
			"audit-expired-pending-accept", "expired-pending-accept-event",
			"audit-update-accepted-expiration",
			"invite-editor", "audit-invite-editor", "audit-mark-editor-revoked", "audit-revoke-editor",
			"audit-revoked-accept", "revoked-accept-event",
			"invite-cancel", "audit-invite-cancel", "audit-cancel-invite",
		},
	})
	server := NewServer(":0", application)

	unauthenticatedCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "", map[string]string{
		"email":        "viewer@example.com",
		"relationship": "viewer",
	})
	if unauthenticatedCreate.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated invitation create status %d, got %d with body %s", http.StatusUnauthorized, unauthenticatedCreate.Code, unauthenticatedCreate.Body.String())
	}
	assertSafeError(t, unauthenticatedCreate, "authentication_required", "Authentication required.")

	invitationResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:inventory-owner", map[string]string{
		"email":        "Viewer@Example.COM",
		"relationship": "viewer",
	})
	if invitationResponse.Code != http.StatusCreated {
		t.Fatalf("expected invitation create status %d, got %d with body %s", http.StatusCreated, invitationResponse.Code, invitationResponse.Body.String())
	}
	if got := invitationResponse.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected invitation creation to disable caching, got Cache-Control %q", got)
	}
	invitation := decodeInventoryAccessInvitation(t, invitationResponse).Data
	if invitation.Email != "viewer@example.com" || invitation.Status != "pending" || invitation.Relationship != "viewer" {
		t.Fatalf("unexpected invitation response: %+v", invitation)
	}
	if invitation.AcceptanceToken != "" {
		t.Fatalf("expected creation response to expose only the complete link, got bare token")
	}
	if strings.Contains(invitationResponse.Body.String(), `"acceptanceToken"`) {
		t.Fatalf("expected bare acceptanceToken field to be absent from creation JSON: %s", invitationResponse.Body.String())
	}
	invitationToken := acceptanceTokenFromInviteURL(t, invitation.InviteURL)
	inviteURL, err := url.Parse(invitation.InviteURL)
	if err != nil {
		t.Fatalf("parse invitation URL: %v", err)
	}
	if inviteURL.Scheme != "https" || inviteURL.Host != "stash.example.test" || inviteURL.Path != "/invitations/accept" || inviteURL.RawQuery == "" {
		t.Fatalf("expected canonical invitation origin and path, got %q", invitation.InviteURL)
	}
	query := inviteURL.Query()
	if len(query) != 3 || len(query["tenant"]) != 1 || query.Get("tenant") != tenantID || len(query["inventory"]) != 1 || query.Get("inventory") != inventoryID || len(query["invitation"]) != 1 || query.Get("invitation") != invitation.ID || query.Has("token") {
		t.Fatalf("expected exact scoped invitation query without token, got %q", inviteURL.RawQuery)
	}
	fragment, err := url.ParseQuery(inviteURL.Fragment)
	if err != nil || len(fragment) != 1 || len(fragment["token"]) != 1 || fragment.Get("token") != invitationToken {
		t.Fatalf("expected exactly one fragment token, got %q (err=%v)", inviteURL.Fragment, err)
	}
	if invitation.ExpiresAt == "" {
		t.Fatalf("expected invitation response to include expiresAt")
	}
	if invitation.IsExpired {
		t.Fatalf("expected new invitation not to be expired: %+v", invitation)
	}

	previewPath := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/access-invitations/" + invitation.ID + "/preview"
	unauthenticatedPreview := performRequest(server, http.MethodPost, previewPath, "", map[string]string{"acceptanceToken": invitationToken})
	if unauthenticatedPreview.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated preview status %d, got %d with body %s", http.StatusUnauthorized, unauthenticatedPreview.Code, unauthenticatedPreview.Body.String())
	}
	assertSafeError(t, unauthenticatedPreview, "authentication_required", "Authentication required.")

	wrongTokenPreview := performRequest(server, http.MethodPost, previewPath, "Bearer dev:viewer-user:viewer@example.com", map[string]string{"acceptanceToken": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"})
	if wrongTokenPreview.Code != http.StatusNotFound {
		t.Fatalf("expected wrong-token preview status %d, got %d with body %s", http.StatusNotFound, wrongTokenPreview.Code, wrongTokenPreview.Body.String())
	}
	assertSafeError(t, wrongTokenPreview, "invitation_invalid", "This invitation link is invalid.")
	for name, token := range map[string]string{
		"missing":       "",
		"short":         strings.Repeat("A", 42),
		"overlong":      strings.Repeat("A", 44),
		"invalid chars": strings.Repeat("A", 42) + "!",
	} {
		t.Run("malformed preview token "+name, func(t *testing.T) {
			response := performRequest(server, http.MethodPost, previewPath, "Bearer dev:viewer-user:viewer@example.com", map[string]string{"acceptanceToken": token})
			if response.Code != http.StatusNotFound {
				t.Fatalf("expected malformed-token preview status %d, got %d with body %s", http.StatusNotFound, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "invitation_invalid", "This invitation link is invalid.")
		})
	}
	for _, wrongPath := range []string{
		"/tenants/" + otherTenantID + "/inventories/" + inventoryID + "/access-invitations/" + invitation.ID + "/preview",
		"/tenants/" + tenantID + "/inventories/" + otherInventoryID + "/access-invitations/" + invitation.ID + "/preview",
	} {
		response := performRequest(server, http.MethodPost, wrongPath, "Bearer dev:viewer-user:viewer@example.com", map[string]string{"acceptanceToken": invitationToken})
		if response.Code != http.StatusNotFound {
			t.Fatalf("expected wrong-scope preview status %d, got %d with body %s", http.StatusNotFound, response.Code, response.Body.String())
		}
		assertSafeError(t, response, "invitation_invalid", "This invitation link is invalid.")
	}

	wrongEmailPreview := performRequest(server, http.MethodPost, previewPath, "Bearer dev:viewer-user:wrong@example.com", map[string]string{"acceptanceToken": invitationToken})
	if wrongEmailPreview.Code != http.StatusForbidden {
		t.Fatalf("expected wrong-email preview status %d, got %d with body %s", http.StatusForbidden, wrongEmailPreview.Code, wrongEmailPreview.Body.String())
	}
	assertSafeError(t, wrongEmailPreview, "invitation_email_mismatch", "This invitation belongs to another signed-in account.")

	previewResponse := performRequest(server, http.MethodPost, previewPath, "Bearer dev:viewer-user:viewer@example.com", map[string]string{"acceptanceToken": invitationToken})
	if previewResponse.Code != http.StatusOK {
		t.Fatalf("expected preview status %d, got %d with body %s", http.StatusOK, previewResponse.Code, previewResponse.Body.String())
	}
	previewJSON := previewResponse.Body.String()
	preview := decodeInventoryAccessInvitationPreview(t, previewResponse).Data
	if preview.InventoryID != inventoryID || preview.InventoryName != "Tools" || preview.Relationship != "viewer" || preview.Status != "pending" || preview.IsExpired || preview.ExpiresAt == "" {
		t.Fatalf("unexpected safe invitation preview: %+v", preview)
	}
	for _, forbidden := range []string{"viewer@example.com", invitationToken, invitation.InviteURL, "inviterPrincipalId", "acceptedPrincipalId", "tokenHash"} {
		if strings.Contains(previewJSON, forbidden) {
			t.Fatalf("preview response leaked %q: %s", forbidden, previewJSON)
		}
	}
	previewDidNotGrant := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", nil)
	if previewDidNotGrant.Code != http.StatusForbidden {
		t.Fatalf("expected preview not to grant access, got %d with body %s", previewDidNotGrant.Code, previewDidNotGrant.Body.String())
	}

	unauthenticatedAccept := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "", map[string]string{
		"acceptanceToken": invitationToken,
	})
	if unauthenticatedAccept.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated invitation accept status %d, got %d with body %s", http.StatusUnauthorized, unauthenticatedAccept.Code, unauthenticatedAccept.Body.String())
	}
	assertSafeError(t, unauthenticatedAccept, "authentication_required", "Authentication required.")

	missingEmailAccept := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user", map[string]string{
		"acceptanceToken": invitationToken,
	})
	if missingEmailAccept.Code != http.StatusForbidden {
		t.Fatalf("expected missing email accept status %d, got %d with body %s", http.StatusForbidden, missingEmailAccept.Code, missingEmailAccept.Body.String())
	}
	assertSafeError(t, missingEmailAccept, "forbidden", "Forbidden.")

	missingTokenAccept := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": "",
	})
	if missingTokenAccept.Code != http.StatusForbidden {
		t.Fatalf("expected missing token accept status %d, got %d with body %s", http.StatusForbidden, missingTokenAccept.Code, missingTokenAccept.Body.String())
	}
	assertSafeError(t, missingTokenAccept, "forbidden", "Forbidden.")

	wrongTokenAccept := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": "wrong-token",
	})
	if wrongTokenAccept.Code != http.StatusForbidden {
		t.Fatalf("expected wrong token accept status %d, got %d with body %s", http.StatusForbidden, wrongTokenAccept.Code, wrongTokenAccept.Body.String())
	}
	assertSafeError(t, wrongTokenAccept, "forbidden", "Forbidden.")

	wrongEmailAccept := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:wrong@example.com", map[string]string{
		"acceptanceToken": invitationToken,
	})
	if wrongEmailAccept.Code != http.StatusForbidden {
		t.Fatalf("expected wrong email accept status %d, got %d with body %s", http.StatusForbidden, wrongEmailAccept.Code, wrongEmailAccept.Body.String())
	}
	assertSafeError(t, wrongEmailAccept, "forbidden", "Forbidden.")

	wrongTenantAccept := performRequest(server, http.MethodPost, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": invitationToken,
	})
	if wrongTenantAccept.Code != http.StatusForbidden {
		t.Fatalf("expected wrong tenant accept status %d, got %d with body %s", http.StatusForbidden, wrongTenantAccept.Code, wrongTenantAccept.Body.String())
	}
	assertSafeError(t, wrongTenantAccept, "forbidden", "Forbidden.")

	wrongInventoryAccept := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+otherInventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": invitationToken,
	})
	if wrongInventoryAccept.Code != http.StatusForbidden {
		t.Fatalf("expected wrong inventory accept status %d, got %d with body %s", http.StatusForbidden, wrongInventoryAccept.Code, wrongInventoryAccept.Body.String())
	}
	assertSafeError(t, wrongInventoryAccept, "forbidden", "Forbidden.")

	acceptResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": invitationToken,
	})
	if acceptResponse.Code != http.StatusOK {
		t.Fatalf("expected invitation accept status %d, got %d with body %s", http.StatusOK, acceptResponse.Code, acceptResponse.Body.String())
	}
	accepted := decodeInventoryAccessInvitationAcceptance(t, acceptResponse).Data
	if accepted.Invitation.Status != "accepted" || accepted.Invitation.AcceptedPrincipalID != "viewer-user" {
		t.Fatalf("expected accepted invitation, got %+v", accepted.Invitation)
	}
	assertInventoryAccessGrant(t, accepted.Grant, tenantID, inventoryID, "viewer-user", "viewer")

	acceptedPreviewResponse := performRequest(server, http.MethodPost, previewPath, "Bearer dev:viewer-user:viewer@example.com", map[string]string{"acceptanceToken": invitationToken})
	if acceptedPreviewResponse.Code != http.StatusOK {
		t.Fatalf("expected accepted invitation preview status %d, got %d with body %s", http.StatusOK, acceptedPreviewResponse.Code, acceptedPreviewResponse.Body.String())
	}
	acceptedPreview := decodeInventoryAccessInvitationPreview(t, acceptedPreviewResponse).Data
	if acceptedPreview.Status != "accepted" || acceptedPreview.InventoryID != inventoryID {
		t.Fatalf("expected safe already-accepted preview, got %+v", acceptedPreview)
	}

	acceptAlreadyAccepted := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": invitationToken,
	})
	if acceptAlreadyAccepted.Code != http.StatusForbidden {
		t.Fatalf("expected already accepted invite status %d, got %d with body %s", http.StatusForbidden, acceptAlreadyAccepted.Code, acceptAlreadyAccepted.Body.String())
	}
	assertSafeError(t, acceptAlreadyAccepted, "forbidden", "Forbidden.")

	viewerListAssets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", nil)
	if viewerListAssets.Code != http.StatusOK {
		t.Fatalf("expected accepted viewer list assets status %d, got %d with body %s", http.StatusOK, viewerListAssets.Code, viewerListAssets.Body.String())
	}
	viewerCreateAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", map[string]string{"kind": "item", "title": "Drill"})
	if viewerCreateAsset.Code != http.StatusForbidden {
		t.Fatalf("expected accepted viewer create asset status %d, got %d with body %s", http.StatusForbidden, viewerCreateAsset.Code, viewerCreateAsset.Body.String())
	}

	editorGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "editor-user",
		"relationship": "editor",
	})
	if editorGrant.Code != http.StatusCreated {
		t.Fatalf("expected editor grant status %d, got %d with body %s", http.StatusCreated, editorGrant.Code, editorGrant.Body.String())
	}

	pendingInviteResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:inventory-owner", map[string]string{
		"email":        "pending@example.com",
		"relationship": "viewer",
	})
	if pendingInviteResponse.Code != http.StatusCreated {
		t.Fatalf("expected pending invitation create status %d, got %d with body %s", http.StatusCreated, pendingInviteResponse.Code, pendingInviteResponse.Body.String())
	}
	pendingInvite := decodeInventoryAccessInvitation(t, pendingInviteResponse).Data
	if pendingInvite.AcceptanceToken != "" {
		t.Fatalf("expected pending invite create response to omit bare token")
	}
	pendingInviteToken := acceptanceTokenFromInviteURL(t, pendingInvite.InviteURL)

	wrongTenantInviteList := performRequest(server, http.MethodGet, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/access-invitations?limit=50", "Bearer dev:inventory-owner", nil)
	if wrongTenantInviteList.Code != http.StatusNotFound {
		t.Fatalf("expected wrong-tenant invitation list status %d, got %d with body %s", http.StatusNotFound, wrongTenantInviteList.Code, wrongTenantInviteList.Body.String())
	}
	assertSafeError(t, wrongTenantInviteList, "resource_not_found", "Resource not found.")

	wrongInventoryInviteList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/01ARZ3NDEKTSV4RRFFQ69G5FB1/access-invitations?limit=50", "Bearer dev:inventory-owner", nil)
	if wrongInventoryInviteList.Code != http.StatusNotFound {
		t.Fatalf("expected wrong-inventory invitation list status %d, got %d with body %s", http.StatusNotFound, wrongInventoryInviteList.Code, wrongInventoryInviteList.Body.String())
	}
	assertSafeError(t, wrongInventoryInviteList, "resource_not_found", "Resource not found.")

	wrongTenantExpiration := performRequest(server, http.MethodPatch, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/access-invitations/"+pendingInvite.ID+"/expiration", "Bearer dev:inventory-owner", map[string]string{
		"expiresAt": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	})
	if wrongTenantExpiration.Code != http.StatusNotFound {
		t.Fatalf("expected wrong-tenant invitation expiration update status %d, got %d with body %s", http.StatusNotFound, wrongTenantExpiration.Code, wrongTenantExpiration.Body.String())
	}
	assertSafeError(t, wrongTenantExpiration, "resource_not_found", "Resource not found.")

	wrongInventoryExpiration := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+otherInventoryID+"/access-invitations/"+pendingInvite.ID+"/expiration", "Bearer dev:inventory-owner", map[string]string{
		"expiresAt": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	})
	if wrongInventoryExpiration.Code != http.StatusNotFound {
		t.Fatalf("expected wrong-inventory invitation expiration update status %d, got %d with body %s", http.StatusNotFound, wrongInventoryExpiration.Code, wrongInventoryExpiration.Body.String())
	}
	assertSafeError(t, wrongInventoryExpiration, "resource_not_found", "Resource not found.")

	missingInventoryExpiration := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/01ARZ3NDEKTSV4RRFFQ69G5FB1/access-invitations/"+pendingInvite.ID+"/expiration", "Bearer dev:inventory-owner", map[string]string{
		"expiresAt": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	})
	if missingInventoryExpiration.Code != http.StatusNotFound {
		t.Fatalf("expected missing-inventory invitation expiration update status %d, got %d with body %s", http.StatusNotFound, missingInventoryExpiration.Code, missingInventoryExpiration.Body.String())
	}
	assertSafeError(t, missingInventoryExpiration, "resource_not_found", "Resource not found.")

	firstInvitePageResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations?limit=1", "Bearer dev:inventory-owner", nil)
	if firstInvitePageResponse.Code != http.StatusOK {
		t.Fatalf("expected first invitation page status %d, got %d with body %s", http.StatusOK, firstInvitePageResponse.Code, firstInvitePageResponse.Body.String())
	}
	firstInvitePage := decodeInventoryAccessInvitationList(t, firstInvitePageResponse)
	if len(firstInvitePage.Data) != 1 || firstInvitePage.Meta.Pagination == nil || firstInvitePage.Meta.Pagination.Limit != 1 || !firstInvitePage.Meta.Pagination.HasMore || firstInvitePage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected first invitation page metadata, got %+v", firstInvitePage)
	}
	if firstInvitePage.Data[0].AcceptanceToken != "" {
		t.Fatalf("expected list response to redact acceptance token, got %+v", firstInvitePage.Data[0])
	}

	secondInvitePageResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations?limit=1&cursor="+*firstInvitePage.Meta.Pagination.NextCursor, "Bearer dev:inventory-owner", nil)
	if secondInvitePageResponse.Code != http.StatusOK {
		t.Fatalf("expected second invitation page status %d, got %d with body %s", http.StatusOK, secondInvitePageResponse.Code, secondInvitePageResponse.Body.String())
	}
	secondInvitePage := decodeInventoryAccessInvitationList(t, secondInvitePageResponse)
	if len(secondInvitePage.Data) != 1 || secondInvitePage.Meta.Pagination == nil || secondInvitePage.Meta.Pagination.HasMore || secondInvitePage.Meta.Pagination.NextCursor != nil {
		t.Fatalf("expected final invitation page metadata, got %+v", secondInvitePage)
	}
	if secondInvitePage.Data[0].AcceptanceToken != "" {
		t.Fatalf("expected second list response to redact acceptance token, got %+v", secondInvitePage.Data[0])
	}

	allInviteResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations?limit=50", "Bearer dev:inventory-owner", nil)
	if allInviteResponse.Code != http.StatusOK {
		t.Fatalf("expected invitation list status %d, got %d with body %s", http.StatusOK, allInviteResponse.Code, allInviteResponse.Body.String())
	}
	allInvites := decodeInventoryAccessInvitationList(t, allInviteResponse)
	if len(allInvites.Data) != 2 {
		t.Fatalf("expected accepted and pending invitations in all filter, got %+v", allInvites.Data)
	}

	pendingInviteListBeforeUpdate := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations?status=pending&limit=50", "Bearer dev:inventory-owner", nil)
	if pendingInviteListBeforeUpdate.Code != http.StatusOK {
		t.Fatalf("expected pending invitation list status %d, got %d with body %s", http.StatusOK, pendingInviteListBeforeUpdate.Code, pendingInviteListBeforeUpdate.Body.String())
	}
	pendingBefore := decodeInventoryAccessInvitationList(t, pendingInviteListBeforeUpdate)
	if len(pendingBefore.Data) != 1 || pendingBefore.Data[0].ID != pendingInvite.ID || pendingBefore.Data[0].IsExpired {
		t.Fatalf("expected only unexpired pending invitation, got %+v", pendingBefore.Data)
	}

	expiredBeforeUpdate := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations?status=expired&limit=50", "Bearer dev:inventory-owner", nil)
	if expiredBeforeUpdate.Code != http.StatusOK {
		t.Fatalf("expected expired invitation list status %d, got %d with body %s", http.StatusOK, expiredBeforeUpdate.Code, expiredBeforeUpdate.Body.String())
	}
	if expiredBefore := decodeInventoryAccessInvitationList(t, expiredBeforeUpdate); len(expiredBefore.Data) != 0 {
		t.Fatalf("expected no expired invitations before update, got %+v", expiredBefore.Data)
	}

	badStatusInviteList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations?status=deleted", "Bearer dev:inventory-owner", nil)
	if badStatusInviteList.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected bad invitation status filter status %d, got %d with body %s", http.StatusUnprocessableEntity, badStatusInviteList.Code, badStatusInviteList.Body.String())
	}
	assertErrorCode(t, badStatusInviteList, "invalid_request")

	manualExpiration := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	expirePendingResponse := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+pendingInvite.ID+"/expiration", "Bearer dev:inventory-owner", map[string]string{
		"expiresAt": manualExpiration,
	})
	if expirePendingResponse.Code != http.StatusOK {
		t.Fatalf("expected expire invitation status %d, got %d with body %s", http.StatusOK, expirePendingResponse.Code, expirePendingResponse.Body.String())
	}
	expiredPendingInvite := decodeInventoryAccessInvitation(t, expirePendingResponse).Data
	if !expiredPendingInvite.IsExpired || expiredPendingInvite.AcceptanceToken != "" {
		t.Fatalf("expected expired response without token, got %+v", expiredPendingInvite)
	}

	expiredAfterUpdate := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations?status=expired&limit=50", "Bearer dev:inventory-owner", nil)
	if expiredAfterUpdate.Code != http.StatusOK {
		t.Fatalf("expected expired invitation list after update status %d, got %d with body %s", http.StatusOK, expiredAfterUpdate.Code, expiredAfterUpdate.Body.String())
	}
	expiredAfter := decodeInventoryAccessInvitationList(t, expiredAfterUpdate)
	if len(expiredAfter.Data) != 1 || expiredAfter.Data[0].ID != pendingInvite.ID || !expiredAfter.Data[0].IsExpired {
		t.Fatalf("expected manually expired invitation, got %+v", expiredAfter.Data)
	}

	pendingAfterUpdate := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations?status=pending&limit=50", "Bearer dev:inventory-owner", nil)
	if pendingAfterUpdate.Code != http.StatusOK {
		t.Fatalf("expected pending invitation list after update status %d, got %d with body %s", http.StatusOK, pendingAfterUpdate.Code, pendingAfterUpdate.Body.String())
	}
	if pendingAfter := decodeInventoryAccessInvitationList(t, pendingAfterUpdate); len(pendingAfter.Data) != 0 {
		t.Fatalf("expected no unexpired pending invitations after manual expiration, got %+v", pendingAfter.Data)
	}

	acceptManuallyExpired := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+pendingInvite.ID+"/accept", "Bearer dev:pending-user:pending@example.com", map[string]string{
		"acceptanceToken": pendingInviteToken,
	})
	if acceptManuallyExpired.Code != http.StatusForbidden {
		t.Fatalf("expected manually expired invitation accept status %d, got %d with body %s", http.StatusForbidden, acceptManuallyExpired.Code, acceptManuallyExpired.Body.String())
	}
	assertSafeError(t, acceptManuallyExpired, "forbidden", "Forbidden.")

	updateAcceptedExpiration := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/expiration", "Bearer dev:inventory-owner", map[string]string{
		"expiresAt": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	})
	if updateAcceptedExpiration.Code != http.StatusBadRequest {
		t.Fatalf("expected accepted invitation expiration update status %d, got %d with body %s", http.StatusBadRequest, updateAcceptedExpiration.Code, updateAcceptedExpiration.Body.String())
	}
	assertSafeError(t, updateAcceptedExpiration, "invalid_request", "Invalid request.")

	for _, item := range []struct {
		name string
		auth string
	}{
		{name: "viewer", auth: "Bearer dev:viewer-user"},
		{name: "editor", auth: "Bearer dev:editor-user"},
		{name: "unrelated user", auth: "Bearer dev:unrelated-user"},
	} {
		t.Run(item.name+" cannot create invitations", func(t *testing.T) {
			response := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", item.auth, map[string]string{
				"email":        item.name + "@example.com",
				"relationship": "viewer",
			})
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected invitation create status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
		})

		t.Run(item.name+" cannot list invitations", func(t *testing.T) {
			response := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", item.auth, nil)
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected invitation list status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "forbidden", "Forbidden.")
		})

		t.Run(item.name+" cannot update invitation expiration", func(t *testing.T) {
			response := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+pendingInvite.ID+"/expiration", item.auth, map[string]string{
				"expiresAt": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
			})
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected invitation expiration update status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "forbidden", "Forbidden.")
		})
	}

	editorInviteResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:inventory-owner", map[string]string{
		"email":        "editor@example.com",
		"relationship": "editor",
	})
	if editorInviteResponse.Code != http.StatusCreated {
		t.Fatalf("expected editor invitation create status %d, got %d with body %s", http.StatusCreated, editorInviteResponse.Code, editorInviteResponse.Body.String())
	}
	editorInvite := decodeInventoryAccessInvitation(t, editorInviteResponse).Data
	editorInviteToken := acceptanceTokenFromInviteURL(t, editorInvite.InviteURL)

	for _, item := range []struct {
		name string
		auth string
	}{
		{name: "viewer", auth: "Bearer dev:viewer-user"},
		{name: "editor", auth: "Bearer dev:editor-user"},
		{name: "unrelated user", auth: "Bearer dev:unrelated-user"},
	} {
		t.Run(item.name+" cannot revoke invitations", func(t *testing.T) {
			response := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+editorInvite.ID, item.auth, nil)
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected invitation revoke status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
		})
	}

	revoked, err := application.RevokeInventoryAccessInvitation(context.Background(), app.RevokeInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: "inventory-owner"},
		Source:       audit.SourceAPI,
		TenantID:     tenant.ID(tenantID),
		InventoryID:  inventory.InventoryID(inventoryID),
		InvitationID: editorInvite.ID,
	})
	if err != nil || !revoked {
		t.Fatalf("arrange revoked invitation: revoked=%t err=%v", revoked, err)
	}
	revokedPreviewResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+editorInvite.ID+"/preview", "Bearer dev:editor-user:editor@example.com", map[string]string{"acceptanceToken": editorInviteToken})
	if revokedPreviewResponse.Code != http.StatusOK {
		t.Fatalf("expected revoked invite preview status %d, got %d with body %s", http.StatusOK, revokedPreviewResponse.Code, revokedPreviewResponse.Body.String())
	}
	revokedPreview := decodeInventoryAccessInvitationPreview(t, revokedPreviewResponse).Data
	if revokedPreview.Status != "revoked" || revokedPreview.InventoryID != inventoryID || revokedPreview.IsExpired {
		t.Fatalf("expected safe revoked preview, got %+v", revokedPreview)
	}
	for _, forbidden := range []string{"editor@example.com", editorInviteToken, editorInvite.InviteURL, "inviterPrincipalId", "acceptedPrincipalId", "tokenHash"} {
		if strings.Contains(revokedPreviewResponse.Body.String(), forbidden) {
			t.Fatalf("revoked preview response leaked %q: %s", forbidden, revokedPreviewResponse.Body.String())
		}
	}
	acceptRevoked := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+editorInvite.ID+"/accept", "Bearer dev:editor-user:editor@example.com", map[string]string{
		"acceptanceToken": editorInviteToken,
	})
	if acceptRevoked.Code != http.StatusForbidden {
		t.Fatalf("expected revoked invite accept status %d, got %d with body %s", http.StatusForbidden, acceptRevoked.Code, acceptRevoked.Body.String())
	}
	assertSafeError(t, acceptRevoked, "forbidden", "Forbidden.")
	revokeEditorInvite := performRequest(server, http.MethodDelete, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+editorInvite.ID, "Bearer dev:inventory-owner", nil)
	if revokeEditorInvite.Code != http.StatusNoContent {
		t.Fatalf("expected invite deletion status %d, got %d with body %s", http.StatusNoContent, revokeEditorInvite.Code, revokeEditorInvite.Body.String())
	}

	cancelInviteResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:inventory-owner", map[string]string{
		"email":        "cancel@example.com",
		"relationship": "viewer",
	})
	if cancelInviteResponse.Code != http.StatusCreated {
		t.Fatalf("expected cancellable invitation create status %d, got %d with body %s", http.StatusCreated, cancelInviteResponse.Code, cancelInviteResponse.Body.String())
	}
	cancelInvite := decodeInventoryAccessInvitation(t, cancelInviteResponse).Data
	cancelInviteToken := acceptanceTokenFromInviteURL(t, cancelInvite.InviteURL)
	cancelPath := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/access-invitations/" + cancelInvite.ID + "/cancel"
	for _, item := range []struct {
		name string
		auth string
	}{
		{name: "unauthenticated"},
		{name: "viewer", auth: "Bearer dev:viewer-user"},
		{name: "editor", auth: "Bearer dev:editor-user"},
		{name: "unrelated user", auth: "Bearer dev:unrelated-user"},
	} {
		t.Run(item.name+" cannot cancel invitations", func(t *testing.T) {
			response := performRequest(server, http.MethodPatch, cancelPath, item.auth, nil)
			expected := http.StatusForbidden
			code, message := "forbidden", "Forbidden."
			if item.auth == "" {
				expected = http.StatusUnauthorized
				code, message = "authentication_required", "Authentication required."
			}
			if response.Code != expected {
				t.Fatalf("expected invitation cancel status %d, got %d with body %s", expected, response.Code, response.Body.String())
			}
			assertSafeError(t, response, code, message)
		})
	}
	cancelResponse := performRequest(server, http.MethodPatch, cancelPath, "Bearer dev:inventory-owner", nil)
	if cancelResponse.Code != http.StatusNoContent {
		t.Fatalf("expected owner invitation cancel status %d, got %d with body %s", http.StatusNoContent, cancelResponse.Code, cancelResponse.Body.String())
	}
	cancelledPreviewResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+cancelInvite.ID+"/preview", "Bearer dev:cancel-user:cancel@example.com", map[string]string{"acceptanceToken": cancelInviteToken})
	if cancelledPreviewResponse.Code != http.StatusOK {
		t.Fatalf("expected cancelled invite preview status %d, got %d with body %s", http.StatusOK, cancelledPreviewResponse.Code, cancelledPreviewResponse.Body.String())
	}
	cancelledPreviewJSON := cancelledPreviewResponse.Body.String()
	cancelledPreview := decodeInventoryAccessInvitationPreview(t, cancelledPreviewResponse).Data
	if cancelledPreview.Status != "cancelled" || cancelledPreview.InventoryID != inventoryID || cancelledPreview.IsExpired {
		t.Fatalf("expected safe cancelled preview, got %+v", cancelledPreview)
	}
	for _, forbidden := range []string{"cancel@example.com", cancelInviteToken, cancelInvite.InviteURL, "inviterPrincipalId", "acceptedPrincipalId", "tokenHash"} {
		if strings.Contains(cancelledPreviewJSON, forbidden) {
			t.Fatalf("cancelled preview response leaked %q: %s", forbidden, cancelledPreviewJSON)
		}
	}
	cancelledPreviewDidNotGrant := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:cancel-user", nil)
	if cancelledPreviewDidNotGrant.Code != http.StatusForbidden {
		t.Fatalf("expected cancelled preview not to grant access, got %d with body %s", cancelledPreviewDidNotGrant.Code, cancelledPreviewDidNotGrant.Body.String())
	}

	wrongTenantInvite := performRequest(server, http.MethodPost, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:inventory-owner", map[string]string{
		"email":        "other@example.com",
		"relationship": "viewer",
	})
	if wrongTenantInvite.Code != http.StatusNotFound {
		t.Fatalf("expected wrong tenant invite status %d, got %d with body %s", http.StatusNotFound, wrongTenantInvite.Code, wrongTenantInvite.Body.String())
	}
	assertSafeError(t, wrongTenantInvite, "resource_not_found", "Resource not found.")
}

func TestInventoryAccessInvitationRejectsMalformedAcceptanceTokens(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	for name, token := range map[string]string{
		"short":         strings.Repeat("A", 42),
		"overlong":      strings.Repeat("A", 44),
		"invalid chars": strings.Repeat("A", 42) + "!",
	} {
		t.Run(name, func(t *testing.T) {
			server := NewServer(":0", newSeededTestApp(t, seededState{
				tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "tenant-owner"}},
				inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"}},
				ids:         []string{"invite-viewer", "audit-invite-viewer", "audit-malformed-accept", "malformed-accept-event"},
			}))
			createdResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:inventory-owner", map[string]string{
				"email":        "viewer@example.com",
				"relationship": "viewer",
			})
			if createdResponse.Code != http.StatusCreated {
				t.Fatalf("expected invitation create status %d, got %d with body %s", http.StatusCreated, createdResponse.Code, createdResponse.Body.String())
			}
			invitation := decodeInventoryAccessInvitation(t, createdResponse).Data
			response := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{"acceptanceToken": token})
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected malformed-token acceptance status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "forbidden", "Forbidden.")
			assets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", nil)
			if assets.Code != http.StatusForbidden {
				t.Fatalf("expected malformed acceptance not to grant access, got %d with body %s", assets.Code, assets.Body.String())
			}
		})
	}
}

func TestInventoryAccessInvitationRejectsExpiredToken(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
		},
		ids:           []string{"invite-viewer", "audit-invite-viewer", "audit-expired-accept", "expired-accept-event"},
		invitationTTL: time.Nanosecond,
	}))

	invitationResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations", "Bearer dev:inventory-owner", map[string]string{
		"email":        "viewer@example.com",
		"relationship": "viewer",
	})
	if invitationResponse.Code != http.StatusCreated {
		t.Fatalf("expected invitation create status %d, got %d with body %s", http.StatusCreated, invitationResponse.Code, invitationResponse.Body.String())
	}
	invitation := decodeInventoryAccessInvitation(t, invitationResponse).Data
	invitationToken := acceptanceTokenFromInviteURL(t, invitation.InviteURL)
	time.Sleep(time.Millisecond)
	previewExpired := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/preview", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": invitationToken,
	})
	if previewExpired.Code != http.StatusOK {
		t.Fatalf("expected expired invite preview status %d, got %d with body %s", http.StatusOK, previewExpired.Code, previewExpired.Body.String())
	}
	expiredPreview := decodeInventoryAccessInvitationPreview(t, previewExpired).Data
	if expiredPreview.Status != "pending" || !expiredPreview.IsExpired {
		t.Fatalf("expected explicit expired preview, got %+v", expiredPreview)
	}

	acceptExpired := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-invitations/"+invitation.ID+"/accept", "Bearer dev:viewer-user:viewer@example.com", map[string]string{
		"acceptanceToken": invitationToken,
	})
	if acceptExpired.Code != http.StatusForbidden {
		t.Fatalf("expected expired invite accept status %d, got %d with body %s", http.StatusForbidden, acceptExpired.Code, acceptExpired.Body.String())
	}
	assertSafeError(t, acceptExpired, "forbidden", "Forbidden.")
}

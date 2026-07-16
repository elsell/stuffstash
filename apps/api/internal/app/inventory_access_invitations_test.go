package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestInventoryAccessInvitationCreateAcceptAndRevoke(t *testing.T) {
	outbox := &fakeOutbox{}
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
		outbox: outbox,
	}
	observer := &fakeObserver{}
	now := time.Date(2030, 6, 1, 12, 30, 0, 0, time.UTC)
	application := New(Dependencies{
		Observer:                  observer,
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    outbox,
		IDs:                       &fakeIDGenerator{ids: []string{"invite-one", "audit-invite", "audit-accept", "grant-event", "grant-claim", "invite-two", "audit-invite-two", "audit-revoke"}},
		Clock:                     fakeClock{now: now},
		InvitationPublicBaseURL:   "https://app.example.test/invitations/accept",
	})

	inviteResult, err := application.CreateInventoryAccessInvitation(context.Background(), CreateInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		Email:        "Viewer@Example.COM",
		Relationship: "viewer",
	})
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	invitation := inviteResult.Invitation
	if invitation.Email.String() != "viewer@example.com" || invitation.Status != ports.InventoryAccessInvitationPending {
		t.Fatalf("unexpected invitation: %+v", invitation)
	}
	inviteToken := acceptanceTokenFromAppInviteURL(t, inviteResult.InviteURL)
	if invitation.TokenHash != hashInventoryInvitationToken(inviteToken) {
		t.Fatalf("expected invitation response to include a matching one-time acceptance token")
	}
	if inviteResult.InviteURL != "https://app.example.test/invitations/accept?inventory=inventory-one&invitation=invite-one&tenant=tenant-one#token="+inviteToken {
		t.Fatalf("expected canonical invitation URL, got %q", inviteResult.InviteURL)
	}
	if !invitation.ExpiresAt.Equal(now.Add(7 * 24 * time.Hour)) {
		t.Fatalf("expected pending invitation expiration from injected clock, got %s", invitation.ExpiresAt)
	}
	if len(repository.auditRecords) != 1 || !repository.auditRecords[0].OccurredAt.Equal(now) {
		t.Fatalf("expected invitation audit timestamp from injected clock, got %+v", repository.auditRecords)
	}
	assertInvitationSecretsAbsent(t, "create", []any{repository.auditRecords, observer.events}, inviteToken, inviteResult.InviteURL)

	accepted, grant, err := application.AcceptInventoryAccessInvitation(context.Background(), AcceptInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: invitation.Email},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: invitation.ID,
		Token:        inviteToken,
	})
	if err != nil {
		t.Fatalf("accept invitation: %v", err)
	}
	if accepted.Status != ports.InventoryAccessInvitationAccepted || accepted.AcceptedPrincipalID != identity.PrincipalID("viewer-user") {
		t.Fatalf("expected accepted invitation, got %+v", accepted)
	}
	if grant.PrincipalID != identity.PrincipalID("viewer-user") || grant.Relationship != ports.InventoryAccessViewer {
		t.Fatalf("unexpected grant: %+v", grant)
	}
	if len(repository.auditRecords) != 2 {
		t.Fatalf("expected invite create/accept audits, got %+v", repository.auditRecords)
	}
	if len(outbox.processed) != 1 || outbox.processed[0] != "grant-event" {
		t.Fatalf("expected accepted invitation grant outbox processed, got %+v", outbox.processed)
	}
	if !observer.hasEvent(ports.EventInventoryInvitationCreated) || !observer.hasEvent(ports.EventInventoryInvitationAccepted) {
		t.Fatalf("expected invitation created and accepted observability events, got %+v", observer.events)
	}

	revokedInviteResult, err := application.CreateInventoryAccessInvitation(context.Background(), CreateInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		Email:        "editor@example.com",
		Relationship: "editor",
	})
	if err != nil {
		t.Fatalf("create invitation to revoke: %v", err)
	}
	revokedInvite := revokedInviteResult.Invitation
	revokedInviteToken := acceptanceTokenFromAppInviteURL(t, revokedInviteResult.InviteURL)
	revoked, err := application.RevokeInventoryAccessInvitation(context.Background(), RevokeInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: revokedInvite.ID,
	})
	if err != nil {
		t.Fatalf("revoke invitation: %v", err)
	}
	if !revoked {
		t.Fatalf("expected pending invitation revoked")
	}
	if !observer.hasEvent(ports.EventInventoryInvitationRevoked) {
		t.Fatalf("expected invitation revoked observability event, got %+v", observer.events)
	}
	_, _, err = application.AcceptInventoryAccessInvitation(context.Background(), AcceptInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("editor-user"), Email: identity.Email("editor@example.com")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: revokedInvite.ID,
		Token:        revokedInviteToken,
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected revoked invitation rejection, got %v", err)
	}
}

func TestCreateInventoryAccessInvitationRejectsDuplicatePendingInvitation(t *testing.T) {
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
		invitations: []ports.InventoryAccessInvitation{
			{
				ID:           "invite-one",
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				Email:        identity.Email("viewer@example.com"),
				TokenHash:    hashInventoryInvitationToken("old-token"),
				Relationship: ports.InventoryAccessViewer,
				Status:       ports.InventoryAccessInvitationPending,
				ExpiresAt:    time.Now().Add(time.Hour),
			},
		},
	}
	application := New(Dependencies{
		Observer:                  &fakeObserver{},
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    &fakeOutbox{},
		IDs:                       &fakeIDGenerator{ids: []string{"invite-two", "audit-invite-two"}},
		InvitationPublicBaseURL:   "https://app.example.test/invitations/accept",
	})

	_, err := application.CreateInventoryAccessInvitation(context.Background(), CreateInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		Email:        "viewer@example.com",
		Relationship: "viewer",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected duplicate pending invitation rejection, got %v", err)
	}
}

func TestAcceptInventoryAccessInvitationRejectsMissingOrWrongEmail(t *testing.T) {
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
		invitations: []ports.InventoryAccessInvitation{
			{
				ID:           "invite-one",
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				Email:        identity.Email("viewer@example.com"),
				TokenHash:    hashInventoryInvitationToken("correct-token"),
				Relationship: ports.InventoryAccessViewer,
				Status:       ports.InventoryAccessInvitationPending,
				ExpiresAt:    time.Now().Add(time.Hour),
			},
			{
				ID:           "expired-invite",
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				Email:        identity.Email("expired@example.com"),
				TokenHash:    hashInventoryInvitationToken("expired-token"),
				Relationship: ports.InventoryAccessViewer,
				Status:       ports.InventoryAccessInvitationPending,
				ExpiresAt:    time.Now().Add(-time.Hour),
			},
		},
		outbox: &fakeOutbox{},
	}
	application := New(Dependencies{
		Observer:                  &fakeObserver{},
		Authorizer:                &fakeAuthorizer{checkInventoryErr: ports.ErrForbidden},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    repository.outbox,
		IDs:                       &fakeIDGenerator{ids: []string{"audit-missing", "event-missing", "audit-wrong", "event-wrong"}},
	})

	for _, item := range []struct {
		name      string
		principal identity.Principal
		token     string
	}{
		{name: "missing email", principal: identity.Principal{ID: identity.PrincipalID("viewer-user")}, token: "correct-token"},
		{name: "wrong email", principal: identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: identity.Email("wrong@example.com")}, token: "correct-token"},
		{name: "missing token", principal: identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: identity.Email("viewer@example.com")}},
		{name: "wrong token", principal: identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: identity.Email("viewer@example.com")}, token: "wrong-token"},
	} {
		t.Run(item.name, func(t *testing.T) {
			_, _, err := application.AcceptInventoryAccessInvitation(context.Background(), AcceptInventoryAccessInvitationInput{
				Principal:    item.principal,
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				InvitationID: "invite-one",
				Token:        item.token,
			})
			if !errors.Is(err, ErrUnauthorized) {
				t.Fatalf("expected unauthorized, got %v", err)
			}
		})
	}

	_, _, err := application.AcceptInventoryAccessInvitation(context.Background(), AcceptInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("expired-user"), Email: identity.Email("expired@example.com")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: "expired-invite",
		Token:        "expired-token",
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected expired invitation rejection, got %v", err)
	}
}

func TestPreviewInventoryAccessInvitationValidatesScopeTokenAndInvitee(t *testing.T) {
	now := time.Date(2026, time.July, 14, 12, 0, 0, 0, time.UTC)
	observer := &fakeObserver{}
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Workshop tools"),
		},
		invitations: []ports.InventoryAccessInvitation{
			{
				ID:           "pending-invite",
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				Email:        identity.Email("viewer@example.com"),
				TokenHash:    hashInventoryInvitationToken("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
				Relationship: ports.InventoryAccessViewer,
				Status:       ports.InventoryAccessInvitationPending,
				ExpiresAt:    now.Add(time.Hour),
			},
			{
				ID:           "expired-invite",
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				Email:        identity.Email("viewer@example.com"),
				TokenHash:    hashInventoryInvitationToken("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
				Relationship: ports.InventoryAccessEditor,
				Status:       ports.InventoryAccessInvitationPending,
				ExpiresAt:    now.Add(-time.Minute),
			},
			{
				ID:                  "accepted-invite",
				TenantID:            tenant.ID("tenant-one"),
				InventoryID:         inventory.InventoryID("inventory-one"),
				Email:               identity.Email("viewer@example.com"),
				TokenHash:           hashInventoryInvitationToken("CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"),
				Relationship:        ports.InventoryAccessViewer,
				Status:              ports.InventoryAccessInvitationAccepted,
				AcceptedPrincipalID: identity.PrincipalID("viewer-user"),
				ExpiresAt:           now.Add(time.Hour),
			},
			{ID: "revoked-invite", TenantID: "tenant-one", InventoryID: "inventory-one", Email: "viewer@example.com", TokenHash: hashInventoryInvitationToken("EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE"), Relationship: ports.InventoryAccessViewer, Status: ports.InventoryAccessInvitationRevoked, ExpiresAt: now.Add(time.Hour)},
			{ID: "cancelled-invite", TenantID: "tenant-one", InventoryID: "inventory-one", Email: "viewer@example.com", TokenHash: hashInventoryInvitationToken("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"), Relationship: ports.InventoryAccessEditor, Status: ports.InventoryAccessInvitationCancelled, ExpiresAt: now.Add(time.Hour)},
			{ID: "other-accepted-invite", TenantID: "tenant-one", InventoryID: "inventory-one", Email: "viewer@example.com", TokenHash: hashInventoryInvitationToken("GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG"), Relationship: ports.InventoryAccessViewer, Status: ports.InventoryAccessInvitationAccepted, AcceptedPrincipalID: "other-user", ExpiresAt: now.Add(time.Hour)},
		},
	}
	application := New(Dependencies{
		Observer:                  observer,
		Authorizer:                &fakeAuthorizer{checkInventoryErr: ports.ErrForbidden},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    &fakeOutbox{},
		Clock:                     fakeClock{now: now},
	})

	preview, err := application.PreviewInventoryAccessInvitation(context.Background(), PreviewInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: identity.Email("Viewer@Example.COM")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: "pending-invite",
		Token:        "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	})
	if err != nil {
		t.Fatalf("preview invitation: %v", err)
	}
	if preview.InventoryID != inventory.InventoryID("inventory-one") || preview.InventoryName != "Workshop tools" || preview.Relationship != ports.InventoryAccessViewer || preview.Status != ports.InventoryAccessInvitationPending || preview.IsExpired {
		t.Fatalf("unexpected safe preview: %+v", preview)
	}

	expired, err := application.PreviewInventoryAccessInvitation(context.Background(), PreviewInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: identity.Email("viewer@example.com")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: "expired-invite",
		Token:        "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	})
	if err != nil || !expired.IsExpired {
		t.Fatalf("expected expired preview state, got preview=%+v err=%v", expired, err)
	}

	accepted, err := application.PreviewInventoryAccessInvitation(context.Background(), PreviewInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: identity.Email("viewer@example.com")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: "accepted-invite",
		Token:        "CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC",
	})
	if err != nil || accepted.Status != ports.InventoryAccessInvitationAccepted {
		t.Fatalf("expected accepted preview state, got preview=%+v err=%v", accepted, err)
	}

	for _, terminal := range []struct {
		id     string
		token  string
		status ports.InventoryAccessInvitationStatus
	}{
		{id: "revoked-invite", token: "EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE", status: ports.InventoryAccessInvitationRevoked},
		{id: "cancelled-invite", token: "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", status: ports.InventoryAccessInvitationCancelled},
	} {
		terminalPreview, err := application.PreviewInventoryAccessInvitation(context.Background(), PreviewInventoryAccessInvitationInput{
			Principal: identity.Principal{ID: "viewer-user", Email: "viewer@example.com"}, TenantID: "tenant-one", InventoryID: "inventory-one", InvitationID: terminal.id, Token: terminal.token,
		})
		if err != nil || terminalPreview.Status != terminal.status {
			t.Fatalf("expected %s preview, got preview=%+v err=%v", terminal.status, terminalPreview, err)
		}
	}

	for _, item := range []struct {
		name         string
		tenantID     tenant.ID
		inventoryID  inventory.InventoryID
		invitationID string
		token        string
	}{
		{name: "missing token", tenantID: "tenant-one", inventoryID: "inventory-one", invitationID: "pending-invite"},
		{name: "wrong token", tenantID: "tenant-one", inventoryID: "inventory-one", invitationID: "pending-invite", token: "DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD"},
		{name: "wrong tenant", tenantID: "tenant-two", inventoryID: "inventory-one", invitationID: "pending-invite", token: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		{name: "wrong inventory", tenantID: "tenant-one", inventoryID: "inventory-two", invitationID: "pending-invite", token: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		{name: "missing invitation", tenantID: "tenant-one", inventoryID: "inventory-one", invitationID: "missing-invite", token: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
	} {
		t.Run(item.name, func(t *testing.T) {
			_, err := application.PreviewInventoryAccessInvitation(context.Background(), PreviewInventoryAccessInvitationInput{
				Principal:    identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: identity.Email("viewer@example.com")},
				TenantID:     item.tenantID,
				InventoryID:  item.inventoryID,
				InvitationID: item.invitationID,
				Token:        item.token,
			})
			if !errors.Is(err, ErrInvitationInvalid) {
				t.Fatalf("expected generic invalid invitation error, got %v", err)
			}
		})
	}

	_, err = application.PreviewInventoryAccessInvitation(context.Background(), PreviewInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("wrong-user"), Email: identity.Email("wrong@example.com")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: "pending-invite",
		Token:        "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	})
	if !errors.Is(err, ErrInvitationEmailMismatch) {
		t.Fatalf("expected email mismatch only after valid token, got %v", err)
	}
	_, err = application.PreviewInventoryAccessInvitation(context.Background(), PreviewInventoryAccessInvitationInput{
		Principal: identity.Principal{ID: "viewer-user"}, TenantID: "tenant-one", InventoryID: "inventory-one", InvitationID: "pending-invite", Token: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	})
	if !errors.Is(err, ErrInvitationEmailMismatch) {
		t.Fatalf("expected missing verified email mismatch, got %v", err)
	}
	_, err = application.PreviewInventoryAccessInvitation(context.Background(), PreviewInventoryAccessInvitationInput{
		Principal: identity.Principal{ID: "viewer-user", Email: "viewer@example.com"}, TenantID: "tenant-one", InventoryID: "inventory-one", InvitationID: "other-accepted-invite", Token: "GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG",
	})
	if !errors.Is(err, ErrInvitationInvalid) {
		t.Fatalf("expected accepted invitation to remain bound to accepting principal, got %v", err)
	}
	if len(repository.auditRecords) != 0 || len(repository.accessGrants) != 0 {
		t.Fatalf("preview must not mutate invitation access state: %+v", repository)
	}
	assertInvitationSecretsAbsent(
		t,
		"preview",
		[]any{repository.auditRecords, observer.events},
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"https://app.example.test/invitations/accept?inventory=inventory-one&invitation=pending-invite&tenant=tenant-one#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	)
}

func TestBuildInventoryInvitationURLUsesCanonicalSafeShape(t *testing.T) {
	invitation := ports.InventoryAccessInvitation{
		ID:          "invite / one",
		TenantID:    tenant.ID("tenant / one"),
		InventoryID: inventory.InventoryID("inventory / one"),
	}
	link, err := buildInventoryInvitationURL("https://stash.example.test/ignored/path", invitation, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", false)
	if err != nil {
		t.Fatalf("build invitation URL: %v", err)
	}
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse invitation URL: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "stash.example.test" || parsed.Path != "/invitations/accept" {
		t.Fatalf("unexpected canonical invitation URL: %s", link)
	}
	if parsed.Query().Get("tenant") != "tenant / one" || parsed.Query().Get("inventory") != "inventory / one" || parsed.Query().Get("invitation") != "invite / one" {
		t.Fatalf("expected encoded scoped identifiers, got %s", link)
	}
	if parsed.Query().Has("token") || parsed.Fragment != "token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" {
		t.Fatalf("expected secret only in fragment, got %s", link)
	}

	for _, baseURL := range []string{
		"http://localhost:5173/invitations/accept",
		"http://127.0.0.1:5173/invitations/accept",
		"http://[::1]:5173/invitations/accept",
		"http://10.1.2.3:5173/invitations/accept",
		"http://172.16.0.1:5173/invitations/accept",
		"http://172.31.255.254:5173/invitations/accept",
		"http://192.168.1.117:5173/invitations/accept",
	} {
		if _, err := buildInventoryInvitationURL(baseURL, invitation, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", true); err != nil {
			t.Fatalf("expected local development URL %q to be accepted: %v", baseURL, err)
		}
	}
	if _, err := buildInventoryInvitationURL("http://localhost:5173/invitations/accept", invitation, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", false); err == nil {
		t.Fatal("expected loopback HTTP to require the explicit local-development switch")
	}

	for _, baseURL := range []string{
		"http://stash.example.test/invitations/accept",
		"http://8.8.8.8:5173/invitations/accept",
		"http://172.32.0.1:5173/invitations/accept",
		"//stash.example.test/invitations/accept",
		"/invitations/accept",
		"javascript:alert(1)",
		"file:///invitations/accept",
		"https://user:password@stash.example.test/invitations/accept",
		"https://stash.example.test/invitations/accept?leak=yes",
		"https://stash.example.test/invitations/accept#old-fragment",
		"https://stash.example.test/%zz",
		"https://stash.example.test\\evil.example/invitations/accept",
	} {
		t.Run(baseURL, func(t *testing.T) {
			if _, err := buildInventoryInvitationURL(baseURL, invitation, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", false); err == nil {
				t.Fatalf("expected unsafe base URL %q to be rejected", baseURL)
			}
		})
	}
}

func acceptanceTokenFromAppInviteURL(t *testing.T, inviteURL string) string {
	t.Helper()
	parsed, err := url.Parse(inviteURL)
	if err != nil {
		t.Fatalf("parse invitation URL: %v", err)
	}
	fragment, err := url.ParseQuery(parsed.Fragment)
	if err != nil {
		t.Fatalf("parse invitation URL fragment: %v", err)
	}
	token := fragment.Get("token")
	if token == "" {
		t.Fatalf("expected invitation token fragment in %q", inviteURL)
	}
	return token
}

func assertInvitationSecretsAbsent(t *testing.T, boundary string, captured []any, secrets ...string) {
	t.Helper()
	payload, err := json.Marshal(captured)
	if err != nil {
		t.Fatalf("marshal captured %s diagnostics: %v", boundary, err)
	}
	for _, secret := range secrets {
		if strings.Contains(string(payload), secret) {
			t.Fatalf("%s diagnostics leaked invitation secret %q: %s", boundary, secret, payload)
		}
	}
}

func TestCreateInventoryAccessInvitationRejectsInvalidPublicURLBeforePersistence(t *testing.T) {
	for _, publicURL := range []string{"", "http://public.example.test/invitations/accept"} {
		t.Run(publicURL, func(t *testing.T) {
			repository := &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}}
			application := New(Dependencies{
				Observer:                  &fakeObserver{},
				Authorizer:                &fakeAuthorizer{},
				Tenants:                   &fakeTenantRepository{exists: true},
				TenantUnitOfWork:          &fakeTenantRepository{exists: true},
				Inventories:               repository,
				InventoryUnitOfWork:       repository,
				InventoryAccess:           repository,
				InventoryAccessUnitOfWork: repository,
				Audit:                     &fakeAuditRepository{},
				Outbox:                    &fakeOutbox{},
				IDs:                       &fakeIDGenerator{ids: []string{"invite-one", "audit-one"}},
				InvitationPublicBaseURL:   publicURL,
			})

			_, err := application.CreateInventoryAccessInvitation(context.Background(), CreateInventoryAccessInvitationInput{
				Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				Email:        "viewer@example.com",
				Relationship: "viewer",
			})
			if err == nil {
				t.Fatal("expected invalid public URL rejection")
			}
			if len(repository.invitations) != 0 || len(repository.auditRecords) != 0 {
				t.Fatalf("invalid public URL must not persist invitation or audit state: %+v", repository)
			}
		})
	}
}

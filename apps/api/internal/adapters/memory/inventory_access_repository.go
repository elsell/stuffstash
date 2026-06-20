package memory

import (
	"context"
	"crypto/subtle"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
	"time"
)

func (s *Store) SaveInventoryAccessGrantAndEnqueue(_ context.Context, eventID string, grant ports.InventoryAccessGrant, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.inventories[grant.InventoryID]
	if !ok || item.TenantID.String() != grant.TenantID.String() {
		return ports.ErrForbidden
	}

	grantKey := inventoryAccessGrantStorageKey(grant)
	if _, exists := s.accessGrants[grantKey]; exists {
		return nil
	}
	eventKind, ok := grant.Relationship.GrantOutboxKind()
	if !ok {
		return fmt.Errorf("invalid inventory access relationship %q", grant.Relationship)
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.accessGrants[grantKey] = grant
	s.auditRecords[auditRecord.ID] = auditRecord
	s.outbox[eventID] = ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        eventKind,
		PrincipalID: grant.PrincipalID,
		TenantID:    grant.TenantID,
		InventoryID: grant.InventoryID,
		CreatedAt:   time.Now(),
	}
	return nil
}

func (s *Store) DeleteInventoryAccessGrantAndClaimRevoke(_ context.Context, eventID string, claimID string, leaseUntil time.Time, grant ports.InventoryAccessGrant, auditRecord audit.Record) (ports.AuthorizationOutboxEvent, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.inventories[grant.InventoryID]
	if !ok || item.TenantID.String() != grant.TenantID.String() {
		return ports.AuthorizationOutboxEvent{}, false, ports.ErrForbidden
	}

	grantKey := inventoryAccessGrantStorageKey(grant)
	eventKind, ok := grant.Relationship.RevokeOutboxKind()
	if !ok {
		return ports.AuthorizationOutboxEvent{}, false, fmt.Errorf("invalid inventory access relationship %q", grant.Relationship)
	}
	_, removed := s.accessGrants[grantKey]
	if removed {
		if _, exists := s.auditRecords[auditRecord.ID]; exists {
			return ports.AuthorizationOutboxEvent{}, false, ports.ErrConflict
		}
		delete(s.accessGrants, grantKey)
		s.auditRecords[auditRecord.ID] = auditRecord
	}
	if _, exists := s.outbox[eventID]; exists {
		return ports.AuthorizationOutboxEvent{}, false, ports.ErrConflict
	}
	event := ports.AuthorizationOutboxEvent{
		ID:           eventID,
		Kind:         eventKind,
		PrincipalID:  grant.PrincipalID,
		TenantID:     grant.TenantID,
		InventoryID:  grant.InventoryID,
		ClaimID:      claimID,
		ClaimedUntil: leaseUntil,
		CreatedAt:    time.Now(),
	}
	s.outbox[eventID] = event
	return event, removed, nil
}

func (s *Store) ListInventoryAccessGrants(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.InventoryAccessGrantPageRequest) ([]ports.InventoryAccessGrant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []ports.InventoryAccessGrant{}
	for _, grant := range s.accessGrants {
		key := grant.CursorKey()
		if grant.TenantID == tenantID && grant.InventoryID == inventoryID && key > page.AfterGrantKey {
			items = append(items, grant)
		}
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].CursorKey() < items[right].CursorKey()
	})
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

func (s *Store) SaveInventoryAccessInvitation(_ context.Context, invitation ports.InventoryAccessInvitation, auditRecord audit.Record) (ports.InventoryAccessInvitation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.inventories[invitation.InventoryID]
	if !ok || item.TenantID.String() != invitation.TenantID.String() {
		return ports.InventoryAccessInvitation{}, ports.ErrForbidden
	}
	for _, existing := range s.invitations {
		if existing.TenantID == invitation.TenantID && existing.InventoryID == invitation.InventoryID && existing.Email == invitation.Email && existing.Relationship == invitation.Relationship && existing.Status == ports.InventoryAccessInvitationPending {
			return ports.InventoryAccessInvitation{}, ports.ErrConflict
		}
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.InventoryAccessInvitation{}, ports.ErrConflict
	}
	invitation.Status = ports.InventoryAccessInvitationPending
	if invitation.CreatedAt.IsZero() {
		invitation.CreatedAt = time.Now()
	}
	if invitation.ExpiresAt.IsZero() {
		return ports.InventoryAccessInvitation{}, ports.ErrConflict
	}
	s.invitations[invitation.ID] = invitation
	s.auditRecords[auditRecord.ID] = auditRecord
	return invitation, nil
}

func (s *Store) AcceptInventoryAccessInvitationAndEnqueue(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, tokenHash string, acceptor identity.Principal, eventID string, auditRecord audit.Record) (ports.InventoryAccessInvitation, ports.InventoryAccessGrant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	invitation, ok := s.invitations[invitationID]
	if !ok || invitation.TenantID != tenantID || invitation.InventoryID != inventoryID || invitation.Status != ports.InventoryAccessInvitationPending || invitation.Email != acceptor.Email || !memoryInventoryInvitationTokenHashMatches(invitation.TokenHash, tokenHash) || invitation.ExpiresAt.IsZero() || !invitation.ExpiresAt.After(time.Now()) {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, ports.ErrConflict
	}
	eventKind, ok := invitation.Relationship.GrantOutboxKind()
	if !ok {
		return ports.InventoryAccessInvitation{}, ports.InventoryAccessGrant{}, fmt.Errorf("invalid inventory access relationship %q", invitation.Relationship)
	}
	grant := ports.InventoryAccessGrant{
		TenantID:     invitation.TenantID,
		InventoryID:  invitation.InventoryID,
		PrincipalID:  acceptor.ID,
		Relationship: invitation.Relationship,
	}
	grantKey := inventoryAccessGrantStorageKey(grant)
	_, grantExists := s.accessGrants[grantKey]
	s.accessGrants[grantKey] = grant
	invitation.Status = ports.InventoryAccessInvitationAccepted
	invitation.AcceptedPrincipalID = acceptor.ID
	invitation.AcceptedAt = time.Now()
	s.invitations[invitation.ID] = invitation
	s.auditRecords[auditRecord.ID] = auditRecord
	if !grantExists {
		s.outbox[eventID] = ports.AuthorizationOutboxEvent{
			ID:          eventID,
			Kind:        eventKind,
			PrincipalID: acceptor.ID,
			TenantID:    invitation.TenantID,
			InventoryID: invitation.InventoryID,
			CreatedAt:   time.Now(),
		}
	}
	return invitation, grant, nil
}

func memoryInventoryInvitationTokenHashMatches(storedHash string, providedHash string) bool {
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(providedHash)) == 1
}

func (s *Store) RevokeInventoryAccessInvitation(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	invitation, ok := s.invitations[invitationID]
	if !ok || invitation.TenantID != tenantID || invitation.InventoryID != inventoryID || invitation.Status != ports.InventoryAccessInvitationPending {
		return false, nil
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return false, ports.ErrConflict
	}
	invitation.Status = ports.InventoryAccessInvitationRevoked
	invitation.RevokedAt = time.Now()
	s.invitations[invitation.ID] = invitation
	s.auditRecords[auditRecord.ID] = auditRecord
	return true, nil
}

func inventoryAccessGrantStorageKey(grant ports.InventoryAccessGrant) string {
	return grant.TenantID.String() + ":" + grant.InventoryID.String() + ":" + grant.CursorKey()
}

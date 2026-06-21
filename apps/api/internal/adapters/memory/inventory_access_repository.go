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

func (s *Store) InventoryAccessGrantByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, principalID identity.PrincipalID, relationship ports.InventoryAccessRelationship) (ports.InventoryAccessGrant, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	grant := ports.InventoryAccessGrant{
		TenantID:     tenantID,
		InventoryID:  inventoryID,
		PrincipalID:  principalID,
		Relationship: relationship,
	}
	stored, ok := s.accessGrants[inventoryAccessGrantStorageKey(grant)]
	return stored, ok, nil
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

func (s *Store) AcceptInventoryAccessInvitationAndEnqueue(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, tokenHash string, acceptor identity.Principal, eventID string, now time.Time, auditRecord audit.Record) (ports.InventoryAccessInvitation, ports.InventoryAccessGrant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	invitation, ok := s.invitations[invitationID]
	if !ok || invitation.TenantID != tenantID || invitation.InventoryID != inventoryID || invitation.Status != ports.InventoryAccessInvitationPending || invitation.Email != acceptor.Email || !memoryInventoryInvitationTokenHashMatches(invitation.TokenHash, tokenHash) || invitation.ExpiresAt.IsZero() || !invitation.ExpiresAt.After(now) {
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

func (s *Store) InventoryAccessInvitationByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string) (ports.InventoryAccessInvitation, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	invitation, ok := s.invitations[invitationID]
	if !ok || invitation.TenantID != tenantID || invitation.InventoryID != inventoryID {
		return ports.InventoryAccessInvitation{}, false, nil
	}
	return invitation, true, nil
}

func (s *Store) ListInventoryAccessInvitations(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.InventoryAccessInvitationPageRequest) ([]ports.InventoryAccessInvitation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := page.Now
	if now.IsZero() {
		now = time.Now()
	}
	items := []ports.InventoryAccessInvitation{}
	for _, invitation := range s.invitations {
		key := invitation.CursorKey()
		if invitation.TenantID != tenantID || invitation.InventoryID != inventoryID || key <= page.AfterInvitationID {
			continue
		}
		if !memoryInvitationMatchesStatusFilter(invitation, page.StatusFilter, now) {
			continue
		}
		items = append(items, invitation)
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].CursorKey() < items[right].CursorKey()
	})
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

func memoryInvitationMatchesStatusFilter(invitation ports.InventoryAccessInvitation, filter ports.InventoryAccessInvitationStatusFilter, now time.Time) bool {
	switch filter {
	case "", ports.InventoryAccessInvitationStatusFilterAll:
		return true
	case ports.InventoryAccessInvitationStatusFilterPending:
		return invitation.Status == ports.InventoryAccessInvitationPending && !invitation.IsExpired(now)
	case ports.InventoryAccessInvitationStatusFilterExpired:
		return invitation.IsExpired(now)
	case ports.InventoryAccessInvitationStatusFilterAccepted:
		return invitation.Status == ports.InventoryAccessInvitationAccepted
	case ports.InventoryAccessInvitationStatusFilterRevoked:
		return invitation.Status == ports.InventoryAccessInvitationRevoked
	case ports.InventoryAccessInvitationStatusFilterCancelled:
		return invitation.Status == ports.InventoryAccessInvitationCancelled
	default:
		return false
	}
}

func memoryInventoryInvitationTokenHashMatches(storedHash string, providedHash string) bool {
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(providedHash)) == 1
}

func (s *Store) UpdateInventoryAccessInvitationExpiration(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, expiresAt time.Time, auditRecord audit.Record) (ports.InventoryAccessInvitation, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	invitation, ok := s.invitations[invitationID]
	if !ok || invitation.TenantID != tenantID || invitation.InventoryID != inventoryID {
		return ports.InventoryAccessInvitation{}, false, nil
	}
	if invitation.Status != ports.InventoryAccessInvitationPending {
		return ports.InventoryAccessInvitation{}, false, ports.ErrConflict
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.InventoryAccessInvitation{}, false, ports.ErrConflict
	}
	invitation.ExpiresAt = expiresAt
	s.invitations[invitation.ID] = invitation
	s.auditRecords[auditRecord.ID] = auditRecord
	return invitation, true, nil
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

func (s *Store) CancelInventoryAccessInvitation(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	invitation, ok := s.invitations[invitationID]
	if !ok || invitation.TenantID != tenantID || invitation.InventoryID != inventoryID || invitation.Status != ports.InventoryAccessInvitationPending {
		return false, nil
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return false, ports.ErrConflict
	}
	invitation.Status = ports.InventoryAccessInvitationCancelled
	invitation.RevokedAt = time.Now()
	s.invitations[invitation.ID] = invitation
	s.auditRecords[auditRecord.ID] = auditRecord
	return true, nil
}

func (s *Store) DeleteInventoryAccessInvitation(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	invitation, ok := s.invitations[invitationID]
	if !ok || invitation.TenantID != tenantID || invitation.InventoryID != inventoryID {
		return false, nil
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return false, ports.ErrConflict
	}
	delete(s.invitations, invitationID)
	s.auditRecords[auditRecord.ID] = auditRecord
	return true, nil
}

func inventoryAccessGrantStorageKey(grant ports.InventoryAccessGrant) string {
	return grant.TenantID.String() + ":" + grant.InventoryID.String() + ":" + grant.CursorKey()
}

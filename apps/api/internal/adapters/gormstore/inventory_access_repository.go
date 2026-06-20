package gormstore

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) SaveInventoryAccessGrantAndEnqueue(ctx context.Context, eventID string, grant ports.InventoryAccessGrant, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var containingInventory inventoryModel
		err := tx.Where(&inventoryModel{
			ID:       grant.InventoryID.String(),
			TenantID: grant.TenantID.String(),
		}).First(&containingInventory).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}

		grantKey := grant.CursorKey()
		var existingGrant inventoryAccessGrantModel
		err = tx.Where(&inventoryAccessGrantModel{
			TenantID:    grant.TenantID.String(),
			InventoryID: grant.InventoryID.String(),
			GrantKey:    grantKey,
		}).First(&existingGrant).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := tx.Save(&inventoryAccessGrantModel{
			TenantID:     grant.TenantID.String(),
			InventoryID:  grant.InventoryID.String(),
			GrantKey:     grantKey,
			PrincipalID:  grant.PrincipalID.String(),
			Relationship: string(grant.Relationship),
		}).Error; err != nil {
			return err
		}

		eventKind, ok := grant.Relationship.GrantOutboxKind()
		if !ok {
			return fmt.Errorf("invalid inventory access relationship %q", grant.Relationship)
		}
		outboxInventoryID := grant.InventoryID.String()
		if err := tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(eventKind),
			PrincipalID: grant.PrincipalID.String(),
			TenantID:    grant.TenantID.String(),
			InventoryID: &outboxInventoryID,
		}).Error; err != nil {
			return err
		}

		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) DeleteInventoryAccessGrantAndClaimRevoke(ctx context.Context, eventID string, claimID string, leaseUntil time.Time, grant ports.InventoryAccessGrant, auditRecord audit.Record) (ports.AuthorizationOutboxEvent, bool, error) {
	removed := false
	var outboxEvent ports.AuthorizationOutboxEvent
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var containingInventory inventoryModel
		err := tx.Where(&inventoryModel{
			ID:       grant.InventoryID.String(),
			TenantID: grant.TenantID.String(),
		}).First(&containingInventory).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}

		eventKind, ok := grant.Relationship.RevokeOutboxKind()
		if !ok {
			return fmt.Errorf("invalid inventory access relationship %q", grant.Relationship)
		}

		grantKey := grant.CursorKey()
		result := tx.Where(&inventoryAccessGrantModel{
			TenantID:    grant.TenantID.String(),
			InventoryID: grant.InventoryID.String(),
			GrantKey:    grantKey,
		}).Delete(&inventoryAccessGrantModel{})
		if result.Error != nil {
			return result.Error
		}
		removed = result.RowsAffected > 0

		inventoryID := grant.InventoryID.String()
		model := authorizationOutboxEventModel{
			ID:           eventID,
			Kind:         string(eventKind),
			PrincipalID:  grant.PrincipalID.String(),
			TenantID:     grant.TenantID.String(),
			InventoryID:  &inventoryID,
			ClaimID:      claimID,
			ClaimedUntil: &leaseUntil,
		}
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		outboxEvent = model.toPort()

		if !removed {
			return nil
		}
		return createAuditRecord(tx, auditRecord)
	})
	return outboxEvent, removed, err
}

func (s Store) ListInventoryAccessGrants(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.InventoryAccessGrantPageRequest) ([]ports.InventoryAccessGrant, error) {
	var models []inventoryAccessGrantModel
	query := s.db.WithContext(ctx).Where(&inventoryAccessGrantModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	})
	if page.AfterGrantKey != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "grant_key"}, Value: page.AfterGrantKey})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "grant_key"}}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]ports.InventoryAccessGrant, 0, len(models))
	for _, model := range models {
		item, ok := model.toPort()
		if !ok {
			return nil, fmt.Errorf("invalid inventory access grant row %q", model.GrantKey)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s Store) InventoryAccessGrantByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, principalID identity.PrincipalID, relationship ports.InventoryAccessRelationship) (ports.InventoryAccessGrant, bool, error) {
	grant := ports.InventoryAccessGrant{
		TenantID:     tenantID,
		InventoryID:  inventoryID,
		PrincipalID:  principalID,
		Relationship: relationship,
	}
	var model inventoryAccessGrantModel
	err := s.db.WithContext(ctx).Where(&inventoryAccessGrantModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
		GrantKey:    grant.CursorKey(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.InventoryAccessGrant{}, false, nil
	}
	if err != nil {
		return ports.InventoryAccessGrant{}, false, err
	}
	item, ok := model.toPort()
	if !ok {
		return ports.InventoryAccessGrant{}, false, fmt.Errorf("invalid inventory access grant row %q", model.GrantKey)
	}
	return item, true, nil
}

func (s Store) SaveInventoryAccessInvitation(ctx context.Context, invitation ports.InventoryAccessInvitation, auditRecord audit.Record) (ports.InventoryAccessInvitation, error) {
	var saved ports.InventoryAccessInvitation
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var containingInventory inventoryModel
		err := tx.Where(&inventoryModel{
			ID:       invitation.InventoryID.String(),
			TenantID: invitation.TenantID.String(),
		}).First(&containingInventory).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}

		var existing inventoryAccessInvitationModel
		err = tx.Where(&inventoryAccessInvitationModel{
			TenantID:     invitation.TenantID.String(),
			InventoryID:  invitation.InventoryID.String(),
			Email:        invitation.Email.String(),
			Relationship: string(invitation.Relationship),
			Status:       string(ports.InventoryAccessInvitationPending),
		}).First(&existing).Error
		if err == nil {
			return ports.ErrConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		invitation.Status = ports.InventoryAccessInvitationPending
		if invitation.CreatedAt.IsZero() {
			invitation.CreatedAt = time.Now()
		}
		if invitation.ExpiresAt.IsZero() {
			return ports.ErrConflict
		}
		model := inventoryAccessInvitationModelFromPort(invitation)
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		var ok bool
		saved, ok = model.toPort()
		if !ok {
			return fmt.Errorf("invalid inventory access invitation row %q", model.ID)
		}
		return nil
	})
	return saved, err
}

func (s Store) InventoryAccessInvitationByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string) (ports.InventoryAccessInvitation, bool, error) {
	var model inventoryAccessInvitationModel
	err := s.db.WithContext(ctx).Where(&inventoryAccessInvitationModel{
		ID:          invitationID,
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.InventoryAccessInvitation{}, false, nil
	}
	if err != nil {
		return ports.InventoryAccessInvitation{}, false, err
	}
	invitation, ok := model.toPort()
	if !ok {
		return ports.InventoryAccessInvitation{}, false, fmt.Errorf("invalid inventory access invitation row %q", model.ID)
	}
	return invitation, true, nil
}

func (s Store) ListInventoryAccessInvitations(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.InventoryAccessInvitationPageRequest) ([]ports.InventoryAccessInvitation, error) {
	var models []inventoryAccessInvitationModel
	query := s.db.WithContext(ctx).Where(&inventoryAccessInvitationModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	})
	if page.AfterInvitationID != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterInvitationID})
	}
	now := page.Now
	if now.IsZero() {
		now = time.Now()
	}
	switch page.StatusFilter {
	case "", ports.InventoryAccessInvitationStatusFilterAll:
	case ports.InventoryAccessInvitationStatusFilterPending:
		query = query.Where(&inventoryAccessInvitationModel{Status: string(ports.InventoryAccessInvitationPending)}).
			Where(clause.Gt{Column: clause.Column{Name: "expires_at"}, Value: now})
	case ports.InventoryAccessInvitationStatusFilterExpired:
		query = query.Where(&inventoryAccessInvitationModel{Status: string(ports.InventoryAccessInvitationPending)}).
			Where(clause.Lte{Column: clause.Column{Name: "expires_at"}, Value: now})
	case ports.InventoryAccessInvitationStatusFilterAccepted:
		query = query.Where(&inventoryAccessInvitationModel{Status: string(ports.InventoryAccessInvitationAccepted)})
	case ports.InventoryAccessInvitationStatusFilterRevoked:
		query = query.Where(&inventoryAccessInvitationModel{Status: string(ports.InventoryAccessInvitationRevoked)})
	case ports.InventoryAccessInvitationStatusFilterCancelled:
		query = query.Where(&inventoryAccessInvitationModel{Status: string(ports.InventoryAccessInvitationCancelled)})
	default:
		return nil, fmt.Errorf("invalid inventory invitation status filter %q", page.StatusFilter)
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]ports.InventoryAccessInvitation, 0, len(models))
	for _, model := range models {
		item, ok := model.toPort()
		if !ok {
			return nil, fmt.Errorf("invalid inventory access invitation row %q", model.ID)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s Store) AcceptInventoryAccessInvitationAndEnqueue(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, tokenHash string, acceptor identity.Principal, eventID string, auditRecord audit.Record) (ports.InventoryAccessInvitation, ports.InventoryAccessGrant, error) {
	var saved ports.InventoryAccessInvitation
	var grant ports.InventoryAccessGrant
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model inventoryAccessInvitationModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&inventoryAccessInvitationModel{ID: invitationID, TenantID: tenantID.String(), InventoryID: inventoryID.String()}).First(&model).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		invitation, ok := model.toPort()
		if !ok {
			return fmt.Errorf("invalid inventory access invitation row %q", model.ID)
		}
		if invitation.Status != ports.InventoryAccessInvitationPending || invitation.Email != acceptor.Email || !inventoryInvitationTokenHashMatches(invitation.TokenHash, tokenHash) || invitation.ExpiresAt.IsZero() || !invitation.ExpiresAt.After(time.Now()) {
			return ports.ErrForbidden
		}
		eventKind, ok := invitation.Relationship.GrantOutboxKind()
		if !ok {
			return fmt.Errorf("invalid inventory access relationship %q", invitation.Relationship)
		}

		grant = ports.InventoryAccessGrant{
			TenantID:     invitation.TenantID,
			InventoryID:  invitation.InventoryID,
			PrincipalID:  acceptor.ID,
			Relationship: invitation.Relationship,
		}
		grantModel := inventoryAccessGrantModel{
			TenantID:     grant.TenantID.String(),
			InventoryID:  grant.InventoryID.String(),
			GrantKey:     grant.CursorKey(),
			PrincipalID:  grant.PrincipalID.String(),
			Relationship: string(grant.Relationship),
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&grantModel)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			outboxInventoryID := grant.InventoryID.String()
			if err := tx.Create(&authorizationOutboxEventModel{
				ID:          eventID,
				Kind:        string(eventKind),
				PrincipalID: grant.PrincipalID.String(),
				TenantID:    grant.TenantID.String(),
				InventoryID: &outboxInventoryID,
			}).Error; err != nil {
				return err
			}
		}

		now := time.Now()
		model.Status = string(ports.InventoryAccessInvitationAccepted)
		model.AcceptedPrincipalID = acceptor.ID.String()
		model.AcceptedAt = &now
		if err := tx.Save(&model).Error; err != nil {
			return err
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}

		saved, ok = model.toPort()
		if !ok {
			return fmt.Errorf("invalid inventory access invitation row %q", model.ID)
		}
		return nil
	})
	return saved, grant, err
}

func inventoryInvitationTokenHashMatches(storedHash string, providedHash string) bool {
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(providedHash)) == 1
}

func (s Store) UpdateInventoryAccessInvitationExpiration(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, expiresAt time.Time, auditRecord audit.Record) (ports.InventoryAccessInvitation, bool, error) {
	var saved ports.InventoryAccessInvitation
	updated := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model inventoryAccessInvitationModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&inventoryAccessInvitationModel{
			ID:          invitationID,
			TenantID:    tenantID.String(),
			InventoryID: inventoryID.String(),
		}).First(&model).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if model.Status != string(ports.InventoryAccessInvitationPending) {
			return ports.ErrConflict
		}
		model.ExpiresAt = expiresAt
		if err := tx.Save(&model).Error; err != nil {
			return err
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		var ok bool
		saved, ok = model.toPort()
		if !ok {
			return fmt.Errorf("invalid inventory access invitation row %q", model.ID)
		}
		updated = true
		return nil
	})
	return saved, updated, err
}

func (s Store) RevokeInventoryAccessInvitation(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error) {
	revoked := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model inventoryAccessInvitationModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&inventoryAccessInvitationModel{
			ID:          invitationID,
			TenantID:    tenantID.String(),
			InventoryID: inventoryID.String(),
		}).First(&model).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if model.Status != string(ports.InventoryAccessInvitationPending) {
			return nil
		}
		now := time.Now()
		model.Status = string(ports.InventoryAccessInvitationRevoked)
		model.RevokedAt = &now
		if err := tx.Save(&model).Error; err != nil {
			return err
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		revoked = true
		return nil
	})
	return revoked, err
}

func (s Store) CancelInventoryAccessInvitation(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error) {
	cancelled := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model inventoryAccessInvitationModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&inventoryAccessInvitationModel{
			ID:          invitationID,
			TenantID:    tenantID.String(),
			InventoryID: inventoryID.String(),
		}).First(&model).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if model.Status != string(ports.InventoryAccessInvitationPending) {
			return nil
		}
		now := time.Now()
		model.Status = string(ports.InventoryAccessInvitationCancelled)
		model.RevokedAt = &now
		if err := tx.Save(&model).Error; err != nil {
			return err
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		cancelled = true
		return nil
	})
	return cancelled, err
}

func (s Store) DeleteInventoryAccessInvitation(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, invitationID string, auditRecord audit.Record) (bool, error) {
	deleted := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where(&inventoryAccessInvitationModel{
			ID:          invitationID,
			TenantID:    tenantID.String(),
			InventoryID: inventoryID.String(),
		}).Delete(&inventoryAccessInvitationModel{})
		if result.Error != nil {
			return result.Error
		}
		deleted = result.RowsAffected > 0
		if !deleted {
			return nil
		}
		return createAuditRecord(tx, auditRecord)
	})
	return deleted, err
}

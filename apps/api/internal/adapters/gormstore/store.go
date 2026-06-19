package gormstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) Store {
	return Store{db: db}
}

func Migrate(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&tenantModel{}, &inventoryModel{}, &inventoryAccessGrantModel{}, &customAssetTypeModel{}, &customFieldDefinitionModel{}, &customFieldDefinitionAssetTypeModel{}, &assetModel{}, &auditRecordModel{}, &authorizationOutboxEventModel{})
}

func (s Store) SaveTenant(ctx context.Context, item tenant.Tenant) error {
	model := tenantModel{
		ID:   item.ID.String(),
		Name: item.Name.String(),
	}

	return s.db.WithContext(ctx).Save(&model).Error
}

func (s Store) TenantExists(ctx context.Context, tenantID tenant.ID) (bool, error) {
	var model tenantModel
	err := s.db.WithContext(ctx).Where(&tenantModel{ID: tenantID.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s Store) SaveInventory(ctx context.Context, item inventory.Inventory) error {
	model := inventoryModel{
		ID:       item.ID.String(),
		TenantID: item.TenantID.String(),
		Name:     item.Name.String(),
	}

	return s.db.WithContext(ctx).Save(&model).Error
}

func (s Store) InventoryByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error) {
	var model inventoryModel
	err := s.db.WithContext(ctx).Where(&inventoryModel{
		ID:       inventoryID.String(),
		TenantID: tenantID.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return inventory.Inventory{}, false, nil
	}
	if err != nil {
		return inventory.Inventory{}, false, err
	}
	item, ok := model.toDomain()
	return item, ok, nil
}

func (s Store) SaveTenantAndEnqueueOwnerGrant(ctx context.Context, eventID string, item tenant.Tenant, principal identity.Principal, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&tenantModel{
			ID:   item.ID.String(),
			Name: item.Name.String(),
		}).Error; err != nil {
			return err
		}

		if err := tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(ports.AuthorizationOutboxGrantTenantOwner),
			PrincipalID: principal.ID.String(),
			TenantID:    item.ID.String(),
		}).Error; err != nil {
			return err
		}

		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) SaveInventoryAndEnqueueOwnerGrant(ctx context.Context, eventID string, item inventory.Inventory, tenantID tenant.ID, principal identity.Principal, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&inventoryModel{
			ID:       item.ID.String(),
			TenantID: item.TenantID.String(),
			Name:     item.Name.String(),
		}).Error; err != nil {
			return err
		}

		inventoryID := item.ID.String()
		if err := tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(ports.AuthorizationOutboxGrantInventoryOwner),
			PrincipalID: principal.ID.String(),
			TenantID:    tenantID.String(),
			InventoryID: &inventoryID,
		}).Error; err != nil {
			return err
		}

		return createAuditRecord(tx, auditRecord)
	})
}

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

		inventoryID := grant.InventoryID.String()
		if err := tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(outboxKindForInventoryAccess(grant.Relationship)),
			PrincipalID: grant.PrincipalID.String(),
			TenantID:    grant.TenantID.String(),
			InventoryID: &inventoryID,
		}).Error; err != nil {
			return err
		}

		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) ClaimPendingAuthorizationOutboxEvents(ctx context.Context, claimID string, limit int, leaseUntil time.Time) ([]ports.AuthorizationOutboxEvent, error) {
	if limit <= 0 {
		limit = 25
	}

	events := []ports.AuthorizationOutboxEvent{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var models []authorizationOutboxEventModel
		now := time.Now()
		if err := tx.
			Clauses(skipLockedForUpdate()).
			Where(claimableAuthorizationOutboxEvent(now)).
			Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).
			Limit(limit).
			Find(&models).Error; err != nil {
			return err
		}

		claimed := make([]authorizationOutboxEventModel, 0, len(models))
		for _, model := range models {
			model.ClaimID = claimID
			model.ClaimedUntil = &leaseUntil
			claimed = append(claimed, model)
		}

		for _, model := range claimed {
			if err := tx.
				Model(&authorizationOutboxEventModel{}).
				Where(&authorizationOutboxEventModel{ID: model.ID}).
				Updates(map[string]any{
					"claim_id":      model.ClaimID,
					"claimed_until": model.ClaimedUntil,
				}).Error; err != nil {
				return err
			}
			events = append(events, model.toPort())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}

func skipLockedForUpdate() clause.Locking {
	return clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}
}

func claimableAuthorizationOutboxEvent(now time.Time) clause.Expression {
	return clause.And(
		clause.Eq{Column: clause.Column{Name: "processed_at"}, Value: nil},
		clause.Eq{Column: clause.Column{Name: "dead_lettered_at"}, Value: nil},
		clause.Or(
			clause.Eq{Column: clause.Column{Name: "claim_id"}, Value: ""},
			clause.Lte{Column: clause.Column{Name: "claimed_until"}, Value: now},
		),
	)
}

func (s Store) MarkAuthorizationOutboxEventProcessed(ctx context.Context, eventID string, claimID string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&authorizationOutboxEventModel{}).
		Where(&authorizationOutboxEventModel{ID: eventID, ClaimID: claimID}).
		Updates(map[string]any{
			"processed_at":  now,
			"last_error":    "",
			"claim_id":      "",
			"claimed_until": nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	return nil
}

func (s Store) MarkAuthorizationOutboxEventFailed(ctx context.Context, eventID string, claimID string, reason string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model authorizationOutboxEventModel
		if err := tx.Where(&authorizationOutboxEventModel{ID: eventID, ClaimID: claimID}).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrAuthorizationOutboxClaimLost
			}
			return err
		}

		model.Attempts++
		model.LastError = reason
		model.ClaimID = ""
		model.ClaimedUntil = nil
		return tx.Save(&model).Error
	})
}

func (s Store) MarkAuthorizationOutboxEventDeadLettered(ctx context.Context, eventID string, claimID string, reason string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&authorizationOutboxEventModel{}).
		Where(&authorizationOutboxEventModel{ID: eventID, ClaimID: claimID}).
		Updates(map[string]any{
			"dead_lettered_at":   now,
			"dead_letter_reason": reason,
			"claim_id":           "",
			"claimed_until":      nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	return nil
}

func (s Store) ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID, page ports.InventoryListPageRequest) ([]inventory.Inventory, error) {
	var models []inventoryModel
	query := s.db.WithContext(ctx).Where(&inventoryModel{TenantID: tenantID.String()})
	if page.AfterInventoryID.String() != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterInventoryID.String()})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]inventory.Inventory, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid inventory row %q", model.ID)
		}
		items = append(items, item)
	}

	return items, nil
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

func (s Store) SaveCustomFieldDefinition(ctx context.Context, definition customfield.Definition, auditRecord audit.Record) error {
	enumOptions, err := json.Marshal(customFieldKeysToStrings(definition.EnumOptions))
	if err != nil {
		return err
	}
	model := customFieldDefinitionModel{
		ID:            definition.ID.String(),
		TenantID:      definition.TenantID.String(),
		Scope:         definition.Scope.String(),
		FieldKey:      definition.Key.String(),
		DisplayName:   definition.DisplayName.String(),
		FieldType:     definition.Type.String(),
		EnumOptions:   string(enumOptions),
		Applicability: definition.Applicability.String(),
	}
	if definition.InventoryID.String() != "" {
		inventoryID := definition.InventoryID.String()
		model.InventoryID = &inventoryID
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing customFieldDefinitionModel
		query := tx.Where(&customFieldDefinitionModel{
			TenantID: definition.TenantID.String(),
			FieldKey: definition.Key.String(),
		})
		if definition.Scope == customfield.ScopeInventory {
			query = query.Where(clause.Or(
				clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
				clause.Eq{Column: "inventory_id", Value: definition.InventoryID.String()},
			))
		}
		err := query.First(&existing).Error
		if err == nil {
			return ports.ErrConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Save(&model).Error; err != nil {
			return customFieldDefinitionWriteError(err)
		}
		for _, targetID := range definition.CustomAssetTypeIDs {
			if err := tx.Create(&customFieldDefinitionAssetTypeModel{
				CustomFieldDefinitionID: definition.ID.String(),
				CustomAssetTypeID:       targetID.String(),
				TenantID:                definition.TenantID.String(),
				InventoryID:             model.InventoryID,
			}).Error; err != nil {
				return customFieldDefinitionWriteError(err)
			}
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) UpdateCustomFieldDefinition(ctx context.Context, definition customfield.Definition, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing customFieldDefinitionModel
		err := tx.Where(&customFieldDefinitionModel{
			ID:       definition.ID.String(),
			TenantID: definition.TenantID.String(),
		}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.Scope != definition.Scope.String() || existing.FieldKey != definition.Key.String() || existing.FieldType != definition.Type.String() || existing.Applicability != definition.Applicability.String() || stringFromPtr(existing.InventoryID) != definition.InventoryID.String() {
			return ports.ErrForbidden
		}
		var rawOptions []string
		if err := json.Unmarshal([]byte(existing.EnumOptions), &rawOptions); err != nil {
			return err
		}
		if !slices.Equal(rawOptions, customFieldKeysToStrings(definition.EnumOptions)) {
			return ports.ErrForbidden
		}
		targetsByDefinitionID, err := customFieldDefinitionTargets(ctx, tx, []customFieldDefinitionModel{existing})
		if err != nil {
			return err
		}
		if !slices.Equal(targetsByDefinitionID[definition.ID.String()], definition.CustomAssetTypeIDs) {
			return ports.ErrForbidden
		}

		if err := tx.Model(&existing).Updates(map[string]any{
			"display_name": definition.DisplayName.String(),
		}).Error; err != nil {
			return customFieldDefinitionWriteError(err)
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) SaveCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	model := customAssetTypeModel{
		ID:          assetType.ID.String(),
		TenantID:    assetType.TenantID.String(),
		Scope:       assetType.Scope.String(),
		TypeKey:     assetType.Key.String(),
		DisplayName: assetType.DisplayName.String(),
		Description: assetType.Description.String(),
	}
	if assetType.InventoryID.String() != "" {
		inventoryID := assetType.InventoryID.String()
		model.InventoryID = &inventoryID
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing customAssetTypeModel
		query := tx.Where(&customAssetTypeModel{
			TenantID: assetType.TenantID.String(),
			TypeKey:  assetType.Key.String(),
		})
		if assetType.Scope == customfield.ScopeInventory {
			query = query.Where(clause.Or(
				clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
				clause.Eq{Column: "inventory_id", Value: assetType.InventoryID.String()},
			))
		}
		err := query.First(&existing).Error
		if err == nil {
			return ports.ErrConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Save(&model).Error; err != nil {
			return customFieldDefinitionWriteError(err)
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) UpdateCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing customAssetTypeModel
		err := tx.Where(&customAssetTypeModel{
			ID:       assetType.ID.String(),
			TenantID: assetType.TenantID.String(),
		}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.Scope != assetType.Scope.String() || existing.TypeKey != assetType.Key.String() || stringFromPtr(existing.InventoryID) != assetType.InventoryID.String() {
			return ports.ErrForbidden
		}

		updates := map[string]any{
			"display_name": assetType.DisplayName.String(),
			"description":  assetType.Description.String(),
		}
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return customFieldDefinitionWriteError(err)
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) ListTenantCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, page ports.CustomFieldDefinitionPageRequest) ([]customfield.Definition, error) {
	query := s.db.WithContext(ctx).Where(&customFieldDefinitionModel{
		TenantID: tenantID.String(),
		Scope:    customfield.ScopeTenant.String(),
	})
	return s.listCustomFieldDefinitions(ctx, query, page)
}

func (s Store) CustomFieldDefinitionByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definitionID customfield.ID) (customfield.Definition, bool, error) {
	query := s.db.WithContext(ctx).Where(&customFieldDefinitionModel{
		ID:       definitionID.String(),
		TenantID: tenantID.String(),
	})
	if inventoryID.String() == "" {
		query = query.Where(&customFieldDefinitionModel{Scope: customfield.ScopeTenant.String()})
	} else {
		query = query.Where(clause.Or(
			clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
			clause.Eq{Column: "inventory_id", Value: inventoryID.String()},
		))
	}
	var model customFieldDefinitionModel
	err := query.First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return customfield.Definition{}, false, nil
	}
	if err != nil {
		return customfield.Definition{}, false, err
	}
	targetsByDefinitionID, err := customFieldDefinitionTargets(ctx, s.db, []customFieldDefinitionModel{model})
	if err != nil {
		return customfield.Definition{}, false, err
	}
	definition, ok := model.toDomain(targetsByDefinitionID[model.ID])
	if !ok {
		return customfield.Definition{}, false, fmt.Errorf("invalid custom field definition row %q", model.ID)
	}
	return definition, true, nil
}

func (s Store) CustomAssetTypeByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) (customfield.AssetType, bool, error) {
	query := s.db.WithContext(ctx).Where(&customAssetTypeModel{
		ID:       assetTypeID.String(),
		TenantID: tenantID.String(),
	})
	if inventoryID.String() == "" {
		query = query.Where(&customAssetTypeModel{Scope: customfield.ScopeTenant.String()})
	} else {
		query = query.Where(clause.Or(
			clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
			clause.Eq{Column: "inventory_id", Value: inventoryID.String()},
		))
	}
	var model customAssetTypeModel
	err := query.First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return customfield.AssetType{}, false, nil
	}
	if err != nil {
		return customfield.AssetType{}, false, err
	}
	assetType, ok := model.toDomain()
	if !ok {
		return customfield.AssetType{}, false, fmt.Errorf("invalid custom asset type row %q", model.ID)
	}
	return assetType, true, nil
}

func (s Store) ListTenantCustomAssetTypes(ctx context.Context, tenantID tenant.ID, page ports.CustomAssetTypePageRequest) ([]customfield.AssetType, error) {
	query := s.db.WithContext(ctx).Where(&customAssetTypeModel{
		TenantID: tenantID.String(),
		Scope:    customfield.ScopeTenant.String(),
	})
	return s.listCustomAssetTypes(query, page)
}

func (s Store) ListInventoryCustomAssetTypes(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CustomAssetTypePageRequest) ([]customfield.AssetType, error) {
	query := s.db.WithContext(ctx).
		Where(&customAssetTypeModel{TenantID: tenantID.String()}).
		Where(clause.Or(
			clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
			clause.Eq{Column: "inventory_id", Value: inventoryID.String()},
		))
	return s.listCustomAssetTypes(query, page)
}

func (s Store) CustomAssetTypesByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, ids []customfield.AssetTypeID) ([]customfield.AssetType, error) {
	rawIDs := customAssetTypeIDsToStrings(ids)
	if len(rawIDs) == 0 {
		return nil, nil
	}
	query := s.db.WithContext(ctx).
		Where(&customAssetTypeModel{TenantID: tenantID.String()}).
		Where(clause.IN{Column: clause.Column{Name: "id"}, Values: stringValues(rawIDs)}).
		Where(clause.Or(
			clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
			clause.Eq{Column: "inventory_id", Value: inventoryID.String()},
		))
	return s.listCustomAssetTypes(query, ports.CustomAssetTypePageRequest{})
}

func (s Store) listCustomAssetTypes(query *gorm.DB, page ports.CustomAssetTypePageRequest) ([]customfield.AssetType, error) {
	var models []customAssetTypeModel
	if page.AfterAssetTypeKey != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "cursor_key"}, Value: page.AfterAssetTypeKey})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "cursor_key"}}).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]customfield.AssetType, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid custom asset type row %q", model.ID)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s Store) ListInventoryCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CustomFieldDefinitionPageRequest) ([]customfield.Definition, error) {
	query := s.db.WithContext(ctx).
		Where(&customFieldDefinitionModel{TenantID: tenantID.String()}).
		Where(clause.Or(
			clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
			clause.Eq{Column: "inventory_id", Value: inventoryID.String()},
		))
	return s.listCustomFieldDefinitions(ctx, query, page)
}

func (s Store) ListEffectiveCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]customfield.Definition, error) {
	if inventoryID.String() == "" {
		return s.ListTenantCustomFieldDefinitions(ctx, tenantID, ports.CustomFieldDefinitionPageRequest{})
	}
	return s.ListInventoryCustomFieldDefinitions(ctx, tenantID, inventoryID, ports.CustomFieldDefinitionPageRequest{})
}

func (s Store) listCustomFieldDefinitions(ctx context.Context, query *gorm.DB, page ports.CustomFieldDefinitionPageRequest) ([]customfield.Definition, error) {
	var models []customFieldDefinitionModel
	if page.AfterDefinitionKey != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "cursor_key"}, Value: page.AfterDefinitionKey})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "cursor_key"}}).Find(&models).Error; err != nil {
		return nil, err
	}
	targetsByDefinitionID, err := customFieldDefinitionTargets(ctx, s.db, models)
	if err != nil {
		return nil, err
	}
	items := make([]customfield.Definition, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain(targetsByDefinitionID[model.ID])
		if !ok {
			return nil, fmt.Errorf("invalid custom field definition row %q", model.ID)
		}
		items = append(items, item)
	}
	return items, nil
}

func customFieldDefinitionTargets(ctx context.Context, db *gorm.DB, definitions []customFieldDefinitionModel) (map[string][]customfield.AssetTypeID, error) {
	result := map[string][]customfield.AssetTypeID{}
	if len(definitions) == 0 {
		return result, nil
	}
	ids := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		ids = append(ids, definition.ID)
	}
	var models []customFieldDefinitionAssetTypeModel
	if err := db.WithContext(ctx).Where(clause.IN{Column: clause.Column{Name: "custom_field_definition_id"}, Values: stringValues(ids)}).Find(&models).Error; err != nil {
		return nil, err
	}
	for _, model := range models {
		targetID, ok := customfield.NewAssetTypeID(model.CustomAssetTypeID)
		if !ok {
			return nil, fmt.Errorf("invalid custom field asset type target row %q", model.CustomAssetTypeID)
		}
		result[model.CustomFieldDefinitionID] = append(result[model.CustomFieldDefinitionID], targetID)
	}
	return result, nil
}

func customFieldDefinitionWriteError(err error) error {
	if err == nil {
		return nil
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505", "23514":
			return ports.ErrConflict
		}
	}

	if strings.Contains(err.Error(), "constraint failed") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ports.ErrConflict
	}

	return err
}

func (s Store) CreateAsset(ctx context.Context, item asset.Asset, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var containingInventory inventoryModel
		err := tx.Where(&inventoryModel{
			ID:       item.InventoryID.String(),
			TenantID: item.TenantID.String(),
		}).First(&containingInventory).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}

		if item.ParentAssetID.String() != "" {
			var parent assetModel
			err = tx.Where(&assetModel{
				ID:          item.ParentAssetID.String(),
				TenantID:    item.TenantID.String(),
				InventoryID: item.InventoryID.String(),
			}).First(&parent).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrForbidden
			}
			if err != nil {
				return err
			}
			parentKind, ok := asset.NewKind(parent.Kind)
			if !ok || !parentKind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive.String() || parent.ID == item.ID.String() {
				return ports.ErrForbidden
			}
			if err := rejectAssetContainmentCycle(tx, item.ID, parent); err != nil {
				return err
			}
		}
		if item.CustomAssetTypeID.String() != "" {
			var assetType customAssetTypeModel
			err = tx.Where(&customAssetTypeModel{
				ID:       item.CustomAssetTypeID.String(),
				TenantID: item.TenantID.String(),
			}).Where(clause.Or(
				clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
				clause.Eq{Column: "inventory_id", Value: item.InventoryID.String()},
			)).First(&assetType).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrForbidden
			}
			if err != nil {
				return err
			}
		}

		parentAssetID := stringPtrFromAssetID(item.ParentAssetID)
		customAssetTypeID := stringPtrFromCustomAssetTypeID(item.CustomAssetTypeID)
		customFields, err := json.Marshal(item.CustomFields.Values())
		if err != nil {
			return err
		}
		if err := tx.Create(&assetModel{
			ID:                item.ID.String(),
			TenantID:          item.TenantID.String(),
			InventoryID:       item.InventoryID.String(),
			ParentAssetID:     parentAssetID,
			CustomAssetTypeID: customAssetTypeID,
			Kind:              item.Kind.String(),
			Title:             item.Title.String(),
			Description:       item.Description.String(),
			CustomFields:      string(customFields),
			LifecycleState:    item.LifecycleState.String(),
		}).Error; err != nil {
			return err
		}

		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) UpdateAsset(ctx context.Context, item asset.Asset, auditRecords []audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing assetModel
		err := tx.Where(&assetModel{
			ID:          item.ID.String(),
			TenantID:    item.TenantID.String(),
			InventoryID: item.InventoryID.String(),
		}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.Kind != item.Kind.String() || existing.LifecycleState != item.LifecycleState.String() {
			return ports.ErrForbidden
		}
		if stringFromPtr(existing.CustomAssetTypeID) != item.CustomAssetTypeID.String() {
			return ports.ErrForbidden
		}

		if item.ParentAssetID.String() != "" {
			var parent assetModel
			err = tx.Where(&assetModel{
				ID:          item.ParentAssetID.String(),
				TenantID:    item.TenantID.String(),
				InventoryID: item.InventoryID.String(),
			}).First(&parent).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrForbidden
			}
			if err != nil {
				return err
			}
			parentKind, ok := asset.NewKind(parent.Kind)
			if !ok || !parentKind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive.String() || parent.ID == item.ID.String() {
				return ports.ErrForbidden
			}
			if err := rejectAssetContainmentCycle(tx, item.ID, parent); err != nil {
				return err
			}
		}

		customFields, err := json.Marshal(item.CustomFields.Values())
		if err != nil {
			return err
		}
		updates := map[string]any{
			"parent_asset_id":      stringPtrFromAssetID(item.ParentAssetID),
			"custom_asset_type_id": stringPtrFromCustomAssetTypeID(item.CustomAssetTypeID),
			"title":                item.Title.String(),
			"description":          item.Description.String(),
			"custom_fields":        string(customFields),
		}
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
		for _, auditRecord := range auditRecords {
			if err := createAuditRecord(tx, auditRecord); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s Store) AssetByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error) {
	var model assetModel
	err := s.db.WithContext(ctx).Where(&assetModel{
		ID:          assetID.String(),
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return asset.Asset{}, false, nil
	}
	if err != nil {
		return asset.Asset{}, false, err
	}
	item, ok := model.toDomain()
	if !ok {
		return asset.Asset{}, false, fmt.Errorf("invalid asset row %q", model.ID)
	}
	return item, true, nil
}

func (s Store) ListAssetsByInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AssetListPageRequest) ([]asset.Asset, error) {
	var models []assetModel
	query := s.db.WithContext(ctx).Where(&assetModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	})
	if page.AfterAssetID.String() != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterAssetID.String()})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]asset.Asset, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid asset row %q", model.ID)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s Store) SaveAuditRecord(ctx context.Context, record audit.Record) error {
	return createAuditRecord(s.db.WithContext(ctx), record)
}

func createAuditRecord(tx *gorm.DB, record audit.Record) error {
	metadata, err := json.Marshal(record.MetadataValues())
	if err != nil {
		return err
	}
	model := auditRecordModel{
		ID:          record.ID.String(),
		TenantID:    record.TenantID.String(),
		PrincipalID: record.PrincipalID.String(),
		Action:      record.Action.String(),
		Source:      record.Source.String(),
		TargetType:  record.TargetType.String(),
		TargetID:    record.TargetID,
		OccurredAt:  record.OccurredAt,
		RequestID:   record.RequestID,
		Metadata:    string(metadata),
	}
	if record.InventoryID.String() != "" {
		inventoryID := record.InventoryID.String()
		model.InventoryID = &inventoryID
	}
	return tx.Create(&model).Error
}

func (s Store) ListTenantAuditRecords(ctx context.Context, tenantID tenant.ID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	query := s.db.WithContext(ctx).Where(&auditRecordModel{TenantID: tenantID.String()})
	return s.listAuditRecords(query, page)
}

func (s Store) ListInventoryAuditRecords(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	query := s.db.WithContext(ctx).Where(&auditRecordModel{
		TenantID: tenantID.String(),
	})
	query = query.Where(&auditRecordModel{InventoryID: stringPtrFromInventoryID(inventoryID)})
	return s.listAuditRecords(query, page)
}

func (s Store) listAuditRecords(query *gorm.DB, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	var models []auditRecordModel
	if !page.AfterOccurredAt.IsZero() && page.AfterRecordID.String() != "" {
		query = query.Where(clause.Or(
			clause.Gt{Column: clause.Column{Name: "occurred_at"}, Value: page.AfterOccurredAt},
			clause.And(
				clause.Eq{Column: clause.Column{Name: "occurred_at"}, Value: page.AfterOccurredAt},
				clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterRecordID.String()},
			),
		))
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderBy{
		Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "occurred_at"}},
			{Column: clause.Column{Name: "id"}},
		},
	}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]audit.Record, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid audit record row %q", model.ID)
		}
		items = append(items, item)
	}
	return items, nil
}

func rejectAssetContainmentCycle(tx *gorm.DB, assetID asset.ID, parent assetModel) error {
	for current := parent; ; {
		if current.ID == assetID.String() {
			return ports.ErrForbidden
		}
		if current.ParentAssetID == nil {
			return nil
		}

		nextID := *current.ParentAssetID
		var next assetModel
		err := tx.Where(&assetModel{
			ID:          nextID,
			TenantID:    current.TenantID,
			InventoryID: current.InventoryID,
		}).First(&next).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		current = next
	}
}

type tenantModel struct {
	ID        string `gorm:"primaryKey;size:26"`
	Name      string `gorm:"not null;size:120"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (tenantModel) TableName() string {
	return "tenants"
}

type inventoryModel struct {
	ID        string      `gorm:"primaryKey;size:26"`
	TenantID  string      `gorm:"not null;size:26;index"`
	Tenant    tenantModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	Name      string      `gorm:"not null;size:120"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type inventoryAccessGrantModel struct {
	TenantID     string         `gorm:"primaryKey;size:26;index:idx_inventory_access_grants_inventory"`
	Tenant       tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID  string         `gorm:"primaryKey;size:26;index:idx_inventory_access_grants_inventory"`
	Inventory    inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:InventoryID;references:ID"`
	GrantKey     string         `gorm:"primaryKey;size:180"`
	PrincipalID  string         `gorm:"not null;size:128;index"`
	Relationship string         `gorm:"not null;size:32;check:chk_inventory_access_grants_relationship,relationship IN ('viewer','editor')"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type customFieldDefinitionModel struct {
	ID            string          `gorm:"primaryKey;size:26"`
	TenantID      string          `gorm:"not null;size:26;index"`
	Tenant        tenantModel     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID   *string         `gorm:"size:26;index"`
	Inventory     *inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:InventoryID;references:ID"`
	Scope         string          `gorm:"not null;size:32;index;check:chk_custom_field_definitions_scope,scope IN ('tenant','inventory')"`
	CursorKey     string          `gorm:"not null;size:32;index"`
	FieldKey      string          `gorm:"not null;size:80;index"`
	DisplayName   string          `gorm:"not null;size:120"`
	FieldType     string          `gorm:"not null;size:32;check:chk_custom_field_definitions_field_type,field_type IN ('text','number','boolean','date','url','enum')"`
	EnumOptions   string          `gorm:"type:jsonb;not null;default:'[]'"`
	Applicability string          `gorm:"not null;size:32;default:'all_assets';check:chk_custom_field_definitions_applicability,applicability IN ('all_assets','custom_asset_types')"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type customAssetTypeModel struct {
	ID          string          `gorm:"primaryKey;size:26"`
	TenantID    string          `gorm:"not null;size:26;index"`
	Tenant      tenantModel     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID *string         `gorm:"size:26;index"`
	Inventory   *inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:InventoryID;references:ID"`
	Scope       string          `gorm:"not null;size:32;index;check:chk_custom_asset_types_scope,scope IN ('tenant','inventory')"`
	CursorKey   string          `gorm:"not null;size:32;index"`
	TypeKey     string          `gorm:"not null;size:80;index"`
	DisplayName string          `gorm:"not null;size:120"`
	Description string          `gorm:"not null;default:'';size:1000"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type customFieldDefinitionAssetTypeModel struct {
	CustomFieldDefinitionID string                     `gorm:"primaryKey;size:26"`
	CustomFieldDefinition   customFieldDefinitionModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:CustomFieldDefinitionID;references:ID"`
	CustomAssetTypeID       string                     `gorm:"primaryKey;size:26"`
	CustomAssetType         customAssetTypeModel       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:CustomAssetTypeID;references:ID"`
	TenantID                string                     `gorm:"not null;size:26;index"`
	Tenant                  tenantModel                `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID             *string                    `gorm:"size:26;index"`
	Inventory               *inventoryModel            `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:InventoryID;references:ID"`
	CreatedAt               time.Time
}

type assetModel struct {
	ID                string                `gorm:"primaryKey;size:26"`
	TenantID          string                `gorm:"not null;size:26;index:idx_assets_tenant_inventory"`
	Tenant            tenantModel           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID       string                `gorm:"not null;size:26;index:idx_assets_tenant_inventory;index:idx_assets_inventory_parent;index:idx_assets_inventory_kind"`
	Inventory         inventoryModel        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	ParentAssetID     *string               `gorm:"size:26;index;index:idx_assets_inventory_parent"`
	ParentAsset       *assetModel           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:ParentAssetID;references:ID"`
	CustomAssetTypeID *string               `gorm:"size:26;index"`
	CustomAssetType   *customAssetTypeModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:CustomAssetTypeID;references:ID"`
	Kind              string                `gorm:"not null;size:32;index:idx_assets_inventory_kind;check:chk_assets_kind,kind IN ('item','container','location')"`
	Title             string                `gorm:"not null;size:160"`
	Description       string                `gorm:"not null;default:''"`
	CustomFields      string                `gorm:"type:jsonb;not null;default:'{}'"`
	LifecycleState    string                `gorm:"not null;size:32;check:chk_assets_lifecycle_state,lifecycle_state IN ('active','archived')"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type auditRecordModel struct {
	ID          string          `gorm:"primaryKey;size:26"`
	TenantID    string          `gorm:"not null;size:26;index:idx_audit_records_tenant_id"`
	Tenant      tenantModel     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID *string         `gorm:"size:26;index:idx_audit_records_inventory_id"`
	Inventory   *inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	PrincipalID string          `gorm:"not null;size:128;index"`
	Action      string          `gorm:"not null;size:80;index"`
	Source      string          `gorm:"not null;size:40"`
	TargetType  string          `gorm:"not null;size:80;index"`
	TargetID    string          `gorm:"not null;size:180;index"`
	OccurredAt  time.Time       `gorm:"not null;index"`
	RequestID   string          `gorm:"not null;default:'';size:128;index"`
	Metadata    string          `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type authorizationOutboxEventModel struct {
	ID               string         `gorm:"primaryKey;size:26"`
	Kind             string         `gorm:"not null;size:80;index;check:chk_authorization_outbox_events_kind,kind IN ('grant_tenant_owner','grant_inventory_owner','grant_inventory_viewer','grant_inventory_editor');check:chk_authorization_outbox_events_inventory_required,(kind IN ('grant_inventory_owner','grant_inventory_viewer','grant_inventory_editor') AND inventory_id IS NOT NULL) OR (kind = 'grant_tenant_owner' AND inventory_id IS NULL)"`
	PrincipalID      string         `gorm:"not null;size:128;index"`
	TenantID         string         `gorm:"not null;size:26;index"`
	Tenant           tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID      *string        `gorm:"size:26;index"`
	Inventory        inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	Attempts         int            `gorm:"not null;default:0"`
	LastError        string         `gorm:"not null;default:''"`
	ClaimID          string         `gorm:"not null;default:'';size:26;index"`
	ClaimedUntil     *time.Time     `gorm:"index"`
	ProcessedAt      *time.Time
	DeadLetteredAt   *time.Time `gorm:"index"`
	DeadLetterReason string     `gorm:"not null;default:''"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (authorizationOutboxEventModel) TableName() string {
	return "authorization_outbox_events"
}

func (inventoryModel) TableName() string {
	return "inventories"
}

func (inventoryAccessGrantModel) TableName() string {
	return "inventory_access_grants"
}

func (customFieldDefinitionModel) TableName() string {
	return "custom_field_definitions"
}

func (customAssetTypeModel) TableName() string {
	return "custom_asset_types"
}

func (customFieldDefinitionAssetTypeModel) TableName() string {
	return "custom_field_definition_asset_types"
}

func (m *customFieldDefinitionModel) BeforeSave(*gorm.DB) error {
	scope := customfield.Scope(m.Scope)
	prefix := "1:"
	if scope == customfield.ScopeTenant {
		prefix = "0:"
	}
	m.CursorKey = prefix + m.ID
	return nil
}

func (m *customAssetTypeModel) BeforeSave(*gorm.DB) error {
	scope := customfield.Scope(m.Scope)
	prefix := "1:"
	if scope == customfield.ScopeTenant {
		prefix = "0:"
	}
	m.CursorKey = prefix + m.ID
	return nil
}

func (m *inventoryAccessGrantModel) BeforeSave(*gorm.DB) error {
	m.GrantKey = ports.InventoryAccessGrant{
		PrincipalID:  identity.PrincipalID(m.PrincipalID),
		Relationship: ports.InventoryAccessRelationship(m.Relationship),
	}.CursorKey()
	return nil
}

func (assetModel) TableName() string {
	return "assets"
}

func (auditRecordModel) TableName() string {
	return "audit_records"
}

func (m inventoryModel) toDomain() (inventory.Inventory, bool) {
	id, ok := inventory.NewID(m.ID)
	if !ok {
		return inventory.Inventory{}, false
	}
	name, ok := inventory.NewName(m.Name)
	if !ok {
		return inventory.Inventory{}, false
	}

	return inventory.Inventory{
		ID:       id,
		TenantID: inventory.TenantID(m.TenantID),
		Name:     name,
	}, true
}

func (m authorizationOutboxEventModel) toPort() ports.AuthorizationOutboxEvent {
	inventoryID := inventory.InventoryID("")
	if m.InventoryID != nil {
		inventoryID = inventory.InventoryID(*m.InventoryID)
	}
	event := ports.AuthorizationOutboxEvent{
		ID:          m.ID,
		Kind:        ports.AuthorizationOutboxEventKind(m.Kind),
		PrincipalID: identity.PrincipalID(m.PrincipalID),
		TenantID:    tenant.ID(m.TenantID),
		InventoryID: inventoryID,
		Attempts:    m.Attempts,
		LastError:   m.LastError,
		ClaimID:     m.ClaimID,
		CreatedAt:   m.CreatedAt,
	}
	if m.ClaimedUntil != nil {
		event.ClaimedUntil = *m.ClaimedUntil
	}
	if m.DeadLetteredAt != nil {
		event.DeadLetteredAt = *m.DeadLetteredAt
	}
	event.DeadLetterReason = m.DeadLetterReason
	return event
}

func (m inventoryAccessGrantModel) toPort() (ports.InventoryAccessGrant, bool) {
	principalID, ok := identity.NewPrincipalID(m.PrincipalID)
	if !ok {
		return ports.InventoryAccessGrant{}, false
	}
	relationship := ports.InventoryAccessRelationship(m.Relationship)
	switch relationship {
	case ports.InventoryAccessViewer, ports.InventoryAccessEditor:
	default:
		return ports.InventoryAccessGrant{}, false
	}
	return ports.InventoryAccessGrant{
		TenantID:     tenant.ID(m.TenantID),
		InventoryID:  inventory.InventoryID(m.InventoryID),
		PrincipalID:  principalID,
		Relationship: relationship,
	}, true
}

func (m customAssetTypeModel) toDomain() (customfield.AssetType, bool) {
	id, ok := customfield.NewAssetTypeID(m.ID)
	if !ok {
		return customfield.AssetType{}, false
	}
	key, ok := customfield.NewKey(m.TypeKey)
	if !ok {
		return customfield.AssetType{}, false
	}
	displayName, ok := customfield.NewDisplayName(m.DisplayName)
	if !ok {
		return customfield.AssetType{}, false
	}
	description, ok := customfield.NewDescription(m.Description)
	if !ok {
		return customfield.AssetType{}, false
	}
	scope := customfield.Scope(m.Scope)
	inventoryID := customfield.InventoryID("")
	if m.InventoryID != nil {
		inventoryID = customfield.InventoryID(*m.InventoryID)
	}
	return customfield.NewAssetType(
		id,
		customfield.TenantID(m.TenantID),
		inventoryID,
		scope,
		key,
		displayName,
		description,
	)
}

func (m customFieldDefinitionModel) toDomain(customAssetTypeIDs []customfield.AssetTypeID) (customfield.Definition, bool) {
	id, ok := customfield.NewID(m.ID)
	if !ok {
		return customfield.Definition{}, false
	}
	key, ok := customfield.NewKey(m.FieldKey)
	if !ok {
		return customfield.Definition{}, false
	}
	displayName, ok := customfield.NewDisplayName(m.DisplayName)
	if !ok {
		return customfield.Definition{}, false
	}
	fieldType, ok := customfield.NewFieldType(m.FieldType)
	if !ok {
		return customfield.Definition{}, false
	}
	scope := customfield.Scope(m.Scope)
	inventoryID := customfield.InventoryID("")
	if m.InventoryID != nil {
		inventoryID = customfield.InventoryID(*m.InventoryID)
	}
	var rawOptions []string
	if err := json.Unmarshal([]byte(m.EnumOptions), &rawOptions); err != nil {
		return customfield.Definition{}, false
	}
	applicability, ok := customfield.NewApplicability(m.Applicability)
	if !ok {
		return customfield.Definition{}, false
	}
	options := make([]customfield.Key, 0, len(rawOptions))
	for _, raw := range rawOptions {
		option, ok := customfield.NewKey(raw)
		if !ok {
			return customfield.Definition{}, false
		}
		options = append(options, option)
	}
	return customfield.NewDefinition(
		id,
		customfield.TenantID(m.TenantID),
		inventoryID,
		scope,
		key,
		displayName,
		fieldType,
		options,
		applicability,
		customAssetTypeIDs,
	)
}

func (m assetModel) toDomain() (asset.Asset, bool) {
	id, ok := asset.NewID(m.ID)
	if !ok {
		return asset.Asset{}, false
	}
	kind, ok := asset.NewKind(m.Kind)
	if !ok {
		return asset.Asset{}, false
	}
	title, ok := asset.NewTitle(m.Title)
	if !ok {
		return asset.Asset{}, false
	}
	var customFieldValues map[string]any
	if err := json.Unmarshal([]byte(m.CustomFields), &customFieldValues); err != nil {
		return asset.Asset{}, false
	}
	customFields, ok := asset.NewCustomFields(customFieldValues)
	if !ok {
		return asset.Asset{}, false
	}
	lifecycleState := asset.LifecycleState(m.LifecycleState)
	switch lifecycleState {
	case asset.LifecycleStateActive, asset.LifecycleStateArchived:
	default:
		return asset.Asset{}, false
	}
	parentID := asset.ID("")
	if m.ParentAssetID != nil {
		parentID, ok = asset.NewID(*m.ParentAssetID)
		if !ok {
			return asset.Asset{}, false
		}
	}
	customAssetTypeID := asset.CustomAssetTypeID("")
	if m.CustomAssetTypeID != nil {
		customAssetTypeID, ok = asset.NewCustomAssetTypeID(*m.CustomAssetTypeID)
		if !ok {
			return asset.Asset{}, false
		}
	}
	return asset.Asset{
		ID:                id,
		TenantID:          asset.TenantID(m.TenantID),
		InventoryID:       asset.InventoryID(m.InventoryID),
		ParentAssetID:     parentID,
		CustomAssetTypeID: customAssetTypeID,
		Kind:              kind,
		Title:             title,
		Description:       asset.NewDescription(m.Description),
		CustomFields:      customFields,
		LifecycleState:    lifecycleState,
	}, true
}

func (m auditRecordModel) toDomain() (audit.Record, bool) {
	id, ok := audit.NewID(m.ID)
	if !ok {
		return audit.Record{}, false
	}
	action, ok := audit.NewAction(m.Action)
	if !ok {
		return audit.Record{}, false
	}
	source, ok := audit.NewSource(m.Source)
	if !ok {
		return audit.Record{}, false
	}
	targetType, ok := audit.NewTargetType(m.TargetType)
	if !ok {
		return audit.Record{}, false
	}
	inventoryID := audit.InventoryID("")
	if m.InventoryID != nil {
		inventoryID = audit.InventoryID(*m.InventoryID)
	}
	metadata := map[string]string{}
	if err := json.Unmarshal([]byte(m.Metadata), &metadata); err != nil {
		return audit.Record{}, false
	}
	return audit.NewRecord(
		id,
		audit.TenantID(m.TenantID),
		inventoryID,
		audit.PrincipalID(m.PrincipalID),
		action,
		source,
		targetType,
		m.TargetID,
		m.OccurredAt,
		m.RequestID,
		metadata,
	)
}

func stringPtrFromAssetID(id asset.ID) *string {
	if id.String() == "" {
		return nil
	}
	value := id.String()
	return &value
}

func stringPtrFromCustomAssetTypeID(id asset.CustomAssetTypeID) *string {
	if id.String() == "" {
		return nil
	}
	value := id.String()
	return &value
}

func stringFromPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringPtrFromInventoryID(id inventory.InventoryID) *string {
	if id.String() == "" {
		return nil
	}
	value := id.String()
	return &value
}

func outboxKindForInventoryAccess(relationship ports.InventoryAccessRelationship) ports.AuthorizationOutboxEventKind {
	switch relationship {
	case ports.InventoryAccessEditor:
		return ports.AuthorizationOutboxGrantInventoryEditor
	default:
		return ports.AuthorizationOutboxGrantInventoryViewer
	}
}

func customFieldKeysToStrings(keys []customfield.Key) []string {
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, key.String())
	}
	return values
}

func customAssetTypeIDsToStrings(ids []customfield.AssetTypeID) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, id.String())
	}
	return values
}

func stringValues(values []string) []any {
	items := make([]any, 0, len(values))
	for _, value := range values {
		items = append(items, value)
	}
	return items
}

package gormstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"slices"
	"strings"
)

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
			var target customAssetTypeModel
			err := tx.Where(&customAssetTypeModel{
				ID:             targetID.String(),
				TenantID:       definition.TenantID.String(),
				LifecycleState: customfield.AssetTypeLifecycleActive.String(),
			}).Where(clause.Or(
				clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
				clause.Eq{Column: "inventory_id", Value: definition.InventoryID.String()},
			)).First(&target).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrForbidden
			}
			if err != nil {
				return err
			}
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

func customFieldKeysToStrings(keys []customfield.Key) []string {
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, key.String())
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

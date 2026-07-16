package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
)

func (s *Store) SaveCustomFieldDefinition(_ context.Context, definition customfield.Definition, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.customFieldDefinitionParentIsValid(definition); err != nil {
		return err
	}
	for _, targetID := range definition.CustomAssetTypeIDs {
		target, ok := s.customAssetTypes[targetID]
		if !ok || !target.IsActive() || !customFieldTargetInScope(definition, target) {
			return ports.ErrForbidden
		}
	}
	for _, existing := range s.customFields {
		if customfield.DefinitionsConflict(existing, definition) {
			return ports.ErrConflict
		}
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	if definition.LifecycleState.String() == "" {
		definition.LifecycleState = customfield.DefinitionLifecycleActive
	}
	s.customFields[definition.ID] = definition
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) UpdateCustomFieldDefinition(_ context.Context, definition customfield.Definition, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.customFields[definition.ID]
	if !exists || existing.TenantID != definition.TenantID || existing.InventoryID != definition.InventoryID || existing.Scope != definition.Scope || !existing.IsActive() || definition.LifecycleState != customfield.DefinitionLifecycleActive {
		return ports.ErrForbidden
	}
	schemaChange, ok := existing.CompatibleSchemaChange(definition)
	if !ok {
		return ports.ErrForbidden
	}
	for _, targetID := range schemaChange.AddedCustomAssetTypeIDs {
		target, ok := s.customAssetTypes[targetID]
		if !ok || !target.IsActive() || !customFieldTargetInScope(definition, target) {
			return ports.ErrForbidden
		}
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.customFields[definition.ID] = definition
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) UpdateCustomFieldDefinitionLifecycle(_ context.Context, definition customfield.Definition, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.customFields[definition.ID]
	if !exists || existing.TenantID != definition.TenantID || existing.InventoryID != definition.InventoryID || existing.Scope != definition.Scope || existing.Key != definition.Key || existing.LifecycleState == definition.LifecycleState {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.customFields[definition.ID] = definition
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) DeleteCustomFieldDefinition(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definitionID customfield.ID, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	definition, ok := s.customFields[definitionID]
	if !ok || definition.TenantID.String() != tenantID.String() {
		return ports.ErrForbidden
	}
	if inventoryID.String() == "" && definition.Scope != customfield.ScopeTenant {
		return ports.ErrForbidden
	}
	if inventoryID.String() != "" && definition.Scope == customfield.ScopeInventory && definition.InventoryID.String() != inventoryID.String() {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.auditRecords[auditRecord.ID] = auditRecord
	delete(s.customFields, definitionID)
	return nil
}

func (s *Store) CustomFieldDefinitionHasActiveAssetValues(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definition customfield.Definition) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.assets {
		if item.TenantID.String() != tenantID.String() || item.LifecycleState.String() != "active" {
			continue
		}
		if inventoryID.String() != "" && item.InventoryID.String() != inventoryID.String() {
			continue
		}
		if item.CustomFields.HasNonEmptyValue(definition.Key.String()) {
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) CustomFieldDefinitionByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, definitionID customfield.ID) (customfield.Definition, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	definition, ok := s.customFields[definitionID]
	if !ok || definition.TenantID.String() != tenantID.String() {
		return customfield.Definition{}, false, nil
	}
	if inventoryID.String() == "" {
		if definition.Scope != customfield.ScopeTenant {
			return customfield.Definition{}, false, nil
		}
		return definition, true, nil
	}
	if definition.Scope == customfield.ScopeInventory && definition.InventoryID.String() != inventoryID.String() {
		return customfield.Definition{}, false, nil
	}
	return definition, true, nil
}

func (s *Store) ListTenantCustomFieldDefinitions(_ context.Context, tenantID tenant.ID, page ports.CustomFieldDefinitionPageRequest) ([]customfield.Definition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []customfield.Definition{}
	for _, definition := range s.customFields {
		if definition.TenantID.String() == tenantID.String() && definition.Scope == customfield.ScopeTenant && page.Lifecycle.Includes(definition.LifecycleState.String()) && definition.CursorKey() > page.AfterDefinitionKey {
			items = append(items, definition)
		}
	}
	return pagedCustomFieldDefinitions(items, page.Limit), nil
}

func (s *Store) ListInventoryCustomFieldDefinitions(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CustomFieldDefinitionPageRequest) ([]customfield.Definition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []customfield.Definition{}
	for _, definition := range s.customFields {
		if definition.TenantID.String() != tenantID.String() || !page.Lifecycle.Includes(definition.LifecycleState.String()) || definition.CursorKey() <= page.AfterDefinitionKey {
			continue
		}
		if definition.Scope == customfield.ScopeTenant || definition.InventoryID.String() == inventoryID.String() {
			items = append(items, definition)
		}
	}
	return pagedCustomFieldDefinitions(items, page.Limit), nil
}

func (s *Store) ListEffectiveCustomFieldDefinitions(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]customfield.Definition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []customfield.Definition{}
	for _, definition := range s.customFields {
		if definition.TenantID.String() != tenantID.String() || !definition.IsActive() {
			continue
		}
		if inventoryID.String() == "" {
			if definition.Scope == customfield.ScopeTenant {
				items = append(items, definition)
			}
			continue
		}
		if definition.Scope == customfield.ScopeTenant || definition.InventoryID.String() == inventoryID.String() {
			items = append(items, definition)
		}
	}
	return pagedCustomFieldDefinitions(items, 0), nil
}

func pagedCustomFieldDefinitions(items []customfield.Definition, limit int) []customfield.Definition {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].CursorKey() < items[right].CursorKey()
	})
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

func (s *Store) customFieldDefinitionParentIsValid(definition customfield.Definition) error {
	if _, exists := s.tenants[tenant.ID(definition.TenantID.String())]; !exists {
		return ports.ErrForbidden
	}
	if definition.Scope == customfield.ScopeInventory {
		item, ok := s.inventories[inventory.InventoryID(definition.InventoryID.String())]
		if !ok || item.TenantID.String() != definition.TenantID.String() {
			return ports.ErrForbidden
		}
	}
	return nil
}

func customFieldTargetInScope(definition customfield.Definition, target customfield.AssetType) bool {
	if target.TenantID != definition.TenantID {
		return false
	}
	if definition.Scope == customfield.ScopeTenant {
		return target.Scope == customfield.ScopeTenant
	}
	return target.Scope == customfield.ScopeTenant || target.InventoryID == definition.InventoryID
}

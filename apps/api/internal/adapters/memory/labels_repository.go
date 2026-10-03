package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func (s *Store) LabelInstance(context.Context) (printing.InstanceID, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.labelInstance, s.labelInstance != "", nil
}
func (s *Store) BootstrapLabelInstance(_ context.Context, id printing.InstanceID) (printing.InstanceID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !printing.ValidOpaqueID(string(id)) {
		return "", ports.ErrConflict
	}
	if s.labelInstance == "" {
		s.labelInstance = id
	}
	return s.labelInstance, nil
}
func (s *Store) LabelForAsset(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (printing.Label, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, value := range s.labels {
		if value.TenantID == tenantID.String() && value.InventoryID == inventoryID.String() && value.AssetID == assetID.String() {
			return value, true, nil
		}
	}
	return printing.Label{}, false, nil
}
func (s *Store) LookupLabel(_ context.Context, instance printing.InstanceID, id printing.LabelID) (printing.Label, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.labels[id]
	return value, found && value.InstanceID == instance, nil
}
func (s *Store) ProvisionLabel(_ context.Context, value printing.Label, record audit.Record) (printing.Label, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, found := s.assets[asset.ID(value.AssetID)]
	if !found || item.TenantID.String() != value.TenantID || item.InventoryID.String() != value.InventoryID || value.InstanceID != s.labelInstance || !printing.ValidOpaqueID(string(value.ID)) {
		return printing.Label{}, false, ports.ErrForbidden
	}
	for _, existing := range s.labels {
		if existing.TenantID == value.TenantID && existing.InventoryID == value.InventoryID && existing.AssetID == value.AssetID {
			return existing, false, nil
		}
	}
	if _, exists := s.labels[value.ID]; exists {
		return printing.Label{}, false, ports.ErrConflict
	}
	if record.TenantID.String() != value.TenantID || record.InventoryID.String() != value.InventoryID || record.TargetID != value.AssetID {
		return printing.Label{}, false, ports.ErrForbidden
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return printing.Label{}, false, ports.ErrConflict
	}
	s.labels[value.ID] = value
	s.auditRecords[record.ID] = record
	return value, true, nil
}
func (s *Store) SaveLabelRender(_ context.Context, value printing.LabelRender, record audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, found := s.assets[asset.ID(value.AssetID)]
	label, hasLabel := s.labels[value.LabelID]
	if !found || !hasLabel || label.Tombstoned || label.AssetID != value.AssetID || item.TenantID.String() != value.TenantID || item.InventoryID.String() != value.InventoryID || record.TenantID.String() != value.TenantID || record.InventoryID.String() != value.InventoryID || record.TargetID != value.AssetID {
		return ports.ErrForbidden
	}
	if _, exists := s.labelRenders[value.ID]; exists {
		return ports.ErrConflict
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.ErrConflict
	}
	s.labelRenders[value.ID] = value.Clone()
	s.auditRecords[record.ID] = record
	return nil
}
func (s *Store) LabelRenderByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, id printing.RenderID) (printing.LabelRender, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.labelRenders[id]
	if !found || value.TenantID != tenantID.String() || value.InventoryID != inventoryID.String() {
		return printing.LabelRender{}, false, nil
	}
	return value.Clone(), true, nil
}
func (s *Store) PurgeExpiredLabelRenders(_ context.Context, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, value := range s.labelRenders {
		if !now.Before(value.ExpiresAt) {
			delete(s.labelRenders, id)
		}
	}
	return nil
}
func (s *Store) tombstoneAssetLabelsLocked(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) {
	for id, value := range s.labels {
		if value.TenantID == tenantID.String() && value.InventoryID == inventoryID.String() && value.AssetID == assetID.String() {
			value.Tombstoned = true
			s.labels[id] = value
		}
	}
	for id, value := range s.labelRenders {
		if value.TenantID == tenantID.String() && value.InventoryID == inventoryID.String() && value.AssetID == assetID.String() {
			delete(s.labelRenders, id)
		}
	}
}

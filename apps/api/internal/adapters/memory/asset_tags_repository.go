package memory

import (
	"context"
	"sort"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) CreateAssetTag(_ context.Context, tag assettag.Tag, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.assetTags[tag.ID]; exists {
		return ports.ErrConflict
	}
	if tag.TenantID.String() == "" || tag.InventoryID.String() == "" || tag.Key.String() == "" {
		return ports.ErrInvalidProviderInput
	}
	for _, existing := range s.assetTags {
		if existing.TenantID == tag.TenantID && existing.InventoryID == tag.InventoryID && existing.Key == tag.Key {
			return ports.ErrConflict
		}
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.assetTags[tag.ID] = tag
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) SetAssetTags(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, tagIDs []assettag.ID, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, found := s.assets[assetID]
	if !found || item.TenantID.String() != tenantID.String() || item.InventoryID.String() != inventoryID.String() {
		return ports.ErrForbidden
	}
	next := map[assettag.ID]struct{}{}
	for _, tagID := range tagIDs {
		tag, found := s.assetTags[tagID]
		if !found || tag.TenantID.String() != tenantID.String() || tag.InventoryID.String() != inventoryID.String() || tag.LifecycleState != assettag.LifecycleStateActive {
			return ports.ErrForbidden
		}
		next[tagID] = struct{}{}
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	if len(next) == 0 {
		delete(s.assetTagLinks, assetID)
	} else {
		s.assetTagLinks[assetID] = next
	}
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) UpdateAssetTag(_ context.Context, tag assettag.Tag, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, found := s.assetTags[tag.ID]
	if !found || existing.TenantID != tag.TenantID || existing.InventoryID != tag.InventoryID || existing.Key != tag.Key || existing.LifecycleState != assettag.LifecycleStateActive || tag.LifecycleState != assettag.LifecycleStateActive {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.assetTags[tag.ID] = tag
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) UpdateAssetTagLifecycle(_ context.Context, tag assettag.Tag, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, found := s.assetTags[tag.ID]
	if !found || existing.TenantID != tag.TenantID || existing.InventoryID != tag.InventoryID || existing.Key != tag.Key {
		return ports.ErrForbidden
	}
	if existing.LifecycleState != assettag.LifecycleStateActive || tag.LifecycleState != assettag.LifecycleStateArchived {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.assetTags[tag.ID] = tag
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) AssetTagByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, tagID assettag.ID) (assettag.Tag, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tag, found := s.assetTags[tagID]
	if !found || tag.TenantID.String() != tenantID.String() || tag.InventoryID.String() != inventoryID.String() {
		return assettag.Tag{}, false, nil
	}
	return tag, true, nil
}

func (s *Store) AssetTagByKey(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, key assettag.Key) (assettag.Tag, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, tag := range s.assetTags {
		if tag.TenantID.String() == tenantID.String() && tag.InventoryID.String() == inventoryID.String() && tag.Key == key {
			return tag, true, nil
		}
	}
	return assettag.Tag{}, false, nil
}

func (s *Store) ListAssetTags(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AssetTagPageRequest) ([]assettag.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tags := []assettag.Tag{}
	for _, tag := range s.assetTags {
		if tag.TenantID.String() != tenantID.String() || tag.InventoryID.String() != inventoryID.String() || tag.LifecycleState != assettag.LifecycleStateActive {
			continue
		}
		if page.AfterTagID.String() != "" && tag.ID.String() <= page.AfterTagID.String() {
			continue
		}
		tags = append(tags, tag)
	}
	sort.Slice(tags, func(left int, right int) bool {
		return tags[left].ID.String() < tags[right].ID.String()
	})
	if page.Limit > 0 && len(tags) > page.Limit {
		tags = tags[:page.Limit]
	}
	if tags == nil {
		return []assettag.Tag{}, nil
	}
	return tags, nil
}

func (s *Store) AssetTagsByAsset(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) ([]assettag.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.assetTagsByAssetLocked(tenantID, inventoryID, assetID), nil
}

func (s *Store) AssetTagsByAssets(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetIDs []asset.ID) (map[asset.ID][]assettag.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := map[asset.ID][]assettag.Tag{}
	for _, assetID := range assetIDs {
		out[assetID] = s.assetTagsByAssetLocked(tenantID, inventoryID, assetID)
	}
	return out, nil
}

func (s *Store) assetTagsByAssetLocked(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) []assettag.Tag {
	links := s.assetTagLinks[assetID]
	tags := make([]assettag.Tag, 0, len(links))
	for tagID := range links {
		tag, found := s.assetTags[tagID]
		if !found || tag.TenantID.String() != tenantID.String() || tag.InventoryID.String() != inventoryID.String() || tag.LifecycleState != assettag.LifecycleStateActive {
			continue
		}
		tags = append(tags, tag)
	}
	sort.Slice(tags, func(left int, right int) bool {
		return tags[left].Key.String() < tags[right].Key.String()
	})
	return tags
}

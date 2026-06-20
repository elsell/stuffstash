package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
)

func (s *Store) SaveAttachment(_ context.Context, attachment media.Attachment, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.attachments[attachment.ID]; exists {
		return ports.ErrConflict
	}
	item, ok := s.assets[asset.ID(attachment.AssetID.String())]
	if !ok || item.TenantID.String() != attachment.TenantID.String() || item.InventoryID.String() != attachment.InventoryID.String() {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	if attachment.LifecycleState.String() == "" {
		attachment.LifecycleState = media.LifecycleStateActive
	}
	s.attachments[attachment.ID] = attachment
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) UpdateAttachmentLifecycle(_ context.Context, attachment media.Attachment, auditRecord audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.attachments[attachment.ID]
	if !ok || existing.TenantID != attachment.TenantID || existing.InventoryID != attachment.InventoryID || existing.AssetID != attachment.AssetID || existing.LifecycleState == attachment.LifecycleState {
		return ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	s.attachments[attachment.ID] = attachment
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
}

func (s *Store) DeleteAttachment(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID, auditRecord audit.Record) (media.Attachment, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	attachment, ok := s.attachments[attachmentID]
	if !ok || attachment.TenantID.String() != tenantID.String() || attachment.InventoryID.String() != inventoryID.String() || attachment.AssetID.String() != assetID.String() {
		return media.Attachment{}, false, nil
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return media.Attachment{}, false, ports.ErrConflict
	}
	s.auditRecords[auditRecord.ID] = auditRecord
	delete(s.attachments, attachmentID)
	return attachment, true, nil
}

func (s *Store) AttachmentByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID) (media.Attachment, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attachment, ok := s.attachments[attachmentID]
	if !ok || attachment.TenantID.String() != tenantID.String() || attachment.InventoryID.String() != inventoryID.String() || attachment.AssetID.String() != assetID.String() {
		return media.Attachment{}, false, nil
	}
	return attachment, true, nil
}

func (s *Store) ListAttachmentsByAsset(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, page ports.AttachmentListPageRequest) ([]media.Attachment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []media.Attachment{}
	for _, attachment := range s.attachments {
		if attachment.TenantID.String() == tenantID.String() && attachment.InventoryID.String() == inventoryID.String() && attachment.AssetID.String() == assetID.String() && attachment.IsActive() && attachment.ID.String() > page.AfterAttachmentID.String() {
			items = append(items, attachment)
		}
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].ID.String() < items[right].ID.String()
	})
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

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
	s.attachments[attachment.ID] = attachment
	s.auditRecords[auditRecord.ID] = auditRecord
	return nil
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
		if attachment.TenantID.String() == tenantID.String() && attachment.InventoryID.String() == inventoryID.String() && attachment.AssetID.String() == assetID.String() && attachment.ID.String() > page.AfterAttachmentID.String() {
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

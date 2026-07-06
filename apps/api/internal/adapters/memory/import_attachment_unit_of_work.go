package memory

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) CreateImportedAttachment(_ context.Context, attachment media.Attachment, auditRecord audit.Record, link ports.ImportSourceLink, record ports.ImportJobResource) error {
	if err := validateImportSourceLink(link); err != nil {
		return err
	}
	if err := validateImportJobResource(record); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.attachments[attachment.ID]; exists {
		return ports.ErrConflict
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.ErrConflict
	}
	linkKey := importSourceLinkKey(link.Key)
	if _, exists := s.importLinks[linkKey]; exists {
		return ports.ErrConflict
	}
	resourceKey := importJobResourceKey(record)
	if _, exists := s.importResources[resourceKey]; exists {
		return ports.ErrConflict
	}
	item, ok := s.assets[asset.ID(attachment.AssetID.String())]
	if !ok || item.TenantID.String() != attachment.TenantID.String() || item.InventoryID.String() != attachment.InventoryID.String() {
		return ports.ErrForbidden
	}
	if attachment.LifecycleState.String() == "" {
		attachment.LifecycleState = media.LifecycleStateActive
	}
	s.attachments[attachment.ID] = attachment
	s.auditRecords[auditRecord.ID] = auditRecord
	s.importLinks[linkKey] = link
	s.importResources[resourceKey] = record
	return nil
}

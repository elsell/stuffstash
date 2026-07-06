package memory

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) CreateImportedAsset(ctx context.Context, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation, promotedParent *asset.Asset, parentAuditRecord *audit.Record, link ports.ImportSourceLink, record ports.ImportJobResource) error {
	if err := validateImportSourceLink(link); err != nil {
		return err
	}
	if err := validateImportJobResource(record); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	linkKey := importSourceLinkKey(link.Key)
	if _, exists := s.importLinks[linkKey]; exists {
		return ports.ErrConflict
	}
	resourceKey := importJobResourceKey(record)
	if _, exists := s.importResources[resourceKey]; exists {
		return ports.ErrConflict
	}
	if promotedParent != nil && parentAuditRecord != nil {
		if err := s.createAssetWithParentPromotionLocked(*promotedParent, *parentAuditRecord, item, auditRecord, undoableOperation); err != nil {
			return err
		}
	} else if err := s.createAssetLocked(item, auditRecord, undoableOperation); err != nil {
		return err
	}
	s.importLinks[linkKey] = link
	s.importResources[resourceKey] = record
	_ = ctx
	return nil
}

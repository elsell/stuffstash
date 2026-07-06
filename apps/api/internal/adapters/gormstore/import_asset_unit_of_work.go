package gormstore

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

func (s Store) CreateImportedAsset(ctx context.Context, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation, promotedParent *asset.Asset, parentAuditRecord *audit.Record, link ports.ImportSourceLink, record ports.ImportJobResource) error {
	if err := validateImportSourceLink(link); err != nil {
		return err
	}
	if err := validateImportJobResource(record); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if promotedParent != nil && parentAuditRecord != nil {
			if err := promoteAssetParentInTx(tx, *promotedParent, *parentAuditRecord); err != nil {
				return err
			}
		}
		if err := createAssetInTx(tx, item, auditRecord, undoableOperation); err != nil {
			return err
		}
		linkModel := importSourceLinkModelFromRecord(link)
		if err := tx.Create(&linkModel).Error; err != nil {
			return importLinkPersistenceError(err)
		}
		resourceModel := importJobResourceModelFromRecord(record)
		if err := tx.Create(&resourceModel).Error; err != nil {
			return importLinkPersistenceError(err)
		}
		return nil
	})
}

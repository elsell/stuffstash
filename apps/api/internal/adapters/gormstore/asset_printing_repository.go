package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

func (s Store) CreateAssetWithPrint(ctx context.Context, in ports.PreparedAssetPrint) (result ports.AssetPrintResult, err error) {
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		job := in.Job.Job
		if _, err := printerByScope(tx, job.Scope, job.PrinterID); err != nil {
			return err
		}
		store := NewStore(tx)
		if previous, e := printRequest(tx, job); e == nil {
			if previous.RequestFingerprint != in.Job.RequestFingerprint {
				return ports.ErrConflict
			}
			saved, e := previous.domain()
			if e != nil {
				return e
			}
			item, found, e := store.AssetByID(ctx, tenant.ID(saved.Scope.TenantID), inventory.InventoryID(saved.Scope.InventoryID), asset.ID(saved.AssetID))
			if e != nil {
				return e
			}
			if !found {
				return ports.ErrPrintJobNotFound
			}
			result = ports.AssetPrintResult{Asset: item, Job: saved}
			return nil
		} else if !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		item := in.Asset.Asset
		if item.TenantID.String() != job.Scope.TenantID || item.InventoryID.String() != job.Scope.InventoryID || item.ID.String() != job.AssetID || in.Label.AssetID != job.AssetID || in.Label.TenantID != job.Scope.TenantID || in.Label.InventoryID != job.Scope.InventoryID || job.LabelReference != string(in.Label.ID) {
			return ports.ErrConflict
		}
		var operation *ports.UndoableOperation
		if in.Asset.UndoableOperation.ID != "" {
			operation = &in.Asset.UndoableOperation
		}
		if in.Asset.PromotedParent != nil {
			if in.Asset.ParentPromotionRecord == nil {
				return ports.ErrConflict
			}
			if err := promoteAssetParentInTx(tx, *in.Asset.PromotedParent, *in.Asset.ParentPromotionRecord); err != nil {
				return err
			}
		}
		if err := createAssetInTx(tx, item, in.Asset.AuditRecord, operation); err != nil {
			return err
		}
		if in.Asset.TagAudit != nil {
			if err := store.SetAssetTags(ctx, tenant.ID(job.Scope.TenantID), inventory.InventoryID(job.Scope.InventoryID), item.ID, in.Asset.TagIDs, *in.Asset.TagAudit); err != nil {
				return err
			}
		} else if len(in.Asset.TagIDs) > 0 {
			return ports.ErrConflict
		}
		if _, created, err := store.ProvisionLabel(ctx, in.Label, in.LabelAudit); err != nil {
			return err
		} else if !created {
			return ports.ErrConflict
		}
		saved, created, err := store.CreatePrintJob(ctx, in.Job)
		if err != nil {
			return err
		}
		if !created {
			return ports.ErrConflict
		}
		result = ports.AssetPrintResult{Asset: item, Job: saved, Created: true}
		return nil
	})
	if err != nil {
		return ports.AssetPrintResult{}, err
	}
	return result, nil
}

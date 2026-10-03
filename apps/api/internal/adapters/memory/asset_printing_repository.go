package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"maps"
)

func (s *Store) CreateAssetWithPrint(ctx context.Context, in ports.PreparedAssetPrint) (ports.AssetPrintResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// A private transaction view shares read-only catalogs and copies each map
	// changed by creation. Nothing is published until every operation succeeds.
	tx := &Store{inventories: s.inventories, customAssetTypes: s.customAssetTypes, assetTags: s.assetTags, labelInstance: s.labelInstance, printingPrinters: s.printingPrinters,
		assets: maps.Clone(s.assets), auditRecords: maps.Clone(s.auditRecords), undoables: maps.Clone(s.undoables), assetTagLinks: maps.Clone(s.assetTagLinks), labels: maps.Clone(s.labels),
		printingJobs: maps.Clone(s.printingJobs), printingJobContents: maps.Clone(s.printingJobContents), printingJobFingerprints: maps.Clone(s.printingJobFingerprints)}
	j := in.Job.Job
	for _, previous := range tx.printingJobs {
		if previous.Scope == j.Scope && previous.RequestedBy == j.RequestedBy && previous.IdempotencyKey == j.IdempotencyKey {
			if tx.printingJobFingerprints[previous.ID] != in.Job.RequestFingerprint {
				return ports.AssetPrintResult{}, ports.ErrConflict
			}
			item, found, err := tx.AssetByID(ctx, tenant.ID(j.Scope.TenantID), inventory.InventoryID(j.Scope.InventoryID), asset.ID(previous.AssetID))
			if err != nil {
				return ports.AssetPrintResult{}, err
			}
			if !found {
				return ports.AssetPrintResult{}, ports.ErrPrintJobNotFound
			}
			return ports.AssetPrintResult{Asset: item, Job: clonePrintJob(previous)}, nil
		}
	}
	item := in.Asset.Asset
	if item.TenantID.String() != j.Scope.TenantID || item.InventoryID.String() != j.Scope.InventoryID || item.ID.String() != j.AssetID || in.Label.AssetID != j.AssetID || in.Label.TenantID != j.Scope.TenantID || in.Label.InventoryID != j.Scope.InventoryID || j.LabelReference != string(in.Label.ID) {
		return ports.AssetPrintResult{}, ports.ErrConflict
	}
	var operation *ports.UndoableOperation
	if in.Asset.UndoableOperation.ID != "" {
		operation = &in.Asset.UndoableOperation
	}
	var err error
	if in.Asset.PromotedParent != nil {
		if in.Asset.ParentPromotionRecord == nil {
			return ports.AssetPrintResult{}, ports.ErrConflict
		}
		err = tx.CreateAssetWithParentPromotion(ctx, *in.Asset.PromotedParent, *in.Asset.ParentPromotionRecord, item, in.Asset.AuditRecord, operation)
	} else {
		err = tx.CreateAsset(ctx, item, in.Asset.AuditRecord, operation)
	}
	if err != nil {
		return ports.AssetPrintResult{}, err
	}
	if in.Asset.TagAudit != nil {
		if err = tx.SetAssetTags(ctx, tenant.ID(j.Scope.TenantID), inventory.InventoryID(j.Scope.InventoryID), item.ID, in.Asset.TagIDs, *in.Asset.TagAudit); err != nil {
			return ports.AssetPrintResult{}, err
		}
	} else if len(in.Asset.TagIDs) > 0 {
		return ports.AssetPrintResult{}, ports.ErrConflict
	}
	if _, created, err := tx.ProvisionLabel(ctx, in.Label, in.LabelAudit); err != nil {
		return ports.AssetPrintResult{}, err
	} else if !created {
		return ports.AssetPrintResult{}, ports.ErrConflict
	}
	saved, created, err := tx.CreatePrintJob(ctx, in.Job)
	if err != nil {
		return ports.AssetPrintResult{}, err
	}
	if !created {
		return ports.AssetPrintResult{}, ports.ErrConflict
	}
	s.assets = tx.assets
	s.auditRecords = tx.auditRecords
	s.undoables = tx.undoables
	s.assetTagLinks = tx.assetTagLinks
	s.labels = tx.labels
	s.printingJobs = tx.printingJobs
	s.printingJobContents = tx.printingJobContents
	s.printingJobFingerprints = tx.printingJobFingerprints
	return ports.AssetPrintResult{Asset: item, Job: saved, Created: true}, nil
}

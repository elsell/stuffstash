package dataportability

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a ImportService) createImportedAsset(ctx context.Context, command ports.ImportJobCommand, jobID importjob.ID, sourceIdentity ImportSourceIdentity, planned importplan.Asset, sourceToAssetID map[string]string, duplicates map[string]struct{}, tagIDsByKey map[string]string) (asset.Asset, bool, error) {
	if a.deps.ImportAssetUnitOfWork == nil {
		return asset.Asset{}, false, apperrors.ErrInvalidInput
	}
	if planned.Archived {
		return asset.Asset{}, true, nil
	}
	if link, found, err := a.deps.ImportLinks.ImportSourceLinkByKey(ctx, importAssetSourceLinkKey(command.TenantID, command.InventoryID, sourceIdentity, planned)); err != nil {
		return asset.Asset{}, false, err
	} else if found {
		if link.ResourceType == ports.ImportResourceAsset && strings.TrimSpace(link.ResourceID) != "" {
			sourceToAssetID[planned.SourceID] = link.ResourceID
		}
		a.recordImportSourceLinkDuplicateSkipped(ctx, command, ports.ImportSourceEntityAsset, jobID)
		return asset.Asset{}, true, nil
	}
	for _, key := range []string{"homebox-source-id", "homebox-asset-id"} {
		if homeboxID, ok := planned.CustomFields[key].(string); ok && strings.TrimSpace(homeboxID) != "" {
			if _, duplicate := duplicates[key+"="+strings.TrimSpace(homeboxID)]; duplicate {
				return asset.Asset{}, true, nil
			}
		}
	}
	parentAssetID := ""
	if planned.ParentSourceID != "" {
		parentAssetID = sourceToAssetID[planned.ParentSourceID]
		if parentAssetID == "" {
			return asset.Asset{}, true, nil
		}
	}
	prepared, err := a.deps.Targets.PrepareAsset(ctx, command, planned, parentAssetID)
	if err != nil {
		return asset.Asset{}, false, err
	}
	link, record, err := a.ImportedResourceRecords(ImportedResourceInput{
		TenantID:         command.TenantID,
		InventoryID:      command.InventoryID,
		JobID:            jobID,
		SourceIdentity:   sourceIdentity,
		SourceEntityType: ports.ImportSourceEntityAsset,
		SourceEntityID:   planned.SourceID,
		ResourceType:     ports.ImportResourceAsset,
		ResourceID:       prepared.Asset.ID.String(),
		CreatedAt:        a.deps.Clock.Now().UTC(),
	})
	if err != nil {
		return asset.Asset{}, false, err
	}
	if err := a.deps.ImportAssetUnitOfWork.CreateImportedAsset(ctx, prepared.Asset, prepared.AuditRecord, &prepared.UndoableOperation, prepared.PromotedParent, prepared.ParentPromotionRecord, link, record); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return asset.Asset{}, true, nil
		}
		return asset.Asset{}, false, err
	}
	tagIDs := plannedImportTagIDs(planned.TagKeys, tagIDsByKey)
	if len(tagIDs) > 0 {
		if err := a.deps.Targets.SetAssetTags(ctx, command, prepared.Asset.ID, tagIDs); err != nil {
			return asset.Asset{}, false, err
		}
	}
	a.deps.Targets.RecordAssetCreated(ctx, prepared.Asset, command.Principal.ID)
	for _, key := range []string{"homebox-source-id", "homebox-asset-id"} {
		if value, ok := planned.CustomFields[key].(string); ok && strings.TrimSpace(value) != "" {
			duplicates[key+"="+strings.TrimSpace(value)] = struct{}{}
		}
	}
	return prepared.Asset, false, nil
}

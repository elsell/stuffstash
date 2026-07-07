package app

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) applyImportTags(ctx context.Context, command ports.ImportJobCommand, plan importplan.Plan, result *ImportResult) (map[string]string, error) {
	tagIDsByKey := map[string]string{}
	if len(plan.Tags) == 0 {
		return tagIDsByKey, nil
	}
	existing, err := a.existingImportTagIDs(ctx, command.TenantID, command.InventoryID)
	if err != nil {
		return nil, err
	}
	total := len(plan.Tags)
	if err := a.updateImportProgress(ctx, command, importjob.PhaseTags, 0, total, "Creating tags"); err != nil {
		return nil, err
	}
	for index, tag := range plan.Tags {
		if err := a.stopIfImportCancelled(ctx, command); err != nil {
			return nil, err
		}
		if existingID := existing[tag.Key]; existingID != "" {
			tagIDsByKey[tag.Key] = existingID
			result.Counts.TagsExisting++
			if err := a.updateImportProgress(ctx, command, importjob.PhaseTags, index+1, total, "Creating tags"); err != nil {
				return nil, err
			}
			continue
		}
		created, err := a.CreateAssetTag(ctx, CreateAssetTagInput{
			Principal:   command.Principal,
			Source:      audit.SourceImport,
			RequestID:   command.RequestID,
			TenantID:    command.TenantID,
			InventoryID: command.InventoryID,
			Key:         tag.Key,
			DisplayName: tag.DisplayName,
			Color:       tag.Color,
		})
		if err != nil {
			if errors.Is(err, ErrInvalidInput) {
				tagID, found, findErr := a.activeImportTagIDByKey(ctx, command.TenantID, command.InventoryID, tag.Key)
				if findErr != nil {
					return nil, findErr
				}
				if !found {
					return nil, err
				}
				tagIDsByKey[tag.Key] = tagID
				result.Counts.TagsExisting++
				if err := a.updateImportProgress(ctx, command, importjob.PhaseTags, index+1, total, "Creating tags"); err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}
		tagIDsByKey[tag.Key] = created.ID.String()
		existing[tag.Key] = created.ID.String()
		result.Counts.TagsCreated++
		if err := a.updateImportProgress(ctx, command, importjob.PhaseTags, index+1, total, "Creating tags"); err != nil {
			return nil, err
		}
	}
	return tagIDsByKey, nil
}

func (a App) activeImportTagIDByKey(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, key string) (string, bool, error) {
	if a.assetTags == nil {
		return "", false, nil
	}
	parsedKey, ok := assettag.NewKey(key)
	if !ok {
		return "", false, nil
	}
	tag, found, err := a.assetTags.AssetTagByKey(ctx, tenantID, inventoryID, parsedKey)
	if err != nil || !found {
		return "", false, err
	}
	if tag.LifecycleState != assettag.LifecycleStateActive {
		return "", false, nil
	}
	return tag.ID.String(), true, nil
}

func (a App) existingImportTagIDs(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (map[string]string, error) {
	tags := map[string]string{}
	if a.assetTags == nil {
		return tags, nil
	}
	items, err := a.assetTags.ListAssetTags(ctx, tenantID, inventoryID, ports.AssetTagPageRequest{Limit: 10000})
	if err != nil {
		return nil, err
	}
	for _, tag := range items {
		tags[tag.Key.String()] = tag.ID.String()
	}
	return tags, nil
}

func plannedImportTagIDs(tagKeys []string, tagIDsByKey map[string]string) []string {
	if len(tagKeys) == 0 {
		return nil
	}
	tagIDs := make([]string, 0, len(tagKeys))
	seen := map[string]struct{}{}
	for _, key := range tagKeys {
		tagID := tagIDsByKey[strings.TrimSpace(key)]
		if tagID == "" {
			continue
		}
		if _, ok := seen[tagID]; ok {
			continue
		}
		seen[tagID] = struct{}{}
		tagIDs = append(tagIDs, tagID)
	}
	return tagIDs
}

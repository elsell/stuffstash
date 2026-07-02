package app

import (
	"context"
	"encoding/base64"
	"errors"
	"sort"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const maxImportCSVBytes = 10 * 1024 * 1024

type ImportSourceInput struct {
	SourceType          string
	BaseURL             string
	Username            string
	Password            string
	IncludeImages       bool
	AllowInsecureTLS    bool
	AllowPrivateNetwork bool
	FileName            string
	ContentBase64       string
}

type PreviewImportInput struct {
	Principal   identity.Principal
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Source      ImportSourceInput
}

type ApplyImportInput = PreviewImportInput

type ImportPreview struct {
	Plan importplan.Plan
}

type ImportResult struct {
	Counts   ImportApplyCounts
	Messages []importplan.Message
}

type ImportApplyCounts struct {
	FieldsCreated      int
	FieldsExisting     int
	LocationsCreated   int
	AssetsCreated      int
	AssetsSkipped      int
	AttachmentsCreated int
	AttachmentsSkipped int
}

func (a App) PreviewLegacyHomeboxImport(ctx context.Context, input PreviewImportInput) (ImportPreview, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return ImportPreview{}, err
	}
	plan, err := a.readImportSource(ctx, input.Source)
	if err != nil {
		return ImportPreview{}, ErrInvalidInput
	}
	plan.Messages = append(plan.Messages, a.duplicateWarnings(ctx, input.TenantID, input.InventoryID, plan)...)
	plan.Messages = append(plan.Messages, archivedWarnings(plan)...)
	stripAttachmentContent(&plan)
	return ImportPreview{Plan: plan}, nil
}

func (a App) ApplyLegacyHomeboxImport(ctx context.Context, input ApplyImportInput) (ImportResult, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return ImportResult{}, err
	}
	plan, err := a.readImportSource(ctx, input.Source)
	if err != nil {
		return ImportResult{}, ErrInvalidInput
	}
	result := ImportResult{}
	result.Messages = append(result.Messages, plan.Messages...)
	result.Messages = append(result.Messages, archivedWarnings(plan)...)
	if plan.Counts().Errors > 0 {
		return result, ErrInvalidInput
	}

	existingFieldKeys, err := a.existingFieldKeys(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return ImportResult{}, err
	}
	for _, field := range plan.Fields {
		if _, ok := existingFieldKeys[field.Key]; ok {
			result.Counts.FieldsExisting++
			continue
		}
		_, err := a.CreateInventoryCustomFieldDefinition(ctx, CreateCustomFieldDefinitionInput{
			Principal:     input.Principal,
			Source:        audit.SourceImport,
			RequestID:     input.RequestID,
			TenantID:      input.TenantID,
			InventoryID:   input.InventoryID,
			Key:           field.Key,
			DisplayName:   field.DisplayName,
			Type:          field.Type,
			Applicability: customfield.ApplicabilityAllAssets.String(),
		})
		if err != nil {
			if errors.Is(err, ErrInvalidInput) {
				result.Counts.FieldsExisting++
				continue
			}
			return result, err
		}
		result.Counts.FieldsCreated++
	}

	duplicates, err := a.existingHomeboxReferences(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return ImportResult{}, err
	}
	sourceToAssetID := map[string]string{}
	for _, planned := range sortedImportAssets(plan.Assets, "location") {
		created, skipped, err := a.createImportedAsset(ctx, input, planned, sourceToAssetID, duplicates)
		if err != nil {
			return result, err
		}
		if skipped {
			result.Counts.AssetsSkipped++
			continue
		}
		sourceToAssetID[planned.SourceID] = created.ID.String()
		result.Counts.LocationsCreated++
	}
	for _, planned := range sortedImportAssets(plan.Assets, "item") {
		created, skipped, err := a.createImportedAsset(ctx, input, planned, sourceToAssetID, duplicates)
		if err != nil {
			return result, err
		}
		if skipped {
			result.Counts.AssetsSkipped++
			continue
		}
		sourceToAssetID[planned.SourceID] = created.ID.String()
		result.Counts.AssetsCreated++
	}

	for _, attachment := range plan.Attachments {
		assetID, ok := sourceToAssetID[attachment.AssetSourceID]
		if !ok {
			result.Counts.AttachmentsSkipped++
			continue
		}
		parsedAssetID, ok := asset.NewID(assetID)
		if !ok {
			result.Counts.AttachmentsSkipped++
			continue
		}
		_, err := a.CreateAttachment(ctx, CreateAttachmentInput{
			Principal:   input.Principal,
			Source:      audit.SourceImport,
			RequestID:   input.RequestID,
			TenantID:    input.TenantID,
			InventoryID: input.InventoryID,
			AssetID:     parsedAssetID,
			FileName:    attachment.FileName,
			ContentType: attachment.ContentType,
			Content:     attachment.Content,
		})
		if err != nil {
			result.Counts.AttachmentsSkipped++
			result.Messages = append(result.Messages, importplan.Message{
				Code:       "attachment-skipped",
				Severity:   importplan.SeverityWarning,
				Summary:    "Attachment could not be imported",
				Detail:     safeImportError(err),
				SourceID:   attachment.SourceID,
				SourceName: attachment.FileName,
			})
			continue
		}
		result.Counts.AttachmentsCreated++
	}
	return result, nil
}

func (a App) readImportSource(ctx context.Context, input ImportSourceInput) (importplan.Plan, error) {
	if a.importSources == nil {
		return importplan.Plan{}, ErrInvalidInput
	}
	var content []byte
	if strings.TrimSpace(input.ContentBase64) != "" {
		if base64.StdEncoding.DecodedLen(len(strings.TrimSpace(input.ContentBase64))) > maxImportCSVBytes {
			return importplan.Plan{}, ErrInvalidInput
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(input.ContentBase64))
		if err != nil {
			return importplan.Plan{}, ErrInvalidInput
		}
		if len(decoded) > maxImportCSVBytes {
			return importplan.Plan{}, ErrInvalidInput
		}
		content = decoded
	}
	return a.importSources.ReadImportPlan(ctx, ports.ImportSourceRequest{
		SourceType:          importplan.SourceType(input.SourceType),
		BaseURL:             input.BaseURL,
		Username:            input.Username,
		Password:            input.Password,
		IncludeImages:       input.IncludeImages,
		AllowInsecureTLS:    input.AllowInsecureTLS,
		AllowPrivateNetwork: input.AllowPrivateNetwork,
		MaxAttachmentBytes:  int64(a.maxAttachmentBytes),
		FileName:            input.FileName,
		Content:             content,
	})
}

func (a App) createImportedAsset(ctx context.Context, input ApplyImportInput, planned importplan.Asset, sourceToAssetID map[string]string, duplicates map[string]struct{}) (asset.Asset, bool, error) {
	if planned.Archived {
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
	created, err := a.CreateAsset(ctx, CreateAssetInput{
		Principal:     input.Principal,
		Source:        audit.SourceImport,
		RequestID:     input.RequestID,
		TenantID:      input.TenantID,
		InventoryID:   input.InventoryID,
		Kind:          planned.Kind,
		Title:         planned.Title,
		Description:   planned.Description,
		ParentAssetID: parentAssetID,
		CustomFields:  planned.CustomFields,
	})
	if err != nil {
		return asset.Asset{}, false, err
	}
	for _, key := range []string{"homebox-source-id", "homebox-asset-id"} {
		if value, ok := planned.CustomFields[key].(string); ok && strings.TrimSpace(value) != "" {
			duplicates[key+"="+strings.TrimSpace(value)] = struct{}{}
		}
	}
	return created, false, nil
}

func (a App) existingFieldKeys(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (map[string]struct{}, error) {
	keys := map[string]struct{}{}
	if a.customFields == nil {
		return keys, nil
	}
	fields, err := a.customFields.ListEffectiveCustomFieldDefinitions(ctx, tenantID, inventoryID)
	if err != nil {
		return nil, err
	}
	for _, field := range fields {
		keys[field.Key.String()] = struct{}{}
	}
	return keys, nil
}

func (a App) existingHomeboxReferences(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (map[string]struct{}, error) {
	ids := map[string]struct{}{}
	if a.assets == nil {
		return ids, nil
	}
	items, err := a.assets.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{
		Limit:           10000,
		LifecycleFilter: ports.AssetLifecycleFilterAll,
		Sort:            ports.AssetListSortIDAsc,
	})
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		for _, key := range []string{"homebox-source-id", "homebox-asset-id"} {
			if value, ok := item.CustomFields.Values()[key].(string); ok && strings.TrimSpace(value) != "" {
				ids[key+"="+strings.TrimSpace(value)] = struct{}{}
			}
		}
	}
	return ids, nil
}

func (a App) duplicateWarnings(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, plan importplan.Plan) []importplan.Message {
	duplicates, err := a.existingHomeboxReferences(ctx, tenantID, inventoryID)
	if err != nil {
		return nil
	}
	var messages []importplan.Message
	for _, planned := range plan.Assets {
		for _, key := range []string{"homebox-source-id", "homebox-asset-id"} {
			value, ok := planned.CustomFields[key].(string)
			if !ok || strings.TrimSpace(value) == "" {
				continue
			}
			if _, duplicate := duplicates[key+"="+strings.TrimSpace(value)]; duplicate {
				messages = append(messages, importplan.Message{
					Code:       "duplicate-source-asset",
					Severity:   importplan.SeverityWarning,
					Summary:    "Asset appears to have already been imported",
					Detail:     key + "=" + strings.TrimSpace(value),
					SourceID:   planned.SourceID,
					SourceName: planned.Title,
				})
				break
			}
		}
	}
	return messages
}

func archivedWarnings(plan importplan.Plan) []importplan.Message {
	var messages []importplan.Message
	for _, planned := range plan.Assets {
		if !planned.Archived {
			continue
		}
		messages = append(messages, importplan.Message{
			Code:       "archived-source-asset-skipped",
			Severity:   importplan.SeverityWarning,
			Summary:    "Archived Homebox asset will be skipped",
			Detail:     "archived source assets are not imported in this version",
			SourceID:   planned.SourceID,
			SourceName: planned.Title,
		})
	}
	return messages
}

func sortedImportAssets(items []importplan.Asset, kind string) []importplan.Asset {
	bySourceID := map[string]importplan.Asset{}
	children := map[string][]importplan.Asset{}
	for _, item := range items {
		if item.Kind == kind {
			bySourceID[item.SourceID] = item
			children[item.ParentSourceID] = append(children[item.ParentSourceID], item)
		}
	}
	for parentID := range children {
		sort.SliceStable(children[parentID], func(left, right int) bool {
			return children[parentID][left].Title < children[parentID][right].Title
		})
	}
	var sorted []importplan.Asset
	visited := map[string]struct{}{}
	var visit func(importplan.Asset)
	visit = func(item importplan.Asset) {
		if _, ok := visited[item.SourceID]; ok {
			return
		}
		if parent, ok := bySourceID[item.ParentSourceID]; ok {
			visit(parent)
		}
		visited[item.SourceID] = struct{}{}
		sorted = append(sorted, item)
		for _, child := range children[item.SourceID] {
			visit(child)
		}
	}
	for _, root := range children[""] {
		visit(root)
	}
	var remaining []importplan.Asset
	for sourceID, item := range bySourceID {
		if _, ok := visited[sourceID]; !ok {
			remaining = append(remaining, item)
		}
	}
	sort.SliceStable(remaining, func(left, right int) bool {
		return remaining[left].Title < remaining[right].Title
	})
	for _, item := range remaining {
		visit(item)
	}
	return sorted
}

func stripAttachmentContent(plan *importplan.Plan) {
	for index := range plan.Attachments {
		plan.Attachments[index].Content = nil
	}
}

func safeImportError(err error) string {
	switch {
	case errors.Is(err, ErrAttachmentTooLarge):
		return "attachment is too large"
	case errors.Is(err, ErrAttachmentContentTypeUnsupported):
		return "attachment file type is unsupported"
	case errors.Is(err, ErrAttachmentContentMismatch):
		return "attachment content did not match its file type"
	default:
		return "import validation failed"
	}
}

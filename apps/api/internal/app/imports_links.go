package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type importSourceIdentity struct {
	sourceType        importplan.SourceType
	sourceInstanceKey string
}

type importImportedResourceInput struct {
	TenantID         tenant.ID
	InventoryID      inventory.InventoryID
	JobID            importjob.ID
	SourceIdentity   importSourceIdentity
	SourceEntityType ports.ImportSourceEntityType
	SourceEntityID   string
	ResourceType     ports.ImportResourceType
	ResourceID       string
	ResourceOwnerID  string
	CreatedAt        time.Time
}

func importSourceIdentityForJob(source importjob.SourceRef) (importSourceIdentity, error) {
	instanceKey := strings.TrimSpace(source.BaseURL)
	sourceType := importplan.SourceType(source.Type)
	switch sourceType {
	case importplan.SourceLegacyHomebox:
		if instanceKey == "" {
			return importSourceIdentity{}, ErrInvalidInput
		}
	case importplan.SourceLegacyHomeboxCSV:
		instanceKey = strings.TrimSpace(source.Fingerprint)
		if instanceKey == "" {
			return importSourceIdentity{}, ErrInvalidInput
		}
	default:
		return importSourceIdentity{}, ErrInvalidInput
	}
	return importSourceIdentity{sourceType: sourceType, sourceInstanceKey: instanceKey}, nil
}

func importAssetSourceLinkKey(tenantID tenant.ID, inventoryID inventory.InventoryID, sourceIdentity importSourceIdentity, planned importplan.Asset) ports.ImportSourceLinkKey {
	return ports.ImportSourceLinkKey{
		TenantID:          tenantID,
		InventoryID:       inventoryID,
		SourceType:        sourceIdentity.sourceType,
		SourceInstanceKey: sourceIdentity.sourceInstanceKey,
		SourceEntityType:  ports.ImportSourceEntityAsset,
		SourceEntityID:    strings.TrimSpace(planned.SourceID),
	}
}

func importAttachmentSourceLinkKey(tenantID tenant.ID, inventoryID inventory.InventoryID, sourceIdentity importSourceIdentity, planned importplan.Attachment) ports.ImportSourceLinkKey {
	return ports.ImportSourceLinkKey{
		TenantID:          tenantID,
		InventoryID:       inventoryID,
		SourceType:        sourceIdentity.sourceType,
		SourceInstanceKey: sourceIdentity.sourceInstanceKey,
		SourceEntityType:  ports.ImportSourceEntityAttachment,
		SourceEntityID:    strings.TrimSpace(planned.SourceID),
	}
}

func (a App) recordImportedResource(ctx context.Context, input importImportedResourceInput) error {
	if a.importLinks == nil {
		return ErrInvalidInput
	}
	link, record, err := a.importedResourceRecords(input)
	if err != nil {
		return err
	}
	if err := a.importLinks.SaveImportSourceLink(ctx, link); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return ErrPrecondition
		}
		return err
	}
	if err := a.importLinks.SaveImportJobResource(ctx, record); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return nil
		}
		return err
	}
	return nil
}

func (a App) importedResourceRecords(input importImportedResourceInput) (ports.ImportSourceLink, ports.ImportJobResource, error) {
	if input.CreatedAt.IsZero() {
		return ports.ImportSourceLink{}, ports.ImportJobResource{}, ErrInvalidInput
	}
	key := ports.ImportSourceLinkKey{
		TenantID:          input.TenantID,
		InventoryID:       input.InventoryID,
		SourceType:        input.SourceIdentity.sourceType,
		SourceInstanceKey: input.SourceIdentity.sourceInstanceKey,
		SourceEntityType:  input.SourceEntityType,
		SourceEntityID:    strings.TrimSpace(input.SourceEntityID),
	}
	link := ports.ImportSourceLink{
		Key:          key,
		ResourceType: input.ResourceType,
		ResourceID:   strings.TrimSpace(input.ResourceID),
		JobID:        input.JobID,
		CreatedAt:    input.CreatedAt.UTC(),
	}
	record := ports.ImportJobResource{
		TenantID:          input.TenantID,
		InventoryID:       input.InventoryID,
		JobID:             input.JobID,
		ResourceType:      input.ResourceType,
		ResourceID:        strings.TrimSpace(input.ResourceID),
		ResourceOwnerID:   strings.TrimSpace(input.ResourceOwnerID),
		SourceType:        input.SourceIdentity.sourceType,
		SourceInstanceKey: input.SourceIdentity.sourceInstanceKey,
		SourceEntityType:  input.SourceEntityType,
		SourceEntityID:    strings.TrimSpace(input.SourceEntityID),
		CreatedAt:         input.CreatedAt.UTC(),
	}
	return link, record, nil
}

func (a App) sourceLinkDuplicateWarnings(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, source importjob.SourceRef, plan importplan.Plan) ([]importplan.Message, map[string]struct{}, error) {
	linkedAssetSourceIDs := map[string]struct{}{}
	if a.importLinks == nil {
		return nil, linkedAssetSourceIDs, nil
	}
	sourceIdentity, err := importSourceIdentityForJob(source)
	if err != nil {
		return nil, linkedAssetSourceIDs, nil
	}
	var messages []importplan.Message
	for _, planned := range plan.Assets {
		link, found, err := a.importLinks.ImportSourceLinkByKey(ctx, importAssetSourceLinkKey(tenantID, inventoryID, sourceIdentity, planned))
		if err != nil {
			return nil, linkedAssetSourceIDs, err
		}
		if !found {
			continue
		}
		if link.ResourceType != ports.ImportResourceAsset || strings.TrimSpace(link.ResourceID) == "" {
			continue
		}
		linkedAssetSourceIDs[planned.SourceID] = struct{}{}
		messages = append(messages, importplan.Message{
			Code:       "duplicate-source-asset",
			Severity:   importplan.SeverityWarning,
			Summary:    "Asset appears to have already been imported",
			Detail:     "source link already exists",
			SourceID:   planned.SourceID,
			SourceName: planned.Title,
		})
	}
	for _, planned := range plan.Attachments {
		link, found, err := a.importLinks.ImportSourceLinkByKey(ctx, importAttachmentSourceLinkKey(tenantID, inventoryID, sourceIdentity, planned))
		if err != nil {
			return nil, linkedAssetSourceIDs, err
		}
		if !found {
			continue
		}
		if link.ResourceType != ports.ImportResourceAttachment || strings.TrimSpace(link.ResourceID) == "" {
			continue
		}
		messages = append(messages, importplan.Message{
			Code:       "duplicate-source-attachment",
			Severity:   importplan.SeverityWarning,
			Summary:    "Attachment appears to have already been imported",
			Detail:     "source link already exists",
			SourceID:   planned.SourceID,
			SourceName: planned.FileName,
		})
	}
	return messages, linkedAssetSourceIDs, nil
}

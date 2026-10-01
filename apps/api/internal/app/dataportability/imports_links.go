package dataportability

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ImportSourceIdentity struct {
	SourceType        importplan.SourceType
	SourceInstanceKey string
}

type ImportedResourceInput struct {
	TenantID         tenant.ID
	InventoryID      inventory.InventoryID
	JobID            importjob.ID
	SourceIdentity   ImportSourceIdentity
	SourceEntityType ports.ImportSourceEntityType
	SourceEntityID   string
	ResourceType     ports.ImportResourceType
	ResourceID       string
	ResourceOwnerID  string
	CreatedAt        time.Time
}

func importSourceIdentityForJob(source importjob.SourceRef) (ImportSourceIdentity, error) {
	instanceKey := strings.TrimSpace(source.BaseURL)
	sourceType := importplan.SourceType(source.Type)
	switch sourceType {
	case importplan.SourceLegacyHomebox:
		if instanceKey == "" {
			return ImportSourceIdentity{}, apperrors.ErrInvalidInput
		}
	case importplan.SourceLegacyHomeboxCSV:
		instanceKey = strings.TrimSpace(source.Fingerprint)
		if instanceKey == "" {
			return ImportSourceIdentity{}, apperrors.ErrInvalidInput
		}
	default:
		return ImportSourceIdentity{}, apperrors.ErrInvalidInput
	}
	return ImportSourceIdentity{SourceType: sourceType, SourceInstanceKey: instanceKey}, nil
}

func importAssetSourceLinkKey(tenantID tenant.ID, inventoryID inventory.InventoryID, sourceIdentity ImportSourceIdentity, planned importplan.Asset) ports.ImportSourceLinkKey {
	return ports.ImportSourceLinkKey{
		TenantID:          tenantID,
		InventoryID:       inventoryID,
		SourceType:        sourceIdentity.SourceType,
		SourceInstanceKey: sourceIdentity.SourceInstanceKey,
		SourceEntityType:  ports.ImportSourceEntityAsset,
		SourceEntityID:    strings.TrimSpace(planned.SourceID),
	}
}

func ImportAttachmentSourceLinkKey(tenantID tenant.ID, inventoryID inventory.InventoryID, sourceIdentity ImportSourceIdentity, planned importplan.Attachment) ports.ImportSourceLinkKey {
	return ports.ImportSourceLinkKey{
		TenantID:          tenantID,
		InventoryID:       inventoryID,
		SourceType:        sourceIdentity.SourceType,
		SourceInstanceKey: sourceIdentity.SourceInstanceKey,
		SourceEntityType:  ports.ImportSourceEntityAttachment,
		SourceEntityID:    strings.TrimSpace(planned.SourceID),
	}
}

func (a ImportService) recordImportedResource(ctx context.Context, input ImportedResourceInput) error {
	if a.deps.ImportLinks == nil {
		return apperrors.ErrInvalidInput
	}
	link, record, err := a.ImportedResourceRecords(input)
	if err != nil {
		return err
	}
	if err := a.deps.ImportLinks.SaveImportSourceLink(ctx, link); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return apperrors.ErrPrecondition
		}
		return err
	}
	if err := a.deps.ImportLinks.SaveImportJobResource(ctx, record); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return nil
		}
		return err
	}
	return nil
}

func (a ImportService) ImportedResourceRecords(input ImportedResourceInput) (ports.ImportSourceLink, ports.ImportJobResource, error) {
	if input.CreatedAt.IsZero() {
		return ports.ImportSourceLink{}, ports.ImportJobResource{}, apperrors.ErrInvalidInput
	}
	key := ports.ImportSourceLinkKey{
		TenantID:          input.TenantID,
		InventoryID:       input.InventoryID,
		SourceType:        input.SourceIdentity.SourceType,
		SourceInstanceKey: input.SourceIdentity.SourceInstanceKey,
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
		SourceType:        input.SourceIdentity.SourceType,
		SourceInstanceKey: input.SourceIdentity.SourceInstanceKey,
		SourceEntityType:  input.SourceEntityType,
		SourceEntityID:    strings.TrimSpace(input.SourceEntityID),
		CreatedAt:         input.CreatedAt.UTC(),
	}
	return link, record, nil
}

func (a ImportService) sourceLinkDuplicateWarnings(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, source importjob.SourceRef, plan importplan.Plan) ([]importplan.Message, map[string]struct{}, error) {
	linkedAssetSourceIDs := map[string]struct{}{}
	if a.deps.ImportLinks == nil {
		return nil, linkedAssetSourceIDs, nil
	}
	sourceIdentity, err := importSourceIdentityForJob(source)
	if err != nil {
		return nil, linkedAssetSourceIDs, nil
	}
	var messages []importplan.Message
	for _, planned := range plan.Assets {
		link, found, err := a.deps.ImportLinks.ImportSourceLinkByKey(ctx, importAssetSourceLinkKey(tenantID, inventoryID, sourceIdentity, planned))
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
		link, found, err := a.deps.ImportLinks.ImportSourceLinkByKey(ctx, ImportAttachmentSourceLinkKey(tenantID, inventoryID, sourceIdentity, planned))
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

package dataportability

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a ImportService) ensureImportJobViewAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return a.deps.Targets.EnsureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, ports.InventoryPermissionViewImportJob)
}

func (a ImportService) ensureImportJobCreateAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return a.deps.Targets.EnsureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, ports.InventoryPermissionCreateImportJob)
}

func (a ImportService) importJob(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (importjob.Record, error) {
	if a.deps.ImportJobs == nil || jobID.String() == "" {
		return importjob.Record{}, apperrors.ErrInvalidInput
	}
	job, ok, err := a.deps.ImportJobs.ImportJobByID(ctx, tenantID, inventoryID, jobID)
	if err != nil {
		return importjob.Record{}, err
	}
	if !ok {
		return importjob.Record{}, apperrors.ErrNotFound
	}
	return job, nil
}

func (a ImportService) withImportJobResources(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobs []importjob.Record) ([]importjob.Record, error) {
	if a.deps.ImportLinks == nil {
		return jobs, nil
	}
	out := make([]importjob.Record, 0, len(jobs))
	for _, job := range jobs {
		enriched, err := a.withImportJobResource(ctx, job)
		if err != nil {
			return nil, err
		}
		if importJobTenantID(enriched.TenantID) == tenantID && importJobInventoryID(enriched.InventoryID) == inventoryID {
			out = append(out, enriched)
		}
	}
	return out, nil
}

func (a ImportService) withImportJobResource(ctx context.Context, job importjob.Record) (importjob.Record, error) {
	if a.deps.ImportLinks == nil {
		return job, nil
	}
	if job.Status == importjob.StatusCancelledDiscarded {
		job.Resources = nil
		return job, nil
	}
	records, err := a.deps.ImportLinks.ListImportJobResources(ctx, importJobTenantID(job.TenantID), importJobInventoryID(job.InventoryID), job.ID, ports.ImportJobResourcePageRequest{Limit: maxImportJobResourceSummaries})
	if err != nil {
		return importjob.Record{}, err
	}
	job.Resources = make([]importjob.ResourceSummary, 0, len(records))
	for _, record := range records {
		displayName, err := a.importJobResourceDisplayName(ctx, record)
		if err != nil {
			return importjob.Record{}, err
		}
		job.Resources = append(job.Resources, importjob.ResourceSummary{
			ResourceType:     string(record.ResourceType),
			ResourceID:       strings.TrimSpace(record.ResourceID),
			DisplayName:      displayName,
			ResourceOwnerID:  strings.TrimSpace(record.ResourceOwnerID),
			SourceEntityType: string(record.SourceEntityType),
			SourceEntityID:   strings.TrimSpace(record.SourceEntityID),
			CreatedAt:        record.CreatedAt.UTC(),
		})
	}
	return job, nil
}

func (a ImportService) importJobResourceDisplayName(ctx context.Context, record ports.ImportJobResource) (string, error) {
	switch record.ResourceType {
	case ports.ImportResourceAsset:
		if a.deps.Assets == nil {
			return "", nil
		}
		assetID, ok := asset.NewID(strings.TrimSpace(record.ResourceID))
		if !ok {
			return "", nil
		}
		item, found, err := a.deps.Assets.AssetByID(ctx, record.TenantID, record.InventoryID, assetID)
		if err != nil || !found {
			return "", err
		}
		return item.Title.String(), nil
	case ports.ImportResourceAttachment:
		if a.deps.Attachments == nil {
			return "", nil
		}
		ownerID, ownerOK := asset.NewID(strings.TrimSpace(record.ResourceOwnerID))
		attachmentID, attachmentOK := media.NewID(strings.TrimSpace(record.ResourceID))
		if !ownerOK || !attachmentOK {
			return "", nil
		}
		attachment, found, err := a.deps.Attachments.AttachmentByID(ctx, record.TenantID, record.InventoryID, ownerID, attachmentID)
		if err != nil || !found {
			return "", err
		}
		return attachment.FileName.String(), nil
	default:
		return "", nil
	}
}

func SourceFingerprint(plan importplan.Plan) (string, error) {
	safe := struct {
		Source      importplan.SourceSummary
		Fields      []importplan.FieldDefinition
		Assets      []importplan.Asset
		Attachments []sourceFingerprintAttachment
	}{
		Source: plan.Source,
		Fields: append([]importplan.FieldDefinition{}, plan.Fields...),
		Assets: append([]importplan.Asset{}, plan.Assets...),
	}
	safe.Attachments = make([]sourceFingerprintAttachment, 0, len(plan.Attachments))
	for _, attachment := range plan.Attachments {
		safe.Attachments = append(safe.Attachments, sourceFingerprintAttachment{
			SourceID:      attachment.SourceID,
			AssetSourceID: attachment.AssetSourceID,
			Primary:       attachment.Primary,
		})
	}
	sort.SliceStable(safe.Attachments, func(left, right int) bool {
		if safe.Attachments[left].AssetSourceID != safe.Attachments[right].AssetSourceID {
			return safe.Attachments[left].AssetSourceID < safe.Attachments[right].AssetSourceID
		}
		if safe.Attachments[left].SourceID != safe.Attachments[right].SourceID {
			return safe.Attachments[left].SourceID < safe.Attachments[right].SourceID
		}
		return !safe.Attachments[left].Primary && safe.Attachments[right].Primary
	})
	payload, err := json.Marshal(safe)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return fmt.Sprintf("sha256:%x", sum), nil
}

type sourceFingerprintAttachment struct {
	SourceID      string
	AssetSourceID string
	Primary       bool
}

func importJobEventFields(job importjob.Record) map[string]string {
	return map[string]string{
		"tenant_id":    job.TenantID.String(),
		"inventory_id": job.InventoryID.String(),
		"job_id":       job.ID.String(),
		"source_type":  string(job.Source.Type),
		"status":       string(job.Status),
	}
}

func (a ImportService) recordImportProgressUpdated(ctx context.Context, job importjob.Record, progress importjob.Progress) {
	fields := importJobEventFields(job)
	fields["phase"] = string(progress.Phase)
	fields["done"] = fmt.Sprintf("%d", progress.Done)
	fields["total"] = fmt.Sprintf("%d", progress.Total)
	a.deps.Observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobProgressUpdated,
		Message: "Import job progress updated.",
		Fields:  fields,
	})
}

func (a ImportService) recordImportSourceLinkDuplicateSkipped(ctx context.Context, command ports.ImportJobCommand, entityType ports.ImportSourceEntityType, jobID importjob.ID) {
	a.deps.Observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobSourceLinkDuplicateSkipped,
		Message: "Import source link duplicate skipped.",
		Fields: map[string]string{
			"tenant_id":          command.TenantID.String(),
			"inventory_id":       command.InventoryID.String(),
			"job_id":             jobID.String(),
			"source_entity_type": string(entityType),
		},
	})
}

func (a ImportService) recordImportDiscardCleanupEvent(ctx context.Context, job importjob.Record, name ports.EventName, recordsDiscarded int, sourceLinksDiscarded int) {
	fields := importJobEventFields(job)
	fields["records_discarded"] = fmt.Sprintf("%d", recordsDiscarded)
	fields["source_links_discarded"] = fmt.Sprintf("%d", sourceLinksDiscarded)
	a.deps.Observer.Record(ctx, ports.Event{
		Name:    name,
		Message: "Import job discard cleanup updated.",
		Fields:  fields,
	})
}

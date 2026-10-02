package dataportability

import (
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type archiveAuditTarget struct {
	kind   audit.TargetType
	id     string
	action audit.Action
}

func archiveAuditTargets(d ports.InventoryExportDocument) []archiveAuditTarget {
	targets := []archiveAuditTarget{{audit.TargetInventory, d.InventoryID, audit.ActionInventoryCreated}}
	for _, t := range d.Tags {
		targets = append(targets, archiveAuditTarget{audit.TargetAssetTag, t.ID.String(), audit.ActionAssetTagCreated})
	}
	for _, t := range d.CustomAssetTypes {
		targets = append(targets, archiveAuditTarget{audit.TargetCustomAssetType, t.ID.String(), audit.ActionCustomAssetTypeCreated})
	}
	for _, f := range d.CustomFieldDefinitions {
		targets = append(targets, archiveAuditTarget{audit.TargetCustomFieldDefinition, f.ID.String(), audit.ActionCustomFieldDefinitionCreated})
	}
	for _, a := range d.Assets {
		targets = append(targets, archiveAuditTarget{audit.TargetAsset, a.ID, audit.ActionAssetCreated})
		for _, m := range a.Attachments {
			targets = append(targets, archiveAuditTarget{audit.TargetAttachment, m.ID.String(), audit.ActionAttachmentCreated})
		}
		if a.CurrentCheckout != nil {
			targets = append(targets, archiveAuditTarget{audit.TargetAsset, a.ID, audit.ActionAssetCheckedOut})
		}
	}
	return targets
}
func BuildArchiveRestoreAudits(plan ports.ArchiveRestorePlan, job archivejob.Record, ids ports.IDGenerator, clock ports.Clock) ([]audit.Record, error) {
	if ids == nil || clock == nil {
		return nil, ErrArchiveMetadata
	}
	d := plan.Document
	principals := map[string]string{}
	for _, a := range d.Assets {
		if c := a.CurrentCheckout; c != nil {
			principals[a.ID] = plan.CheckoutSourcePrincipals[c.ID.String()]
		}
	}
	records := []audit.Record{}
	targets := append(archiveAuditTargets(d), archiveAuditTarget{audit.TargetArchiveJob, job.ID, audit.ActionArchiveJobUpdated})
	for _, target := range targets {
		metadata := map[string]string{"archive_job_id": job.ID}
		if target.action == audit.ActionAssetCheckedOut {
			metadata["source_principal_id"] = principals[target.id]
		}
		inventoryID := audit.InventoryID(d.InventoryID)
		if target.kind == audit.TargetArchiveJob {
			inventoryID = ""
			metadata["state"] = string(archivejob.Queued)
			metadata["kind"] = string(archivejob.Restore)
		}
		record, ok := audit.NewRecord(audit.ID(ids.NewID()), audit.TenantID(d.TenantID), inventoryID, audit.PrincipalID(job.PrincipalID), target.action, audit.SourceImport, target.kind, target.id, clock.Now(), job.ID, metadata)
		if !ok {
			return nil, ErrArchiveMetadata
		}
		records = append(records, record)
	}
	return records, nil
}
func ValidateArchiveRestoreAudits(d ports.InventoryExportDocument, records []audit.Record, principal, jobID string) error {
	required := map[archiveAuditTarget]bool{}
	for _, target := range archiveAuditTargets(d) {
		required[target] = true
	}
	required[archiveAuditTarget{audit.TargetArchiveJob, jobID, audit.ActionArchiveJobUpdated}] = true
	seenIDs := map[string]bool{}
	for _, r := range records {
		target := archiveAuditTarget{r.TargetType, r.TargetID, r.Action}
		expectedInventory := d.InventoryID
		if r.TargetType == audit.TargetArchiveJob {
			expectedInventory = ""
		}
		if !required[target] || r.TenantID.String() != d.TenantID || r.InventoryID.String() != expectedInventory || r.PrincipalID.String() != principal || r.Source != audit.SourceImport || r.Metadata["archive_job_id"] != jobID || r.OccurredAt.IsZero() || r.ID.String() == "" || seenIDs[r.ID.String()] {
			return ErrArchiveMetadata
		}
		delete(required, target)
		seenIDs[r.ID.String()] = true
	}
	if len(required) != 0 {
		return ErrArchiveMetadata
	}
	return nil
}

package dataportability

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ArchivePreview = ports.ArchivePreview

func (s ArchiveService) Preview(ctx context.Context, a ArchiveAccess, id string) (ArchivePreview, error) {
	empty := ArchivePreview{}
	job, err := s.Job(ctx, a, id)
	if err != nil {
		return empty, err
	}
	if job.Kind != archivejob.Restore || job.PlanArtifactID == "" || !s.deps.Clock.Now().Before(job.ExpiresAt) || job.State == archivejob.Cancelled || job.State == archivejob.Expired {
		return empty, archivejob.ErrTransition
	}
	plan, err := s.loadRestorePlan(ctx, job)
	if err != nil {
		return empty, err
	}
	if err = s.auditArchiveRead(ctx, a, id, "preview"); err != nil {
		return empty, err
	}
	d := plan.Document
	result := ArchivePreview{InventoryName: d.InventoryName, Assets: len(d.Assets), Tags: len(d.Tags), CustomAssetTypes: len(d.CustomAssetTypes), CustomFields: len(d.CustomFieldDefinitions), OmittedAttachments: plan.OmittedAttachments, KeyRemappings: plan.KeyRemappings}
	for _, asset := range d.Assets {
		for _, attachment := range asset.Attachments {
			if attachment.ContentType.IsImage() {
				result.Photos++
			} else {
				result.OtherFiles++
			}
		}
	}
	return result, nil
}
func (s ArchiveService) Download(ctx context.Context, a ArchiveAccess, id string) (ports.BlobReadStream, int64, error) {
	job, err := s.Job(ctx, a, id)
	if err != nil {
		return nil, 0, err
	}
	if job.Kind != archivejob.Export || job.State != archivejob.Ready || !s.deps.Clock.Now().Before(job.ExpiresAt) {
		return nil, 0, archivejob.ErrTransition
	}
	key, ok := media.NewStorageKey(job.ResultArtifactID)
	if !ok {
		return nil, 0, archivejob.ErrInvalid
	}
	stream, size, err := s.deps.Storage.OpenBlobStream(ctx, key)
	if err != nil {
		return nil, 0, err
	}
	a.InventoryID = inventory.InventoryID(job.SourceInventoryID)
	if err = s.auditArchiveRead(ctx, a, id, "download"); err != nil {
		stream.Close()
		return nil, 0, err
	}
	return stream, size, nil
}
func (s ArchiveService) auditArchiveRead(ctx context.Context, a ArchiveAccess, id, operation string) error {
	if id == "" {
		id = a.InventoryID.String()
		if id == "" {
			id = a.TenantID.String()
		}
	}
	record, ok := audit.NewRecord(audit.ID(s.deps.IDs.NewID()), audit.TenantID(a.TenantID), audit.InventoryID(a.InventoryID), audit.PrincipalID(a.Principal.ID), audit.ActionArchiveJobViewed, audit.SourceAPI, audit.TargetArchiveJob, id, s.deps.Clock.Now(), "", map[string]string{"operation": operation})
	if !ok {
		return archivejob.ErrInvalid
	}
	return s.deps.Audit.SaveAuditRecord(ctx, record)
}

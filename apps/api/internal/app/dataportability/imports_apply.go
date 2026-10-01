package dataportability

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a ImportService) applyImportFields(ctx context.Context, command ports.ImportJobCommand, plan importplan.Plan, result *ImportResult) error {
	existingFieldKeys, err := a.existingFieldKeys(ctx, command.TenantID, command.InventoryID)
	if err != nil {
		return err
	}
	total := len(plan.Fields)
	if total > 0 {
		if err := a.UpdateImportProgress(ctx, command, importjob.PhaseFields, 0, total, "Creating custom fields"); err != nil {
			return err
		}
	}
	for index, field := range plan.Fields {
		if err := a.stopIfImportCancelled(ctx, command); err != nil {
			return err
		}
		if _, ok := existingFieldKeys[field.Key]; ok {
			result.Counts.FieldsExisting++
			if err := a.UpdateImportProgress(ctx, command, importjob.PhaseFields, index+1, total, "Creating custom fields"); err != nil {
				return err
			}
			continue
		}
		err := a.deps.Targets.CreateField(ctx, command, field)
		if err != nil {
			if errors.Is(err, apperrors.ErrInvalidInput) {
				result.Counts.FieldsExisting++
				if err := a.UpdateImportProgress(ctx, command, importjob.PhaseFields, index+1, total, "Creating custom fields"); err != nil {
					return err
				}
				continue
			}
			return err
		}
		result.Counts.FieldsCreated++
		if err := a.UpdateImportProgress(ctx, command, importjob.PhaseFields, index+1, total, "Creating custom fields"); err != nil {
			return err
		}
	}
	return nil
}

func (a ImportService) applyImportAssets(ctx context.Context, command ports.ImportJobCommand, jobID importjob.ID, sourceIdentity ImportSourceIdentity, plan importplan.Plan, duplicates map[string]struct{}, tagIDsByKey map[string]string, result *ImportResult) (map[string]string, error) {
	sourceToAssetID := map[string]string{}
	locations := sortedImportAssets(plan.Assets, "location")
	if len(locations) > 0 {
		if err := a.UpdateImportProgress(ctx, command, importjob.PhaseLocations, 0, len(locations), "Creating locations"); err != nil {
			return nil, err
		}
	}
	for index, planned := range locations {
		if err := a.stopIfImportCancelled(ctx, command); err != nil {
			return nil, err
		}
		created, skipped, err := a.createImportedAsset(ctx, command, jobID, sourceIdentity, planned, sourceToAssetID, duplicates, tagIDsByKey)
		if err != nil {
			return nil, err
		}
		if skipped {
			result.Counts.AssetsSkipped++
			if err := a.UpdateImportProgress(ctx, command, importjob.PhaseLocations, index+1, len(locations), "Creating locations"); err != nil {
				return nil, err
			}
			continue
		}
		sourceToAssetID[planned.SourceID] = created.ID.String()
		result.Counts.LocationsCreated++
		if err := a.UpdateImportProgress(ctx, command, importjob.PhaseLocations, index+1, len(locations), "Creating locations"); err != nil {
			return nil, err
		}
	}
	items := sortedNonLocationImportAssets(plan.Assets)
	if len(items) > 0 {
		if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAssets, 0, len(items), "Creating assets"); err != nil {
			return nil, err
		}
	}
	for index, planned := range items {
		if err := a.stopIfImportCancelled(ctx, command); err != nil {
			return nil, err
		}
		created, skipped, err := a.createImportedAsset(ctx, command, jobID, sourceIdentity, planned, sourceToAssetID, duplicates, tagIDsByKey)
		if err != nil {
			return nil, err
		}
		if skipped {
			result.Counts.AssetsSkipped++
			if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAssets, index+1, len(items), "Creating assets"); err != nil {
				return nil, err
			}
			continue
		}
		sourceToAssetID[planned.SourceID] = created.ID.String()
		result.Counts.AssetsCreated++
		if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAssets, index+1, len(items), "Creating assets"); err != nil {
			return nil, err
		}
	}
	return sourceToAssetID, nil
}

func (a ImportService) applyImportAttachments(ctx context.Context, command ports.ImportJobCommand, jobID importjob.ID, sourceIdentity ImportSourceIdentity, sourceRequest ports.ImportSourceRequest, plan importplan.Plan, sourceToAssetID map[string]string, result *ImportResult) error {
	total := len(plan.Attachments)
	var attachmentSession ports.ImportAttachmentSession
	if total > 0 {
		if a.deps.ImportAttachmentSources == nil {
			return apperrors.ErrInvalidInput
		}
		if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAttachments, 0, total, "Importing attachments"); err != nil {
			return err
		}
		var err error
		attachmentSession, err = a.deps.ImportAttachmentSources.OpenImportAttachmentSession(ctx, sourceRequest)
		if err != nil {
			message := ImportAttachmentSessionFailureMessage(err)
			result.Messages = append(result.Messages, message)
			return importAttachmentSessionStartError{detail: message.Detail}
		}
	}
	for index, attachment := range plan.Attachments {
		if err := a.stopIfImportCancelled(ctx, command); err != nil {
			return err
		}
		assetID, ok := sourceToAssetID[attachment.AssetSourceID]
		if !ok {
			result.Counts.AttachmentsSkipped++
			if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
				return err
			}
			continue
		}
		if link, found, err := a.deps.ImportLinks.ImportSourceLinkByKey(ctx, ImportAttachmentSourceLinkKey(command.TenantID, command.InventoryID, sourceIdentity, attachment)); err != nil {
			return err
		} else if found {
			if link.ResourceType == ports.ImportResourceAttachment && strings.TrimSpace(link.ResourceID) != "" {
				a.recordImportSourceLinkDuplicateSkipped(ctx, command, ports.ImportSourceEntityAttachment, jobID)
				result.Counts.AttachmentsSkipped++
				if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
					return err
				}
				continue
			}
		}
		parsedAssetID, ok := asset.NewID(assetID)
		if !ok {
			result.Counts.AttachmentsSkipped++
			if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
				return err
			}
			continue
		}
		if strings.TrimSpace(attachment.UnavailableReason) != "" {
			result.Counts.AttachmentsSkipped++
			result.Messages = append(result.Messages, importplan.Message{
				Code:       "attachment-unavailable",
				Severity:   importplan.SeverityWarning,
				Summary:    "Attachment could not be downloaded",
				Detail:     safeImportAttachmentUnavailableReason(attachment.UnavailableReason),
				SourceID:   attachment.SourceID,
				SourceName: attachment.FileName,
			})
			if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
				return err
			}
			continue
		}
		content, err := attachmentSession.ReadImportAttachment(ctx, attachment)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			result.Counts.AttachmentsSkipped++
			result.Messages = append(result.Messages, ImportAttachmentReadFailureMessage(err, attachment))
			if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
				return err
			}
			continue
		}
		attachment.FileName = content.FileName
		attachment.ContentType = content.ContentType
		attachment.Content = content.Content
		attachment.SizeBytes = len(content.Content)
		created, err := a.createImportedAttachment(ctx, command, jobID, sourceIdentity, parsedAssetID, attachment)
		if err != nil {
			var storageErr importAttachmentStorageError
			if errors.As(err, &storageErr) {
				result.Messages = append(result.Messages, importplan.Message{
					Code:       "attachment-storage-unavailable",
					Severity:   importplan.SeverityError,
					Summary:    "Image could not be saved",
					Detail:     storageErr.Error(),
					SourceID:   attachment.SourceID,
					SourceName: attachment.FileName,
				})
				return err
			}
			if errors.Is(err, ports.ErrConflict) {
				result.Counts.AttachmentsSkipped++
				if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
					return err
				}
				continue
			}
			if errors.Is(err, apperrors.ErrInvalidInput) || errors.Is(err, apperrors.ErrAttachmentFileNameInvalid) || errors.Is(err, apperrors.ErrAttachmentContentTypeUnsupported) || errors.Is(err, apperrors.ErrAttachmentContentMismatch) || errors.Is(err, apperrors.ErrAttachmentContentEmpty) || errors.Is(err, apperrors.ErrAttachmentTooLarge) {
				result.Counts.AttachmentsSkipped++
				result.Messages = append(result.Messages, importplan.Message{
					Code:       "attachment-skipped",
					Severity:   importplan.SeverityWarning,
					Summary:    "Attachment could not be imported",
					Detail:     safeImportError(err),
					SourceID:   attachment.SourceID,
					SourceName: attachment.FileName,
				})
				if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
					return err
				}
				continue
			}
			return err
		}
		if created.ID.String() == "" {
			result.Counts.AttachmentsSkipped++
			if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
				return err
			}
			continue
		}
		result.Counts.AttachmentsCreated++
		if err := a.UpdateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
			return err
		}
	}
	return nil
}

func (a ImportService) createImportedAttachment(ctx context.Context, command ports.ImportJobCommand, jobID importjob.ID, sourceIdentity ImportSourceIdentity, assetID asset.ID, planned importplan.Attachment) (media.Attachment, error) {
	if a.deps.ImportAttachmentUnitOfWork == nil {
		return media.Attachment{}, apperrors.ErrInvalidInput
	}
	prepared, err := a.deps.Targets.PrepareAttachment(ctx, command, assetID, planned)
	if err != nil {
		return media.Attachment{}, err
	}
	link, record, err := a.ImportedResourceRecords(ImportedResourceInput{
		TenantID:         command.TenantID,
		InventoryID:      command.InventoryID,
		JobID:            jobID,
		SourceIdentity:   sourceIdentity,
		SourceEntityType: ports.ImportSourceEntityAttachment,
		SourceEntityID:   planned.SourceID,
		ResourceType:     ports.ImportResourceAttachment,
		ResourceID:       prepared.Attachment.ID.String(),
		ResourceOwnerID:  assetID.String(),
		CreatedAt:        a.deps.Clock.Now().UTC(),
	})
	if err != nil {
		return media.Attachment{}, err
	}
	if err := a.deps.Blobs.PutBlob(ctx, prepared.StorageKey, prepared.ContentType, planned.Content); err != nil {
		a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob storage failed"})
		return media.Attachment{}, importAttachmentStorageError{}
	}
	if err := a.deps.ImportAttachmentUnitOfWork.CreateImportedAttachment(ctx, prepared.Attachment, prepared.AuditRecord, link, record, prepared.ThumbnailJob); err != nil {
		if deleteErr := a.deps.Blobs.DeleteBlob(ctx, prepared.StorageKey); deleteErr != nil {
			a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob cleanup failed"})
		}
		return media.Attachment{}, err
	}
	a.deps.Targets.RecordAttachmentCreated(ctx, command, assetID, prepared.Attachment)
	return prepared.Attachment, nil
}

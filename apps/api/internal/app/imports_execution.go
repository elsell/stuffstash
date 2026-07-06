package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type importCancelledError struct {
	mode importjob.CancellationMode
}

func (e importCancelledError) Error() string {
	return "import cancelled"
}

func (a App) ExecuteImportJob(ctx context.Context, command ports.ImportJobCommand) (importjob.Record, error) {
	job, err := a.importJob(ctx, command.TenantID, command.InventoryID, command.JobID)
	if err != nil {
		return importjob.Record{}, err
	}
	if job.Status != importjob.StatusRunning && job.Status != importjob.StatusCancelRequested && job.Status != importjob.StatusDiscardFailed {
		return importjob.Record{}, ErrPrecondition
	}
	expectedTerminalStatus := job.Status
	result := ImportResult{Counts: job.Counts}
	var applyErr error
	if job.Status == importjob.StatusCancelRequested {
		mode := job.CancellationMode
		if mode == "" {
			mode = importjob.CancellationModeKeepPartial
		}
		applyErr = importCancelledError{mode: mode}
	} else if job.Status == importjob.StatusDiscardFailed {
		applyErr = importCancelledError{mode: importjob.CancellationModeDiscardPartial}
	} else {
		result, applyErr = a.applyImportPlan(ctx, command, &job)
	}
	if latest, found, latestErr := a.importJobs.ImportJobByID(ctx, command.TenantID, command.InventoryID, command.JobID); latestErr != nil {
		return importjob.Record{}, latestErr
	} else if found {
		job.Progress = latest.Progress
		job.ProgressHistory = latest.ProgressHistory
		if latest.Status == importjob.StatusCancelRequested {
			job.CancellationMode = latest.CancellationMode
			job.CancellationRequestID = latest.CancellationRequestID
			expectedTerminalStatus = importjob.StatusCancelRequested
			if applyErr == nil {
				mode := latest.CancellationMode
				if mode == "" {
					mode = importjob.CancellationModeKeepPartial
				}
				applyErr = importCancelledError{mode: mode}
			}
		} else if latest.Status == importjob.StatusDiscardFailed {
			expectedTerminalStatus = importjob.StatusDiscardFailed
		}
	}
	now := a.clock.Now().UTC().Truncate(time.Microsecond)
	if !job.Progress.UpdatedAt.IsZero() && !now.After(job.Progress.UpdatedAt) {
		now = job.Progress.UpdatedAt.UTC().Truncate(time.Microsecond).Add(time.Microsecond)
	}
	job.Counts.FieldsCreated = result.Counts.FieldsCreated
	job.Counts.FieldsExisting = result.Counts.FieldsExisting
	job.Counts.LocationsCreated = result.Counts.LocationsCreated
	job.Counts.AssetsCreated = result.Counts.AssetsCreated
	job.Counts.AssetsSkipped = result.Counts.AssetsSkipped
	job.Counts.AttachmentsCreated = result.Counts.AttachmentsCreated
	job.Counts.AttachmentsSkipped = result.Counts.AttachmentsSkipped
	job.Messages = append(job.Messages, result.Messages...)
	job.CompletedAt = now
	job.UpdatedAt = now
	terminalDone := job.Progress.Done
	terminalTotal := job.Progress.Total
	if applyErr == nil {
		terminalDone = terminalTotal
	}
	job.Progress = importjob.Progress{Phase: importjob.PhaseTerminal, Done: terminalDone, Total: terminalTotal, UpdatedAt: now}
	var cancelled importCancelledError
	if errors.As(applyErr, &cancelled) {
		if cancelled.mode == importjob.CancellationModeDiscardPartial {
			discarded, links, discardErr := a.discardImportedJobResources(ctx, command, job.ID)
			job.Counts.RecordsDiscarded += discarded
			job.Counts.SourceLinksDiscarded += links
			if discardErr != nil {
				job.Status = importjob.StatusDiscardFailed
				job.Progress.Message = "Import cancellation cleanup failed"
				job.Messages = append(job.Messages, importplan.Message{
					Code:     "import-discard-failed",
					Severity: importplan.SeverityError,
					Summary:  "Import cancellation cleanup failed",
					Detail:   safeImportError(discardErr),
				})
			} else {
				job.Status = importjob.StatusCancelledDiscarded
				job.Progress.Message = "Import cancelled and partial progress discarded"
			}
		} else {
			job.Status = importjob.StatusCancelledKept
			job.Progress.Message = "Import cancelled and partial progress kept"
		}
	} else if applyErr != nil {
		job.Status = importjob.StatusFailed
		job.Progress.Message = "Import failed"
		job.Messages = append(job.Messages, importplan.Message{
			Code:     "import-failed",
			Severity: importplan.SeverityError,
			Summary:  "Import failed",
			Detail:   safeImportError(applyErr),
		})
	} else {
		job.Status = importjob.StatusSucceeded
		job.Progress.Message = "Import completed"
	}
	job.ProgressHistory = importjob.AppendProgressHistory(job.ProgressHistory, job.Progress)
	updated, err := a.importJobs.UpdateImportJobIfStatus(ctx, job, expectedTerminalStatus)
	if err != nil {
		return importjob.Record{}, err
	}
	if !updated {
		latest, found, latestErr := a.importJobs.ImportJobByID(ctx, command.TenantID, command.InventoryID, command.JobID)
		if latestErr != nil {
			return importjob.Record{}, latestErr
		}
		if found && latest.Status == importjob.StatusCancelRequested {
			return a.ExecuteImportJob(ctx, command)
		}
		if found && isTerminalImportJobStatus(latest.Status) {
			return latest, nil
		}
		return importjob.Record{}, ErrPrecondition
	}
	if job.Status == importjob.StatusCancelledDiscarded {
		a.recordImportDiscardCleanupEvent(ctx, job, ports.EventImportJobDiscardCleanupCompleted, job.Counts.RecordsDiscarded, job.Counts.SourceLinksDiscarded)
	} else if job.Status == importjob.StatusDiscardFailed {
		a.recordImportDiscardCleanupEvent(ctx, job, ports.EventImportJobDiscardCleanupFailed, job.Counts.RecordsDiscarded, job.Counts.SourceLinksDiscarded)
	}
	requestID := command.RequestID
	if job.CancellationRequestID != "" && (job.Status == importjob.StatusCancelledKept || job.Status == importjob.StatusCancelledDiscarded || job.Status == importjob.StatusDiscardFailed) {
		requestID = job.CancellationRequestID
	}
	if err := a.saveImportJobAuditRecord(ctx, command.Principal, requestID, job, importJobTerminalAuditAction(job), map[string]string{
		"records_discarded":      fmt.Sprintf("%d", job.Counts.RecordsDiscarded),
		"source_links_discarded": fmt.Sprintf("%d", job.Counts.SourceLinksDiscarded),
	}); err != nil {
		return importjob.Record{}, err
	}
	if a.importSourceVault != nil {
		scope := a.importJobSourceScope(command.TenantID, command.InventoryID, command.JobID)
		deleted, err := a.importSourceVault.DeleteImportJobSource(ctx, scope)
		if err != nil {
			job.Messages = append(job.Messages, importplan.Message{
				Code:     "import-source-cleanup-failed",
				Severity: importplan.SeverityWarning,
				Summary:  "Temporary import credentials could not be cleaned up automatically",
				Detail:   "credential cleanup will be retried by the import credential vacuum",
			})
			job.UpdatedAt = a.clock.Now().UTC()
			if updateErr := a.importJobs.UpdateImportJob(ctx, job); updateErr != nil {
				return importjob.Record{}, updateErr
			}
		} else if deleted {
			if err := a.saveImportJobCredentialCleanedAuditRecord(ctx, scope); err != nil {
				return importjob.Record{}, err
			}
		}
	}
	return job, nil
}

func (a App) applyImportPlan(ctx context.Context, command ports.ImportJobCommand, job *importjob.Record) (ImportResult, error) {
	if a.importSources == nil || a.importLinks == nil {
		return ImportResult{}, ErrInvalidInput
	}
	sourceRequest, err := a.importJobSourceRequest(ctx, command.TenantID, command.InventoryID, command.JobID)
	if err != nil {
		return ImportResult{}, err
	}
	plan, err := a.importSources.ReadImportPlan(ctx, sourceRequest)
	if err != nil {
		return ImportResult{}, importSourceInputError(err)
	}
	result := ImportResult{}
	checkPlan := a.normalizedImportPlanForJob(ctx, command.TenantID, command.InventoryID, plan)
	result.Messages = append(result.Messages, checkPlan.Messages...)
	fingerprint, err := sourceFingerprint(checkPlan)
	if err != nil {
		return result, err
	}
	if fingerprint != job.Source.Fingerprint {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventImportJobSourceFingerprintMismatch,
			Message: "Import source changed after preview.",
			Fields:  importJobEventFields(*job),
		})
		return result, ErrPrecondition
	}
	if plan.Counts().Errors > 0 {
		return result, ErrInvalidInput
	}
	sourceIdentity, err := importSourceIdentityForJob(job.Source)
	if err != nil {
		return result, err
	}
	if err := a.applyImportFields(ctx, command, plan, &result); err != nil {
		return result, err
	}
	duplicates, err := a.existingHomeboxReferences(ctx, command.TenantID, command.InventoryID)
	if err != nil {
		return result, err
	}
	sourceToAssetID, err := a.applyImportAssets(ctx, command, job.ID, sourceIdentity, plan, duplicates, &result)
	if err != nil {
		return result, err
	}
	if err := a.applyImportAttachments(ctx, command, job.ID, sourceIdentity, plan, sourceToAssetID, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (a App) applyImportFields(ctx context.Context, command ports.ImportJobCommand, plan importplan.Plan, result *ImportResult) error {
	existingFieldKeys, err := a.existingFieldKeys(ctx, command.TenantID, command.InventoryID)
	if err != nil {
		return err
	}
	total := len(plan.Fields)
	if total > 0 {
		if err := a.updateImportProgress(ctx, command, importjob.PhaseFields, 0, total, "Creating custom fields"); err != nil {
			return err
		}
	}
	for index, field := range plan.Fields {
		if err := a.stopIfImportCancelled(ctx, command); err != nil {
			return err
		}
		if _, ok := existingFieldKeys[field.Key]; ok {
			result.Counts.FieldsExisting++
			if err := a.updateImportProgress(ctx, command, importjob.PhaseFields, index+1, total, "Creating custom fields"); err != nil {
				return err
			}
			continue
		}
		_, err := a.CreateInventoryCustomFieldDefinition(ctx, CreateCustomFieldDefinitionInput{
			Principal:     command.Principal,
			Source:        audit.SourceImport,
			RequestID:     command.RequestID,
			TenantID:      command.TenantID,
			InventoryID:   command.InventoryID,
			Key:           field.Key,
			DisplayName:   field.DisplayName,
			Type:          field.Type,
			Applicability: customfield.ApplicabilityAllAssets.String(),
		})
		if err != nil {
			if errors.Is(err, ErrInvalidInput) {
				result.Counts.FieldsExisting++
				if err := a.updateImportProgress(ctx, command, importjob.PhaseFields, index+1, total, "Creating custom fields"); err != nil {
					return err
				}
				continue
			}
			return err
		}
		result.Counts.FieldsCreated++
		if err := a.updateImportProgress(ctx, command, importjob.PhaseFields, index+1, total, "Creating custom fields"); err != nil {
			return err
		}
	}
	return nil
}

func (a App) applyImportAssets(ctx context.Context, command ports.ImportJobCommand, jobID importjob.ID, sourceIdentity importSourceIdentity, plan importplan.Plan, duplicates map[string]struct{}, result *ImportResult) (map[string]string, error) {
	sourceToAssetID := map[string]string{}
	locations := sortedImportAssets(plan.Assets, "location")
	if len(locations) > 0 {
		if err := a.updateImportProgress(ctx, command, importjob.PhaseLocations, 0, len(locations), "Creating locations"); err != nil {
			return nil, err
		}
	}
	for index, planned := range locations {
		if err := a.stopIfImportCancelled(ctx, command); err != nil {
			return nil, err
		}
		created, skipped, err := a.createImportedAsset(ctx, command, jobID, sourceIdentity, planned, sourceToAssetID, duplicates)
		if err != nil {
			return nil, err
		}
		if skipped {
			result.Counts.AssetsSkipped++
			if err := a.updateImportProgress(ctx, command, importjob.PhaseLocations, index+1, len(locations), "Creating locations"); err != nil {
				return nil, err
			}
			continue
		}
		sourceToAssetID[planned.SourceID] = created.ID.String()
		result.Counts.LocationsCreated++
		if err := a.updateImportProgress(ctx, command, importjob.PhaseLocations, index+1, len(locations), "Creating locations"); err != nil {
			return nil, err
		}
	}
	items := sortedNonLocationImportAssets(plan.Assets)
	if len(items) > 0 {
		if err := a.updateImportProgress(ctx, command, importjob.PhaseAssets, 0, len(items), "Creating assets"); err != nil {
			return nil, err
		}
	}
	for index, planned := range items {
		if err := a.stopIfImportCancelled(ctx, command); err != nil {
			return nil, err
		}
		created, skipped, err := a.createImportedAsset(ctx, command, jobID, sourceIdentity, planned, sourceToAssetID, duplicates)
		if err != nil {
			return nil, err
		}
		if skipped {
			result.Counts.AssetsSkipped++
			if err := a.updateImportProgress(ctx, command, importjob.PhaseAssets, index+1, len(items), "Creating assets"); err != nil {
				return nil, err
			}
			continue
		}
		sourceToAssetID[planned.SourceID] = created.ID.String()
		result.Counts.AssetsCreated++
		if err := a.updateImportProgress(ctx, command, importjob.PhaseAssets, index+1, len(items), "Creating assets"); err != nil {
			return nil, err
		}
	}
	return sourceToAssetID, nil
}

func (a App) applyImportAttachments(ctx context.Context, command ports.ImportJobCommand, jobID importjob.ID, sourceIdentity importSourceIdentity, plan importplan.Plan, sourceToAssetID map[string]string, result *ImportResult) error {
	total := len(plan.Attachments)
	if total > 0 {
		if err := a.updateImportProgress(ctx, command, importjob.PhaseAttachments, 0, total, "Importing attachments"); err != nil {
			return err
		}
	}
	for index, attachment := range plan.Attachments {
		if err := a.stopIfImportCancelled(ctx, command); err != nil {
			return err
		}
		assetID, ok := sourceToAssetID[attachment.AssetSourceID]
		if !ok {
			result.Counts.AttachmentsSkipped++
			if err := a.updateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
				return err
			}
			continue
		}
		if link, found, err := a.importLinks.ImportSourceLinkByKey(ctx, importAttachmentSourceLinkKey(command.TenantID, command.InventoryID, sourceIdentity, attachment)); err != nil {
			return err
		} else if found {
			if link.ResourceType == ports.ImportResourceAttachment && strings.TrimSpace(link.ResourceID) != "" {
				a.recordImportSourceLinkDuplicateSkipped(ctx, command, ports.ImportSourceEntityAttachment, jobID)
				result.Counts.AttachmentsSkipped++
				if err := a.updateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
					return err
				}
				continue
			}
		}
		parsedAssetID, ok := asset.NewID(assetID)
		if !ok {
			result.Counts.AttachmentsSkipped++
			if err := a.updateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
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
			if err := a.updateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
				return err
			}
			continue
		}
		created, err := a.createImportedAttachment(ctx, command, jobID, sourceIdentity, parsedAssetID, attachment)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				result.Counts.AttachmentsSkipped++
				if err := a.updateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
					return err
				}
				continue
			}
			if errors.Is(err, ErrInvalidInput) || errors.Is(err, ErrAttachmentFileNameInvalid) || errors.Is(err, ErrAttachmentContentTypeUnsupported) || errors.Is(err, ErrAttachmentContentMismatch) || errors.Is(err, ErrAttachmentContentEmpty) || errors.Is(err, ErrAttachmentTooLarge) {
				result.Counts.AttachmentsSkipped++
				result.Messages = append(result.Messages, importplan.Message{
					Code:       "attachment-skipped",
					Severity:   importplan.SeverityWarning,
					Summary:    "Attachment could not be imported",
					Detail:     safeImportError(err),
					SourceID:   attachment.SourceID,
					SourceName: attachment.FileName,
				})
				if err := a.updateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
					return err
				}
				continue
			}
			return err
		}
		if created.ID.String() == "" {
			result.Counts.AttachmentsSkipped++
			if err := a.updateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
				return err
			}
			continue
		}
		result.Counts.AttachmentsCreated++
		if err := a.updateImportProgress(ctx, command, importjob.PhaseAttachments, index+1, total, "Importing attachments"); err != nil {
			return err
		}
	}
	return nil
}

func (a App) createImportedAttachment(ctx context.Context, command ports.ImportJobCommand, jobID importjob.ID, sourceIdentity importSourceIdentity, assetID asset.ID, planned importplan.Attachment) (media.Attachment, error) {
	if a.importAttachmentUnitOfWork == nil {
		return media.Attachment{}, ErrInvalidInput
	}
	prepared, err := a.prepareAttachment(ctx, CreateAttachmentInput{
		Principal:   command.Principal,
		Source:      audit.SourceImport,
		RequestID:   command.RequestID,
		TenantID:    command.TenantID,
		InventoryID: command.InventoryID,
		AssetID:     assetID,
		FileName:    planned.FileName,
		ContentType: planned.ContentType,
		Content:     planned.Content,
	})
	if err != nil {
		return media.Attachment{}, err
	}
	link, record, err := a.importedResourceRecords(importImportedResourceInput{
		TenantID:         command.TenantID,
		InventoryID:      command.InventoryID,
		JobID:            jobID,
		SourceIdentity:   sourceIdentity,
		SourceEntityType: ports.ImportSourceEntityAttachment,
		SourceEntityID:   planned.SourceID,
		ResourceType:     ports.ImportResourceAttachment,
		ResourceID:       prepared.Attachment.ID.String(),
		ResourceOwnerID:  assetID.String(),
		CreatedAt:        a.clock.Now().UTC(),
	})
	if err != nil {
		return media.Attachment{}, err
	}
	if err := a.blobs.PutBlob(ctx, prepared.StorageKey, prepared.ContentType, planned.Content); err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob storage failed"})
		return media.Attachment{}, err
	}
	if err := a.importAttachmentUnitOfWork.CreateImportedAttachment(ctx, prepared.Attachment, prepared.AuditRecord, link, record); err != nil {
		if deleteErr := a.blobs.DeleteBlob(ctx, prepared.StorageKey); deleteErr != nil {
			a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob cleanup failed"})
		}
		return media.Attachment{}, err
	}
	a.recordAttachmentCreated(ctx, CreateAttachmentInput{
		Principal:   command.Principal,
		TenantID:    command.TenantID,
		InventoryID: command.InventoryID,
		AssetID:     assetID,
	}, prepared.Attachment)
	return prepared.Attachment, nil
}

func (a App) updateImportProgress(ctx context.Context, command ports.ImportJobCommand, phase importjob.Phase, done int, total int, message string) error {
	for {
		job, err := a.importJob(ctx, command.TenantID, command.InventoryID, command.JobID)
		if err != nil {
			return err
		}
		updatedAt := a.clock.Now().UTC().Truncate(time.Microsecond)
		if !updatedAt.After(job.UpdatedAt) {
			updatedAt = job.UpdatedAt.UTC().Truncate(time.Microsecond).Add(time.Microsecond)
		}
		progress := importjob.Progress{Phase: phase, Done: done, Total: total, Message: message, UpdatedAt: updatedAt}
		if job.Status == importjob.StatusCancelRequested {
			mode := job.CancellationMode
			if mode == "" {
				mode = importjob.CancellationModeKeepPartial
			}
			if shouldPersistProgressAfterCancellation(job.Progress, progress) {
				updated, err := a.importJobs.UpdateImportJobProgress(ctx, command.TenantID, command.InventoryID, command.JobID, progress, job.UpdatedAt)
				if err != nil {
					return err
				}
				if !updated {
					continue
				}
				a.recordImportProgressUpdated(ctx, job, progress)
			}
			return importCancelledError{mode: mode}
		}
		if job.Status != importjob.StatusRunning {
			return ErrPrecondition
		}
		updated, err := a.importJobs.UpdateImportJobProgress(ctx, command.TenantID, command.InventoryID, command.JobID, progress, job.UpdatedAt)
		if err != nil {
			return err
		}
		if updated {
			a.recordImportProgressUpdated(ctx, job, progress)
			return nil
		}
	}
}

func shouldPersistProgressAfterCancellation(current importjob.Progress, next importjob.Progress) bool {
	if next.Done <= 0 {
		return false
	}
	if current.Phase != next.Phase {
		return true
	}
	return next.Done > current.Done
}

func (a App) createImportedAsset(ctx context.Context, command ports.ImportJobCommand, jobID importjob.ID, sourceIdentity importSourceIdentity, planned importplan.Asset, sourceToAssetID map[string]string, duplicates map[string]struct{}) (asset.Asset, bool, error) {
	if a.importAssetUnitOfWork == nil {
		return asset.Asset{}, false, ErrInvalidInput
	}
	if planned.Archived {
		return asset.Asset{}, true, nil
	}
	if link, found, err := a.importLinks.ImportSourceLinkByKey(ctx, importAssetSourceLinkKey(command.TenantID, command.InventoryID, sourceIdentity, planned)); err != nil {
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
	prepared, err := a.assetService.PrepareCreateAsset(ctx, CreateAssetInput{
		Principal:     command.Principal,
		Source:        audit.SourceImport,
		RequestID:     command.RequestID,
		TenantID:      command.TenantID,
		InventoryID:   command.InventoryID,
		Kind:          planned.Kind,
		Title:         planned.Title,
		Description:   planned.Description,
		ParentAssetID: parentAssetID,
		CustomFields:  planned.CustomFields,
	})
	if err != nil {
		return asset.Asset{}, false, err
	}
	link, record, err := a.importedResourceRecords(importImportedResourceInput{
		TenantID:         command.TenantID,
		InventoryID:      command.InventoryID,
		JobID:            jobID,
		SourceIdentity:   sourceIdentity,
		SourceEntityType: ports.ImportSourceEntityAsset,
		SourceEntityID:   planned.SourceID,
		ResourceType:     ports.ImportResourceAsset,
		ResourceID:       prepared.Asset.ID.String(),
		CreatedAt:        a.clock.Now().UTC(),
	})
	if err != nil {
		return asset.Asset{}, false, err
	}
	if err := a.importAssetUnitOfWork.CreateImportedAsset(ctx, prepared.Asset, prepared.AuditRecord, &prepared.UndoableOperation, prepared.PromotedParent, prepared.ParentPromotionRecord, link, record); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return asset.Asset{}, true, nil
		}
		return asset.Asset{}, false, err
	}
	a.assetService.RecordAssetCreated(ctx, prepared.Asset, command.Principal.ID)
	for _, key := range []string{"homebox-source-id", "homebox-asset-id"} {
		if value, ok := planned.CustomFields[key].(string); ok && strings.TrimSpace(value) != "" {
			duplicates[key+"="+strings.TrimSpace(value)] = struct{}{}
		}
	}
	return prepared.Asset, false, nil
}

func (a App) stopIfImportCancelled(ctx context.Context, command ports.ImportJobCommand) error {
	job, err := a.importJob(ctx, command.TenantID, command.InventoryID, command.JobID)
	if err != nil {
		return err
	}
	if job.Status != importjob.StatusCancelRequested {
		return nil
	}
	mode := job.CancellationMode
	if mode == "" {
		mode = importjob.CancellationModeKeepPartial
	}
	return importCancelledError{mode: mode}
}

func (a App) discardImportedJobResources(ctx context.Context, command ports.ImportJobCommand, jobID importjob.ID) (int, int, error) {
	if a.importLinks == nil {
		return 0, 0, ErrInvalidInput
	}
	resources, err := a.importLinks.ListAllImportJobResources(ctx, command.TenantID, command.InventoryID, jobID)
	if err != nil {
		return 0, 0, err
	}
	discarded := 0
	for index := len(resources) - 1; index >= 0; index-- {
		resource := resources[index]
		switch resource.ResourceType {
		case ports.ImportResourceAttachment:
			assetID, ok := asset.NewID(resource.ResourceOwnerID)
			if !ok {
				return discarded, 0, ErrInvalidInput
			}
			attachmentID, ok := media.NewID(resource.ResourceID)
			if !ok {
				return discarded, 0, ErrInvalidInput
			}
			if err := a.DeleteAttachment(ctx, UpdateAttachmentLifecycleInput{
				Principal:    command.Principal,
				Source:       audit.SourceImport,
				RequestID:    command.RequestID,
				TenantID:     command.TenantID,
				InventoryID:  command.InventoryID,
				AssetID:      assetID,
				AttachmentID: attachmentID,
			}); err != nil && !errors.Is(err, ErrNotFound) {
				return discarded, 0, err
			}
			discarded++
		case ports.ImportResourceAsset:
			assetID, ok := asset.NewID(resource.ResourceID)
			if !ok {
				return discarded, 0, ErrInvalidInput
			}
			if err := a.DeleteAsset(ctx, UpdateAssetLifecycleInput{
				Principal:   command.Principal,
				Source:      audit.SourceImport,
				RequestID:   command.RequestID,
				TenantID:    command.TenantID,
				InventoryID: command.InventoryID,
				AssetID:     assetID,
			}); err != nil && !errors.Is(err, ErrNotFound) {
				return discarded, 0, err
			}
			discarded++
		}
	}
	links, err := a.importLinks.DeleteImportSourceLinksForJob(ctx, command.TenantID, command.InventoryID, jobID)
	if err != nil {
		return discarded, 0, err
	}
	return discarded, links, nil
}

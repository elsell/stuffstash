package dataportability

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a ImportService) stopIfImportCancelled(ctx context.Context, command ports.ImportJobCommand) error {
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
	return ImportCancelledError{Mode: mode}
}

func (a ImportService) discardImportedJobResources(ctx context.Context, command ports.ImportJobCommand, jobID importjob.ID) (int, int, error) {
	if a.deps.ImportLinks == nil {
		return 0, 0, apperrors.ErrInvalidInput
	}
	resources, err := a.deps.ImportLinks.ListAllImportJobResources(ctx, command.TenantID, command.InventoryID, jobID)
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
				return discarded, 0, apperrors.ErrInvalidInput
			}
			attachmentID, ok := media.NewID(resource.ResourceID)
			if !ok {
				return discarded, 0, apperrors.ErrInvalidInput
			}
			if err := a.deps.Targets.DeleteAttachment(ctx, command, assetID, attachmentID); err != nil {
				if errors.Is(err, apperrors.ErrNotFound) {
					continue
				}
				return discarded, 0, err
			}
			discarded++
		case ports.ImportResourceAsset:
			assetID, ok := asset.NewID(resource.ResourceID)
			if !ok {
				return discarded, 0, apperrors.ErrInvalidInput
			}
			if err := a.deps.Targets.DeleteAsset(ctx, command, assetID); err != nil {
				if errors.Is(err, apperrors.ErrNotFound) {
					continue
				}
				return discarded, 0, err
			}
			discarded++
		}
	}
	links, err := a.deps.ImportLinks.DeleteImportSourceLinksForJob(ctx, command.TenantID, command.InventoryID, jobID)
	if err != nil {
		return discarded, 0, err
	}
	return discarded, links, nil
}

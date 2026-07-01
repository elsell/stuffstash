package app

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ActionPlanPhotoAttachmentMetadataInput struct {
	Decision ActionPlanDecisionInput
	Photos   []ActionPlanPhotoAttachmentMetadata
}

type ActionPlanPhotoAttachmentMetadata struct {
	CommandID   string
	FileName    string
	ContentType string
	SizeBytes   int64
}

func (a App) ValidateActionPlanPhotoAttachmentMetadata(ctx context.Context, input ActionPlanPhotoAttachmentMetadataInput) error {
	if len(input.Photos) == 0 {
		return nil
	}
	if err := a.ensureActionPlanDependencies(); err != nil {
		return err
	}
	if err := a.ensureActiveInventoryAccess(ctx, input.Decision.Principal, input.Decision.TenantID, input.Decision.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return err
	}
	record, found, err := a.actionPlans.ActionPlanByID(ctx, input.Decision.TenantID, input.Decision.InventoryID, strings.TrimSpace(input.Decision.PlanID))
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}
	if record.PrincipalID != input.Decision.Principal.ID || record.State != actionplan.StateProposed {
		return ErrConflict
	}
	attachableCommands := map[string]struct{}{}
	for _, command := range record.Commands {
		if actionPlanCommandCanReceivePhotos(command) {
			attachableCommands[command.ID] = struct{}{}
		}
	}
	for _, photo := range input.Photos {
		if _, ok := attachableCommands[strings.TrimSpace(photo.CommandID)]; !ok {
			return ErrInvalidInput
		}
		if _, ok := media.NewFileName(photo.FileName); !ok {
			return ErrInvalidInput
		}
		contentType, ok := media.NewContentType(photo.ContentType)
		if !ok || !actionPlanPhotoContentType(contentType) || photo.SizeBytes <= 0 || photo.SizeBytes > int64(a.maxAttachmentBytes) {
			return ErrInvalidInput
		}
	}
	return nil
}

func actionPlanCommandCanReceivePhotos(command ports.ActionPlanCommandRecord) bool {
	switch command.Kind {
	case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation, actionplan.CommandKindMoveAsset:
		return strings.TrimSpace(command.ID) != ""
	default:
		return false
	}
}

func actionPlanPhotoContentType(contentType media.ContentType) bool {
	switch contentType {
	case media.ContentTypeJPEG, media.ContentTypePNG, media.ContentTypeWEBP:
		return true
	default:
		return false
	}
}

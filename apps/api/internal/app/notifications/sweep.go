package notifications

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type GenerationCursor struct {
	AfterRecipientID  string
	ActiveRecipientID string
	AfterAssetID      string
}

func (s Service) GenerateSweepPage(ctx context.Context, cursor GenerationCursor, limit int) (GenerationCursor, bool, error) {
	if s.deps.Preferences == nil || limit < 1 || limit > maxGenerationPageSize {
		return cursor, false, apperrors.ErrInvalidInput
	}
	recipients, err := s.deps.Preferences.ListNotificationRecipients(ctx, cursor.AfterRecipientID, 1)
	if err != nil {
		return cursor, false, err
	}
	if len(recipients) == 0 {
		return GenerationCursor{}, true, nil
	}
	recipient := recipients[0]
	afterAsset := ""
	if recipient.ID == cursor.ActiveRecipientID {
		afterAsset = cursor.AfterAssetID
	}
	scope := recipient.Scope
	input := ScopeInput{Principal: identity.Principal{ID: scope.PrincipalID}, TenantID: scope.TenantID, InventoryID: scope.InventoryID, Source: audit.SourceBackgroundJob}
	page, err := s.GenerateRecipientPage(ctx, input, afterAsset, limit)
	if err != nil && !errors.Is(err, ports.ErrForbidden) && !errors.Is(err, apperrors.ErrNotFound) {
		return cursor, false, err
	}
	if err == nil && page.NextCursor != "" {
		return GenerationCursor{AfterRecipientID: cursor.AfterRecipientID, ActiveRecipientID: recipient.ID, AfterAssetID: page.NextCursor}, false, nil
	}
	return GenerationCursor{AfterRecipientID: recipient.ID}, false, nil
}

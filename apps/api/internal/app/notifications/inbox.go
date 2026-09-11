package notifications

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

type NotificationView struct {
	ParentTrail           []ports.NotificationAncestor
	ParentTrailIncomplete bool
	Notification          ports.NotificationRecord
	Asset                 asset.Asset
}

func (s Service) GetNotification(ctx context.Context, input ScopeInput, id string) (NotificationView, error) {
	view, err := s.currentNotification(ctx, input, id)
	if err != nil {
		return view, err
	}
	if err := s.enrichPlacement(ctx, input, &view, placementCache{}); err != nil {
		return NotificationView{}, err
	}
	auditInput := s.auditInput(input, audit.ActionNotificationListed, id)
	auditInput.TargetType = audit.TargetNotification
	if err := appsupport.SaveReadAuditRecord(ctx, s.deps.Audit, s.deps.IDs, s.deps.Clock, auditInput); err != nil {
		return NotificationView{}, err
	}
	return view, nil
}

func (s Service) MarkRead(ctx context.Context, input ScopeInput, id string) error {
	view, err := s.currentNotification(ctx, input, id)
	if err != nil {
		return err
	}
	if view.Notification.ReadAt != nil {
		return nil
	}
	auditInput := s.auditInput(input, audit.ActionNotificationRead, id)
	auditInput.TargetType = audit.TargetNotification
	record, err := appsupport.NewAuditRecord(s.deps.IDs, s.deps.Clock, auditInput)
	if err != nil {
		return err
	}
	_, err = s.deps.Inbox.MarkNotificationRead(ctx, input.Scope(), id, s.deps.Clock.Now(), record)
	return err
}

func (s Service) currentNotification(ctx context.Context, input ScopeInput, id string) (NotificationView, error) {
	if err := s.access(ctx, input); err != nil {
		return NotificationView{}, err
	}
	if id == "" || s.deps.Inbox == nil || s.deps.Assets == nil || s.deps.Types == nil || s.deps.Clock == nil {
		return NotificationView{}, apperrors.ErrInvalidInput
	}
	entry, found, err := s.deps.Inbox.NotificationByID(ctx, input.Scope(), id)
	if err != nil {
		return NotificationView{}, err
	}
	if !found {
		return NotificationView{}, apperrors.ErrNotFound
	}
	item, found, err := s.deps.Assets.AssetByID(ctx, input.TenantID, input.InventoryID, asset.ID(entry.Milestone.AssetID))
	if err != nil {
		return NotificationView{}, err
	}
	if !found || item.LifecycleState != asset.LifecycleStateActive || item.CustomAssetTypeID == "" {
		return NotificationView{}, apperrors.ErrNotFound
	}
	kind, found, err := s.deps.Types.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, customfield.AssetTypeID(item.CustomAssetTypeID))
	if err != nil {
		return NotificationView{}, err
	}
	if !found || !kind.IsActive() || !kind.ExpirationEnabled {
		return NotificationView{}, apperrors.ErrNotFound
	}
	preferences, found, err := s.deps.Preferences.NotificationPreferences(ctx, input.Scope())
	if err != nil {
		return NotificationView{}, err
	}
	settings := notification.DefaultSettings("UTC")
	if found {
		settings = preferences.Settings
	}
	zone, err := time.LoadLocation(settings.Timezone)
	if err != nil {
		return NotificationView{}, err
	}
	candidate := notification.ExpirationCandidate{AssetID: item.ID.String(), TypeID: notification.AssetTypeID(item.CustomAssetTypeID), Date: item.Expiration, Eligible: true}
	if !candidate.Visible(entry.Milestone, settings, s.deps.Clock.Now(), zone) {
		return NotificationView{}, apperrors.ErrNotFound
	}
	return NotificationView{Notification: entry, Asset: item}, nil
}

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

const maxGenerationPageSize = 100

type GenerationPage struct {
	Created    int
	NextCursor string
}

func (s Service) GenerateRecipientPage(ctx context.Context, input ScopeInput, afterID string, limit int) (GenerationPage, error) {
	if err := s.access(ctx, input); err != nil {
		return GenerationPage{}, err
	}
	if limit < 1 || limit > maxGenerationPageSize || s.deps.Assets == nil || s.deps.Deliveries == nil || s.deps.Types == nil || s.deps.Clock == nil || s.deps.IDs == nil {
		return GenerationPage{}, apperrors.ErrInvalidInput
	}
	preferences, found, err := s.deps.Preferences.NotificationPreferences(ctx, input.Scope())
	if err != nil {
		return GenerationPage{}, err
	}
	if !found {
		return GenerationPage{}, apperrors.ErrNotFound
	}
	zone, err := time.LoadLocation(preferences.Settings.Timezone)
	if err != nil {
		return GenerationPage{}, err
	}
	items, err := s.deps.Assets.ListAssetsByInventory(ctx, input.TenantID, input.InventoryID, ports.AssetListPageRequest{AfterAssetID: asset.ID(afterID), Limit: limit + 1, Sort: ports.AssetListSortIDAsc, LifecycleFilter: ports.AssetLifecycleFilterActive})
	if err != nil {
		return GenerationPage{}, err
	}
	result := GenerationPage{}
	now := s.deps.Clock.Now()
	for index, item := range items {
		if index == limit {
			break
		}
		if index+1 < len(items) {
			result.NextCursor = item.ID.String()
		} else {
			result.NextCursor = ""
		}
		if item.CustomAssetTypeID == "" || item.Expiration.Value() == "" || item.LifecycleState != asset.LifecycleStateActive {
			continue
		}
		kind, found, err := s.deps.Types.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, customfield.AssetTypeID(item.CustomAssetTypeID))
		if err != nil {
			return result, err
		}
		if !found || !kind.IsActive() || !kind.ExpirationEnabled {
			continue
		}
		candidate := notification.ExpirationCandidate{AssetID: item.ID.String(), TypeID: notification.AssetTypeID(item.CustomAssetTypeID), Date: item.Expiration, Eligible: true}
		milestone, due := candidate.Due(preferences.Settings, now, zone)
		if !due {
			continue
		}
		// Recheck current access before each publication; stale recipient registrations
		// must not authorize notifications after membership is removed.
		if err := s.access(ctx, input); err != nil {
			return result, err
		}
		value := ports.NotificationRecord{ID: s.deps.IDs.NewID(), Scope: input.Scope(), Milestone: milestone, CreatedAt: now}
		auditInput := s.auditInput(input, audit.ActionNotificationCreated, value.ID)
		auditInput.TargetType = audit.TargetNotification
		record, err := appsupport.NewAuditRecord(s.deps.IDs, s.deps.Clock, auditInput)
		if err != nil {
			return result, err
		}
		destinations, err := s.deliveriesForNotification(ctx, value, preferences.Settings.PushEnabled)
		if err != nil {
			return result, err
		}
		if err := s.access(ctx, input); err != nil {
			return result, err
		}
		_, created, err := s.deps.Deliveries.InsertNotificationWithDeliveries(ctx, value, destinations, record)
		if err != nil {
			return result, err
		}
		if created {
			result.Created++
			if s.deps.Observer != nil {
				s.deps.Observer.Record(ctx, ports.Event{Name: ports.EventNotificationCreated, Message: "expiration notification created", Fields: map[string]string{"tenant_id": input.TenantID.String(), "inventory_id": input.InventoryID.String(), "notification_id": value.ID}})
			}
		}
	}
	return result, nil
}

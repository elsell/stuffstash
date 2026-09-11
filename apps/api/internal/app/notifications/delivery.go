package notifications

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func (s Service) DeliverPage(ctx context.Context, limit int, lease time.Duration, policy notification.RetryPolicy) (int, error) {
	if s.deps.Authorizer == nil || s.deps.Inventories == nil || s.deps.Preferences == nil || s.deps.Inbox == nil || s.deps.Assets == nil || s.deps.Types == nil || s.deps.Deliveries == nil || s.deps.PushSender == nil || s.deps.Devices == nil || s.deps.IDs == nil || s.deps.Clock == nil {
		return 0, apperrors.ErrInvalidInput
	}
	if limit < 1 || limit > 100 {
		return 0, apperrors.ErrInvalidInput
	}
	completed := 0
	for completed < limit {
		if err := ctx.Err(); err != nil {
			return completed, err
		}
		jobs, err := s.deps.Deliveries.ClaimNotificationDeliveries(ctx, s.deps.Clock.Now(), s.deps.IDs.NewID(), lease, policy, 1)
		if err != nil {
			return completed, err
		}
		if len(jobs) == 0 {
			break
		}
		job := jobs[0]
		outcome, notBefore := s.deliver(ctx, job)
		if err := ctx.Err(); err != nil {
			return completed, err
		}
		if err := s.deps.Deliveries.SettleNotificationDelivery(ctx, job.ID, job.State.Fence, s.deps.Clock.Now(), outcome, policy, notBefore); err != nil {
			return completed, err
		}
		completed++
		if s.deps.Observer != nil {
			s.deps.Observer.Record(ctx, ports.Event{Name: ports.EventNotificationDeliverySettled, Message: "notification delivery processed", Fields: map[string]string{"delivery_id": job.ID, "outcome": string(outcome)}})
		}
	}
	return completed, nil
}
func (s Service) deliver(ctx context.Context, job ports.NotificationDelivery) (ports.NotificationDeliveryOutcome, time.Time) {
	input := ScopeInput{Principal: identity.Principal{ID: job.Scope.PrincipalID}, TenantID: job.Scope.TenantID, InventoryID: job.Scope.InventoryID}
	view, err := s.currentNotification(ctx, input, job.NotificationID)
	if err != nil {
		return deliveryFailureOutcome(err), time.Time{}
	}
	preferences, found, err := s.deps.Preferences.NotificationPreferences(ctx, job.Scope)
	if err != nil {
		return ports.NotificationDeliveryRetry, time.Time{}
	}
	if !found || !preferences.Settings.PushEnabled {
		return ports.NotificationDeliveryCancelled, time.Time{}
	}
	zone, err := time.LoadLocation(preferences.Settings.Timezone)
	if err != nil {
		return ports.NotificationDeliveryRetry, time.Time{}
	}
	candidate := notification.ExpirationCandidate{AssetID: view.Asset.ID.String(), TypeID: notification.AssetTypeID(view.Asset.CustomAssetTypeID), Date: view.Asset.Expiration, Eligible: true}
	due, eligible := candidate.Due(preferences.Settings, s.deps.Clock.Now(), zone)
	if !eligible || due != view.Notification.Milestone {
		return ports.NotificationDeliveryCancelled, time.Time{}
	}
	device, found, err := s.deps.Devices.NotificationDeviceByID(ctx, job.Scope, job.DeviceID)
	if err != nil {
		return ports.NotificationDeliveryRetry, time.Time{}
	}
	if !found || !device.Active || device.Revision != job.DeviceRevision {
		return ports.NotificationDeliveryCancelled, time.Time{}
	}
	if err := s.access(ctx, input); err != nil {
		return deliveryFailureOutcome(err), time.Time{}
	}
	remaining := job.State.LeaseUntil.Sub(s.deps.Clock.Now())
	if remaining <= 0 {
		return ports.NotificationDeliveryRetry, time.Time{}
	}
	sendCtx, cancel := context.WithTimeout(ctx, remaining)
	defer cancel()
	body := "An item is expiring soon."
	if due.Kind == notification.MilestoneExpired {
		body = "An item has expired."
	}
	result, err := s.deps.PushSender.SendNotification(sendCtx, ports.NotificationPushMessage{DeliveryID: job.ID, NotificationID: job.NotificationID, Scope: job.Scope, Transport: device.Transport, Token: device.Token, Title: "Stuff Stash", Body: body})
	if err != nil {
		return ports.NotificationDeliveryRetry, time.Time{}
	}
	switch result.Outcome {
	case ports.NotificationPushAccepted:
		return ports.NotificationDeliveryAccepted, time.Time{}
	case ports.NotificationPushInvalidDevice:
		if !result.InvalidatedAt.IsZero() && device.UpdatedAt.After(result.InvalidatedAt) {
			return ports.NotificationDeliveryRetry, time.Time{}
		}
		_, err := s.RevokeDevice(ctx, input, job.DeviceID, job.DeviceRevision)
		if err != nil && !errors.Is(err, ports.ErrConflict) && !errors.Is(err, apperrors.ErrNotFound) {
			return ports.NotificationDeliveryRetry, time.Time{}
		}
		return ports.NotificationDeliveryCancelled, time.Time{}
	default:
		return ports.NotificationDeliveryRetry, result.RetryNotBefore
	}
}
func deliveryFailureOutcome(err error) ports.NotificationDeliveryOutcome {
	if errors.Is(err, apperrors.ErrNotFound) || errors.Is(err, apperrors.ErrUnauthorized) || errors.Is(err, apperrors.ErrUnauthenticated) {
		return ports.NotificationDeliveryCancelled
	}
	return ports.NotificationDeliveryRetry
}

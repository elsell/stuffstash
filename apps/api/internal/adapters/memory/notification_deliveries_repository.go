package memory

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
	"time"
)

var _ ports.NotificationDeliveryRepository = (*Store)(nil)

func (s *Store) InsertNotificationWithDeliveries(ctx context.Context, value ports.NotificationRecord, deliveries []ports.NotificationDelivery, record audit.Record) (ports.NotificationRecord, bool, error) {
	if err := ctx.Err(); err != nil {
		return ports.NotificationRecord{}, false, err
	}
	if !value.Valid() || value.ReadAt != nil {
		return ports.NotificationRecord{}, false, ports.ErrInvalidProviderInput
	}
	if !notificationAuditScopeMatches(value.Scope, record) {
		return ports.NotificationRecord{}, false, ports.ErrForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.notificationInbox {
		if existing.Scope == value.Scope && existing.Milestone == value.Milestone {
			return existing.Clone(), false, nil
		}
	}
	seenIDs := map[string]bool{}
	seenDevices := map[string]bool{}
	for _, delivery := range deliveries {
		state, err := notification.NewDelivery(value.CreatedAt)
		if err != nil || delivery.ID == "" || delivery.Scope != value.Scope || delivery.NotificationID != value.ID || !delivery.CreatedAt.Equal(value.CreatedAt) || delivery.State != state || seenIDs[delivery.ID] || seenDevices[delivery.DeviceID] {
			return ports.NotificationRecord{}, false, ports.ErrInvalidProviderInput
		}
		device, found := s.notificationDevices[delivery.DeviceID]
		if !found || device.Scope != value.Scope || !device.Active || device.Revision != delivery.DeviceRevision {
			return ports.NotificationRecord{}, false, ports.ErrConflict
		}
		if _, exists := s.notificationDeliveries[delivery.ID]; exists {
			return ports.NotificationRecord{}, false, ports.ErrConflict
		}
		seenIDs[delivery.ID] = true
		seenDevices[delivery.DeviceID] = true
	}
	result, created, err := s.insertNotificationLocked(value, record)
	if err != nil || !created {
		return result, created, err
	}
	if s.notificationDeliveries == nil {
		s.notificationDeliveries = map[string]ports.NotificationDelivery{}
	}
	for _, delivery := range deliveries {
		s.notificationDeliveries[delivery.ID] = delivery
	}
	return result, true, nil
}
func (s *Store) ClaimNotificationDeliveries(ctx context.Context, now time.Time, fence string, lease time.Duration, policy notification.RetryPolicy, limit int) ([]ports.NotificationDelivery, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if now.IsZero() || fence == "" || len(fence) > 128 || lease <= 0 || policy.Validate() != nil || limit < 1 || limit > 100 {
		return nil, ports.ErrInvalidProviderInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.notificationDeliveries))
	for id := range s.notificationDeliveries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := []ports.NotificationDelivery{}
	updates := map[string]ports.NotificationDelivery{}
	for _, id := range ids {
		if len(updates) == limit {
			break
		}
		delivery := s.notificationDeliveries[id]
		state, err := delivery.State.Claim(now, fence, lease, policy)
		if errors.Is(err, notification.ErrDeliveryNotDue) {
			continue
		}
		if err != nil {
			return nil, err
		}
		delivery.State = state
		updates[id] = delivery
		if state.Status == notification.DeliveryLeased {
			result = append(result, delivery)
		}
	}
	for id, delivery := range updates {
		s.notificationDeliveries[id] = delivery
	}
	return result, nil
}
func (s *Store) SettleNotificationDelivery(ctx context.Context, id, fence string, now time.Time, outcome ports.NotificationDeliveryOutcome, policy notification.RetryPolicy, notBefore time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delivery, found := s.notificationDeliveries[id]
	if !found {
		return notification.ErrStaleDeliveryLease
	}
	var state notification.DeliveryState
	var err error
	switch outcome {
	case ports.NotificationDeliveryAccepted:
		state, err = delivery.State.Accept(now, fence)
	case ports.NotificationDeliveryCancelled:
		state, err = delivery.State.Cancel(now, fence)
	case ports.NotificationDeliveryRetry:
		state, err = delivery.State.Retry(now, fence, policy, notBefore)
	default:
		return ports.ErrInvalidProviderInput
	}
	if err != nil {
		return err
	}
	delivery.State = state
	s.notificationDeliveries[id] = delivery
	return nil
}

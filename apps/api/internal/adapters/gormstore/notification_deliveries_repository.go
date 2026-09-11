package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sort"
	"time"
)

var _ ports.NotificationDeliveryRepository = Store{}

func (s Store) InsertNotificationWithDeliveries(ctx context.Context, value ports.NotificationRecord, deliveries []ports.NotificationDelivery, record audit.Record) (ports.NotificationRecord, bool, error) {
	var result ports.NotificationRecord
	var created bool
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		result, created, err = (Store{db: tx}).InsertNotification(ctx, value, record)
		if err != nil || !created {
			return err
		}
		destinations := append([]ports.NotificationDelivery(nil), deliveries...)
		sort.Slice(destinations, func(i, j int) bool { return destinations[i].DeviceID < destinations[j].DeviceID })
		seen := map[string]bool{}
		ids := map[string]bool{}
		for _, delivery := range destinations {
			state, _ := notification.NewDelivery(value.CreatedAt)
			if delivery.ID == "" || delivery.DeviceID == "" || delivery.Scope != value.Scope || delivery.NotificationID != value.ID || !delivery.CreatedAt.Equal(value.CreatedAt) || delivery.State != state || seen[delivery.DeviceID] || ids[delivery.ID] {
				return ports.ErrInvalidProviderInput
			}
			seen[delivery.DeviceID] = true
			ids[delivery.ID] = true
			var device notificationDeviceModel
			err := tx.Select("id", "active", "revision").Clauses(clause.Locking{Strength: "UPDATE"}).Where(notificationDeviceScope(value.Scope)).Where(&notificationDeviceModel{ID: delivery.DeviceID}).First(&device).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrConflict
			}
			if err != nil {
				return err
			}
			if !device.Active || device.Revision != delivery.DeviceRevision {
				return ports.ErrConflict
			}
			model := deliveryModel(delivery)
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ports.NotificationRecord{}, false, err
	}
	return result, created, nil
}
func (s Store) ClaimNotificationDeliveries(ctx context.Context, now time.Time, fence string, lease time.Duration, policy notification.RetryPolicy, limit int) ([]ports.NotificationDelivery, error) {
	if now.IsZero() || fence == "" || len(fence) > 128 || lease <= 0 || policy.Validate() != nil || limit < 1 || limit > 100 {
		return nil, ports.ErrInvalidProviderInput
	}
	result := []ports.NotificationDelivery{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []notificationDeliveryModel
		due := clause.Or(clause.And(clause.Eq{Column: "status", Value: string(notification.DeliveryPending)}, clause.Lte{Column: "next_attempt_at", Value: now}), clause.And(clause.Eq{Column: "status", Value: string(notification.DeliveryLeased)}, clause.Lte{Column: "lease_until", Value: now}))
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where(due).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(limit).Find(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			delivery := row.toDelivery()
			state, err := delivery.State.Claim(now, fence, lease, policy)
			if err != nil {
				return err
			}
			if err := saveDeliveryState(tx, row.ID, state); err != nil {
				return err
			}
			delivery.State = state
			if state.Status == notification.DeliveryLeased {
				result = append(result, delivery)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (s Store) SettleNotificationDelivery(ctx context.Context, id, fence string, now time.Time, outcome ports.NotificationDeliveryOutcome, policy notification.RetryPolicy) error {
	if id == "" {
		return notification.ErrStaleDeliveryLease
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row notificationDeliveryModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&notificationDeliveryModel{ID: id}).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return notification.ErrStaleDeliveryLease
		}
		if err != nil {
			return err
		}
		current := row.toDelivery().State
		var next notification.DeliveryState
		switch outcome {
		case ports.NotificationDeliveryAccepted:
			next, err = current.Accept(now, fence)
		case ports.NotificationDeliveryCancelled:
			next, err = current.Cancel(now, fence)
		case ports.NotificationDeliveryRetry:
			next, err = current.Retry(now, fence, policy)
		default:
			return ports.ErrInvalidProviderInput
		}
		if err != nil {
			return err
		}
		return saveDeliveryState(tx, id, next)
	})
}
func saveDeliveryState(tx *gorm.DB, id string, state notification.DeliveryState) error {
	var model notificationDeliveryModel
	model.setState(state)
	return tx.Model(&notificationDeliveryModel{}).Where(&notificationDeliveryModel{ID: id}).Updates(map[string]any{"status": model.Status, "attempts": model.Attempts, "next_attempt_at": model.NextAttemptAt, "lease_until": model.LeaseUntil, "fence": model.Fence}).Error
}

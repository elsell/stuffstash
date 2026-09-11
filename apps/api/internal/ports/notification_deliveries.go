package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"time"
)

type NotificationDelivery struct {
	ID             string
	Scope          NotificationScope
	NotificationID string
	DeviceID       string
	DeviceRevision int64
	CreatedAt      time.Time
	State          notification.DeliveryState
}
type NotificationDeliveryOutcome string

const (
	NotificationDeliveryAccepted  NotificationDeliveryOutcome = "accepted"
	NotificationDeliveryCancelled NotificationDeliveryOutcome = "cancelled"
	NotificationDeliveryRetry     NotificationDeliveryOutcome = "retry"
)

type NotificationDeliveryRepository interface {
	InsertNotificationWithDeliveries(context.Context, NotificationRecord, []NotificationDelivery, audit.Record) (NotificationRecord, bool, error)
	ClaimNotificationDeliveries(context.Context, time.Time, string, time.Duration, notification.RetryPolicy, int) ([]NotificationDelivery, error)
	SettleNotificationDelivery(context.Context, string, string, time.Time, NotificationDeliveryOutcome, notification.RetryPolicy) error
}

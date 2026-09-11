package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"time"
)

type NotificationPushOutcome string

const (
	NotificationPushAccepted      NotificationPushOutcome = "accepted"
	NotificationPushRetry         NotificationPushOutcome = "retry"
	NotificationPushInvalidDevice NotificationPushOutcome = "invalid_device"
)

type NotificationPushMessage struct {
	DeliveryID     string
	NotificationID string
	Scope          NotificationScope
	Transport      notification.PushTransport
	Token          notification.DeviceToken
	Title          string
	Body           string
}
type NotificationPushResult struct {
	Outcome        NotificationPushOutcome
	InvalidatedAt  time.Time
	RetryNotBefore time.Time
}
type NotificationPushSender interface {
	SendNotification(context.Context, NotificationPushMessage) (NotificationPushResult, error)
}

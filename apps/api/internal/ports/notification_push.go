package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
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
type NotificationPushSender interface {
	SendNotification(context.Context, NotificationPushMessage) (NotificationPushOutcome, error)
}

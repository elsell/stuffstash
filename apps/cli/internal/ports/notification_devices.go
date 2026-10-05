package ports

import "context"

type NotificationDevice struct {
	ID             string `json:"id"`
	InstallationID string `json:"installationId"`
	Transport      string `json:"transport"`
	Revision       int64  `json:"revision"`
	Active         bool   `json:"active"`
}
type NotificationDevicesAPI interface {
	NotificationDevice(context.Context, Scope, string) (Result[NotificationDevice], error)
	RegisterNotificationDevice(context.Context, Scope, []byte) (Result[NotificationDevice], error)
	RemoveNotificationDevice(context.Context, Scope, string, int64) (Result[NotificationDevice], error)
}

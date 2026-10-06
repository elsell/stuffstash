package ports

import (
	"context"
	"time"
)

type NotificationAncestor struct {
	AssetID string `json:"assetId"`
	Title   string `json:"title"`
	Kind    string `json:"kind"`
}
type Notification struct {
	ID                    string                 `json:"id"`
	AssetID               string                 `json:"assetId"`
	Title                 string                 `json:"title"`
	ParentAssetID         string                 `json:"parentAssetId"`
	CustomAssetTypeID     string                 `json:"customAssetTypeId"`
	ExpirationDate        string                 `json:"expirationDate"`
	ExpirationPrecision   string                 `json:"expirationPrecision"`
	Milestone             string                 `json:"milestone"`
	CreatedAt             time.Time              `json:"createdAt"`
	ReadAt                *time.Time             `json:"readAt,omitempty"`
	ParentTrail           []NotificationAncestor `json:"parentTrail"`
	ParentTrailIncomplete bool                   `json:"parentTrailIncomplete"`
}
type NotificationRead struct {
	ID   string `json:"id"`
	Read bool   `json:"read"`
}
type NotificationReadAll struct {
	Complete bool `json:"complete"`
}
type NotificationCount struct {
	Count int64 `json:"count"`
}
type NotificationsAPI interface {
	Notifications(context.Context, Scope, Page, bool) (Result[[]Notification], error)
	Notification(context.Context, Scope, string) (Result[Notification], error)
	NotificationUnreadCount(context.Context, Scope, string) (Result[NotificationCount], error)
	SetNotificationRead(context.Context, Scope, string, bool) (Result[NotificationRead], error)
	ReadAllNotifications(context.Context, Scope, string) (Result[NotificationReadAll], error)
}

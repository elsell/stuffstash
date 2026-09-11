package dto

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"time"
)

type InboxListInput struct {
	ScopeInput
	Cursor     string `query:"cursor" maxLength:"128"`
	Limit      int    `query:"limit" default:"30" minimum:"1" maximum:"100"`
	UnreadOnly bool   `query:"unreadOnly"`
}
type NotificationInput struct {
	ScopeInput
	NotificationID string `path:"notificationId"`
}
type NotificationAncestor struct {
	AssetID string `json:"assetId"`
	Title   string `json:"title"`
	Kind    string `json:"kind" enum:"item,container,location"`
}
type NotificationResponse struct {
	ParentTrail           []NotificationAncestor `json:"parentTrail"`
	ParentTrailIncomplete bool                   `json:"parentTrailIncomplete"`
	ID                    string                 `json:"id"`
	AssetID               string                 `json:"assetId"`
	Title                 string                 `json:"title"`
	ParentAssetID         string                 `json:"parentAssetId"`
	CustomAssetTypeID     string                 `json:"customAssetTypeId"`
	ExpirationDate        string                 `json:"expirationDate"`
	ExpirationPrecision   string                 `json:"expirationPrecision" enum:"day,month"`
	Milestone             string                 `json:"milestone" enum:"upcoming,expired"`
	CreatedAt             time.Time              `json:"createdAt"`
	ReadAt                *time.Time             `json:"readAt,omitempty"`
}
type InboxOutput struct {
	Body shared.SuccessEnvelope[[]NotificationResponse]
}
type NotificationOutput struct {
	Body shared.SuccessEnvelope[NotificationResponse]
}
type NotificationReadResponse struct {
	ID   string `json:"id"`
	Read bool   `json:"read"`
}
type NotificationReadOutput struct {
	Body shared.SuccessEnvelope[NotificationReadResponse]
}

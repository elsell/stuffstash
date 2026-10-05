package ports

import "context"

type ExpirationPolicy struct {
	Enabled     bool  `json:"enabled"`
	Upcoming    bool  `json:"upcoming"`
	Expired     bool  `json:"expired"`
	AdvanceDays int64 `json:"advanceDays"`
}
type NotificationTypeOverride struct {
	CustomAssetTypeID string           `json:"customAssetTypeId"`
	Settings          ExpirationPolicy `json:"settings"`
}
type NotificationPreferences struct {
	Revision    int64                      `json:"revision"`
	Defaults    ExpirationPolicy           `json:"defaults"`
	Timezone    string                     `json:"timezone"`
	PushEnabled bool                       `json:"pushEnabled"`
	Overrides   []NotificationTypeOverride `json:"overrides"`
}
type PreferenceAction string

const (
	UpdatePreferences        PreferenceAction = "update"
	InitializePreferences    PreferenceAction = "initialize"
	OverridePreferences      PreferenceAction = "override"
	RemovePreferenceOverride PreferenceAction = "remove-override"
)

type NotificationPreferencesAPI interface {
	NotificationPreferences(context.Context, Scope) (Result[NotificationPreferences], error)
	ChangeNotificationPreferences(context.Context, Scope, PreferenceAction, string, int64, []byte) (Result[NotificationPreferences], error)
}

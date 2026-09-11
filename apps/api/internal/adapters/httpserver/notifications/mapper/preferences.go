package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"sort"
)

func PolicyToDomain(value dto.ExpirationPolicy) notification.ExpirationPreferences {
	return notification.ExpirationPreferences{Enabled: value.Enabled, Upcoming: value.Upcoming, Expired: value.Expired, AdvanceDays: value.AdvanceDays}
}
func PolicyToResponse(value notification.ExpirationPreferences) dto.ExpirationPolicy {
	return dto.ExpirationPolicy{Enabled: value.Enabled, Upcoming: value.Upcoming, Expired: value.Expired, AdvanceDays: value.AdvanceDays}
}
func SettingsToResponse(value notification.Settings, revision int64) dto.PreferencesResponse {
	overrides := make([]dto.TypeOverrideResponse, 0, len(value.Overrides))
	for id, policy := range value.Overrides {
		overrides = append(overrides, dto.TypeOverrideResponse{CustomAssetTypeID: string(id), Settings: PolicyToResponse(policy)})
	}
	sort.Slice(overrides, func(i, j int) bool { return overrides[i].CustomAssetTypeID < overrides[j].CustomAssetTypeID })
	return dto.PreferencesResponse{Revision: revision, Defaults: PolicyToResponse(value.Defaults), Timezone: value.Timezone, PushEnabled: value.PushEnabled, Overrides: overrides}
}

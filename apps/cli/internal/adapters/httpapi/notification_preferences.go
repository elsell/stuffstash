package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func expirationPolicy(v generated.ExpirationPolicy) ports.ExpirationPolicy {
	return ports.ExpirationPolicy{Enabled: v.Enabled, Upcoming: v.Upcoming, Expired: v.Expired, AdvanceDays: v.AdvanceDays}
}
func preferenceResult(response *http.Response, err error) (ports.Result[ports.NotificationPreferences], error) {
	r, err := read[generated.SuccessEnvelopePreferencesResponse](response, err)
	if err != nil {
		return ports.Result[ports.NotificationPreferences]{}, err
	}
	v := r.Data
	var overrides []ports.NotificationTypeOverride
	if v.Overrides.GetOrEmpty() != nil {
		overrides = make([]ports.NotificationTypeOverride, 0, len(v.Overrides.GetOrEmpty()))
	}
	for _, o := range v.Overrides.GetOrEmpty() {
		overrides = append(overrides, ports.NotificationTypeOverride{CustomAssetTypeID: o.CustomAssetTypeId, Settings: expirationPolicy(o.Settings)})
	}
	return ports.Result[ports.NotificationPreferences]{Schema: r.Schema, Meta: metadata(r.Meta), Data: ports.NotificationPreferences{Revision: v.Revision, Defaults: expirationPolicy(v.Defaults), Timezone: v.Timezone, PushEnabled: v.PushEnabled, Overrides: overrides}}, nil
}
func (c *Client) NotificationPreferences(ctx context.Context, s ports.Scope) (ports.Result[ports.NotificationPreferences], error) {
	return preferenceResult(c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdNotificationPreferences(ctx, s.Tenant, s.Inventory, nil))
}
func (c *Client) ChangeNotificationPreferences(ctx context.Context, s ports.Scope, action ports.PreferenceAction, id string, revision int64, body []byte) (ports.Result[ports.NotificationPreferences], error) {
	switch action {
	case ports.UpdatePreferences:
		return preferenceResult(c.sdk.PutTenantsByTenantIdInventoriesByInventoryIdNotificationPreferencesWithBody(ctx, s.Tenant, s.Inventory, nil, "application/json", bytes.NewReader(body)))
	case ports.InitializePreferences:
		return preferenceResult(c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdNotificationPreferencesInitializeWithBody(ctx, s.Tenant, s.Inventory, nil, "application/json", bytes.NewReader(body)))
	case ports.OverridePreferences:
		return preferenceResult(c.sdk.PutTenantsByTenantIdInventoriesByInventoryIdNotificationPreferencesTypesByCustomAssetTypeIdWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body)))
	case ports.RemovePreferenceOverride:
		return preferenceResult(c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryIdNotificationPreferencesTypesByCustomAssetTypeId(ctx, s.Tenant, s.Inventory, id, &generated.DeleteTenantsByTenantIdInventoriesByInventoryIdNotificationPreferencesTypesByCustomAssetTypeIdParams{Revision: revision}))
	default:
		return ports.Result[ports.NotificationPreferences]{}, ports.Failure("usage", "The preference action is not available. Use --help to select a command.")
	}
}

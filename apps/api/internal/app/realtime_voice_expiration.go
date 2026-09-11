package app

import (
	"context"
	expirationapp "github.com/stuffstash/stuff-stash/internal/app/expiration"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type realtimeVoiceExpiration struct {
	Date            string                   `json:"date"`
	Precision       expirationdate.Precision `json:"precision"`
	State           expirationdate.State     `json:"state"`
	TrackingEnabled bool                     `json:"trackingEnabled"`
	AdvanceDays     int                      `json:"advanceDays"`
	Timezone        string                   `json:"timezone"`
}

func (a App) realtimeVoiceExpiration(ctx context.Context, session RealtimeVoiceSession, item asset.Asset) (*realtimeVoiceExpiration, error) {
	if item.Expiration.Value() == "" {
		return nil, nil
	}
	preferences, err := a.notificationService.GetPreferences(ctx, notificationapp.ScopeInput{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID, Source: audit.SourceConversation})
	if err != nil {
		return nil, err
	}
	if a.customAssetTypes == nil || a.clock == nil {
		return nil, ports.ErrInvalidProviderInput
	}
	kind, found, err := a.customAssetTypes.CustomAssetTypeByID(ctx, session.TenantID, session.InventoryID, customfield.AssetTypeID(item.CustomAssetTypeID))
	if err != nil {
		return nil, err
	}
	enabled := found && kind.ExpirationEnabled && kind.LifecycleState == customfield.AssetTypeLifecycleActive && item.LifecycleState == asset.LifecycleStateActive
	value, err := expirationapp.Describe(item.Expiration, notification.AssetTypeID(item.CustomAssetTypeID), enabled, preferences.Settings, a.clock.Now())
	if err != nil {
		return nil, err
	}
	return &realtimeVoiceExpiration{Date: value.Date.Value(), Precision: value.Date.Precision(), State: value.State, TrackingEnabled: value.TrackingEnabled, AdvanceDays: value.AdvanceDays, Timezone: value.Timezone}, nil
}

package agentmodel

import (
	"context"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
)

type RealtimeVoiceExpiration struct {
	Date            string                   `json:"date"`
	Precision       expirationdate.Precision `json:"precision"`
	State           expirationdate.State     `json:"state"`
	TrackingEnabled bool                     `json:"trackingEnabled"`
	AdvanceDays     int                      `json:"advanceDays"`
	Timezone        string                   `json:"timezone"`
}

func (a RealtimeReadTools) RealtimeVoiceExpiration(ctx context.Context, session RealtimeReadToolScope, item asset.Asset) (*RealtimeVoiceExpiration, error) {
	value, err := a.DescribeAssetExpiration(ctx, notificationapp.ScopeInput{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID, Source: audit.SourceConversation}, item)
	if err != nil || value == nil {
		return nil, err
	}
	return &RealtimeVoiceExpiration{Date: value.Date.Value(), Precision: value.Date.Precision(), State: value.State, TrackingEnabled: value.TrackingEnabled, AdvanceDays: value.AdvanceDays, Timezone: value.Timezone}, nil
}

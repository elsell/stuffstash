package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
)

func ExpirationToResponse(value expirationdate.Date) *dto.Expiration {
	if value.Value() == "" {
		return nil
	}
	return &dto.Expiration{Date: value.Value(), Precision: string(value.Precision())}
}

func ExpirationContextToResponse(state expirationdate.State, enabled bool, days int, timezone string) *dto.ExpirationContext {
	return &dto.ExpirationContext{State: string(state), TrackingEnabled: enabled, AdvanceDays: days, Timezone: timezone}
}

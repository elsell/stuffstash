package app

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type App struct {
	observer         ports.Observer
	auth             ports.Authenticator
	authorizer       ports.Authorizer
	tenants          ports.TenantRepository
	inventories      ports.InventoryRepository
	assets           ports.AssetRepository
	outbox           ports.AuthorizationOutbox
	ids              ports.IDGenerator
	outboxDrainLimit int
	outboxClaimLease time.Duration
	defaultPageLimit int
	maxPageLimit     int
}

type Dependencies struct {
	Observer                      ports.Observer
	Auth                          ports.Authenticator
	Authorizer                    ports.Authorizer
	Tenants                       ports.TenantRepository
	Inventories                   ports.InventoryRepository
	Assets                        ports.AssetRepository
	Outbox                        ports.AuthorizationOutbox
	IDs                           ports.IDGenerator
	AuthorizationOutboxDrainLimit int
	AuthorizationOutboxClaimLease time.Duration
	DefaultPageLimit              int
	MaxPageLimit                  int
}

func New(deps Dependencies) App {
	maxPageLimit := normalizeMaxPageLimit(deps.MaxPageLimit)
	return App{
		observer:         deps.Observer,
		auth:             deps.Auth,
		authorizer:       deps.Authorizer,
		tenants:          deps.Tenants,
		inventories:      deps.Inventories,
		assets:           deps.Assets,
		outbox:           deps.Outbox,
		ids:              deps.IDs,
		outboxDrainLimit: deps.AuthorizationOutboxDrainLimit,
		outboxClaimLease: deps.AuthorizationOutboxClaimLease,
		defaultPageLimit: normalizeDefaultPageLimit(deps.DefaultPageLimit, maxPageLimit),
		maxPageLimit:     maxPageLimit,
	}
}

func normalizeDefaultPageLimit(defaultLimit int, maxLimit int) int {
	if defaultLimit <= 0 {
		return 50
	}
	if defaultLimit > maxLimit {
		return maxLimit
	}
	return defaultLimit
}

func normalizeMaxPageLimit(maxLimit int) int {
	if maxLimit <= 0 {
		return 100
	}
	return maxLimit
}

func (a App) Authenticate(ctx context.Context, authorizationHeader string) (identity.Principal, error) {
	principal, err := a.auth.Authenticate(ctx, authorizationHeader)
	if err != nil {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventAuthenticationFailed,
			Message: "authentication failed",
		})
		return identity.Principal{}, err
	}

	return principal, nil
}

func (a App) Health(ctx context.Context) HealthStatus {
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventHealthChecked,
		Message: "health check completed",
	})

	return HealthStatus{
		Service: ServiceNameStuffStash,
		Status:  HealthStatusHealthy,
	}
}

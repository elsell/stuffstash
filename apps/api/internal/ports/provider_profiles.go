package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type ProviderProfileRepository interface {
	ProviderProfileByID(ctx context.Context, tenantID tenant.ID, profileID agentmodel.ProviderProfileID) (agentmodel.ProviderProfile, bool, error)
	ListProviderProfiles(ctx context.Context, tenantID tenant.ID) ([]agentmodel.ProviderProfile, error)
}

type ProviderProfileUnitOfWork interface {
	SaveProviderProfile(ctx context.Context, profile agentmodel.ProviderProfile, auditRecord audit.Record) error
	UpdateProviderProfile(ctx context.Context, profile agentmodel.ProviderProfile, auditRecord audit.Record) error
}

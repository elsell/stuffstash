package agentmodel

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Service struct {
	observer                  ports.Observer
	authorizer                ports.Authorizer
	providerProfiles          ports.ProviderProfileRepository
	providerProfileUnitOfWork ports.ProviderProfileUnitOfWork
	voiceProviderConfigs      ports.VoiceProviderConfigurationRepository
	providerCredentialVault   ports.ProviderCredentialVault
	providerProfileTester     ports.ProviderProfileTester
	ids                       ports.IDGenerator
	clock                     ports.Clock
}

type Dependencies struct {
	Observer                  ports.Observer
	Authorizer                ports.Authorizer
	ProviderProfiles          ports.ProviderProfileRepository
	ProviderProfileUnitOfWork ports.ProviderProfileUnitOfWork
	VoiceProviderConfigs      ports.VoiceProviderConfigurationRepository
	ProviderCredentialVault   ports.ProviderCredentialVault
	ProviderProfileTester     ports.ProviderProfileTester
	IDs                       ports.IDGenerator
	Clock                     ports.Clock
}

func New(deps Dependencies) Service {
	observer := deps.Observer
	if observer == nil {
		observer = noopObserver{}
	}
	unitOfWork := deps.ProviderProfileUnitOfWork
	if unitOfWork == nil {
		if cast, ok := deps.ProviderProfiles.(ports.ProviderProfileUnitOfWork); ok {
			unitOfWork = cast
		}
	}
	clock := deps.Clock
	if clock == nil {
		clock = ports.SystemClock{}
	}
	return Service{
		observer:                  observer,
		authorizer:                deps.Authorizer,
		providerProfiles:          deps.ProviderProfiles,
		providerProfileUnitOfWork: unitOfWork,
		voiceProviderConfigs:      deps.VoiceProviderConfigs,
		providerCredentialVault:   deps.ProviderCredentialVault,
		providerProfileTester:     deps.ProviderProfileTester,
		ids:                       deps.IDs,
		clock:                     clock,
	}
}

type noopObserver struct{}

func (noopObserver) Record(context.Context, ports.Event) {}

func (s Service) ensureTenantConfigure(ctx context.Context, principal identity.Principal, tenantID tenant.ID) error {
	if err := s.authorizer.CheckTenant(ctx, principal, ports.TenantPermissionConfigure, tenantID); err != nil {
		s.observer.Record(ctx, ports.Event{
			Name:    ports.EventAuthorizationDenied,
			Message: "authorization denied",
			Fields: map[string]string{
				"tenant_id":    tenantID.String(),
				"principal_id": principal.ID.String(),
			},
		})
		return err
	}
	return nil
}

func (s Service) auditRecord(input providerProfileAuditInput) (audit.Record, error) {
	return appsupport.NewAuditRecord(s.ids, s.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: inventory.InventoryID(""),
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      input.Action,
		TargetType:  audit.TargetProviderProfile,
		TargetID:    input.Profile.ID.String(),
		Metadata: map[string]string{
			"capability":      input.Profile.Capability.String(),
			"provider_kind":   input.Profile.ProviderKind.String(),
			"lifecycle_state": input.Profile.LifecycleState.String(),
		},
	})
}

func (s Service) recordProviderProfileEvent(ctx context.Context, eventName ports.EventName, principal identity.Principal, profile domain.ProviderProfile) {
	s.observer.Record(ctx, ports.Event{
		Name:    eventName,
		Message: string(eventName),
		Fields: map[string]string{
			"tenant_id":       profile.TenantID.String(),
			"principal_id":    principal.ID.String(),
			"profile_id":      profile.ID.String(),
			"capability":      profile.Capability.String(),
			"provider_kind":   profile.ProviderKind.String(),
			"lifecycle_state": profile.LifecycleState.String(),
		},
	})
}

type providerProfileAuditInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
	Profile   domain.ProviderProfile
	Action    audit.Action
}

package agentmodel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateProviderProfileInput struct {
	Principal          identity.Principal
	Source             audit.Source
	RequestID          string
	TenantID           tenant.ID
	Capability         string
	ProviderKind       string
	DisplayName        string
	EndpointURL        string
	ModelName          string
	RuntimeOptionsJSON []byte
	CapabilityJSON     []byte
	Enable             bool
}

type ListProviderProfilesInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
}

type GetProviderProfileInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
	ProfileID domain.ProviderProfileID
}

type ProviderProfileLifecycleInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
	ProfileID domain.ProviderProfileID
}

func (s Service) CreateProviderProfile(ctx context.Context, input CreateProviderProfileInput) (domain.ProviderProfile, error) {
	if err := s.ensureTenantConfigure(ctx, input.Principal, input.TenantID); err != nil {
		return domain.ProviderProfile{}, err
	}
	if rejectsCredentialMaterial(input.RuntimeOptionsJSON) || rejectsCredentialMaterial(input.CapabilityJSON) {
		return domain.ProviderProfile{}, apperrors.ErrValidation
	}
	profileID, ok := domain.NewProviderProfileID(s.ids.NewID())
	if !ok {
		return domain.ProviderProfile{}, apperrors.ErrValidation
	}
	capability, ok := domain.NewProviderCapability(input.Capability)
	if !ok {
		return domain.ProviderProfile{}, apperrors.ErrValidation
	}
	providerKind, ok := domain.NewProviderKind(input.ProviderKind)
	if !ok {
		return domain.ProviderProfile{}, apperrors.ErrValidation
	}
	displayName, ok := domain.NewDisplayName(input.DisplayName)
	if !ok {
		return domain.ProviderProfile{}, apperrors.ErrValidation
	}
	endpointURL, ok := domain.NewEndpointURL(input.EndpointURL)
	if !ok {
		return domain.ProviderProfile{}, apperrors.ErrValidation
	}
	modelName, ok := domain.NewModelName(input.ModelName)
	if !ok {
		return domain.ProviderProfile{}, apperrors.ErrValidation
	}
	lifecycle := domain.ProviderProfileDisabled
	if input.Enable {
		lifecycle = domain.ProviderProfileEnabled
	}
	now := s.clock.Now()
	profile, ok := domain.NewProviderProfile(domain.ProviderProfileInput{
		ID:                 profileID,
		TenantID:           domain.TenantID(input.TenantID.String()),
		Capability:         capability,
		ProviderKind:       providerKind,
		DisplayName:        displayName,
		EndpointURL:        endpointURL,
		ModelName:          modelName,
		RuntimeOptionsJSON: input.RuntimeOptionsJSON,
		CapabilityJSON:     input.CapabilityJSON,
		CredentialStatus:   domain.CredentialStatusMissing,
		LifecycleState:     lifecycle,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if !ok {
		return domain.ProviderProfile{}, apperrors.ErrValidation
	}
	auditRecord, err := s.auditRecord(providerProfileAuditInput{
		Principal: input.Principal,
		Source:    input.Source,
		RequestID: input.RequestID,
		TenantID:  input.TenantID,
		Profile:   profile,
		Action:    audit.ActionProviderProfileCreated,
	})
	if err != nil {
		return domain.ProviderProfile{}, err
	}
	if err := s.providerProfileUnitOfWork.SaveProviderProfile(ctx, profile, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return domain.ProviderProfile{}, apperrors.ErrConflict
		}
		return domain.ProviderProfile{}, err
	}
	s.recordProviderProfileEvent(ctx, ports.EventProviderProfileCreated, input.Principal, profile)
	return profile, nil
}

func (s Service) ListProviderProfiles(ctx context.Context, input ListProviderProfilesInput) ([]domain.ProviderProfile, error) {
	if err := s.ensureTenantConfigure(ctx, input.Principal, input.TenantID); err != nil {
		return nil, err
	}
	profiles, err := s.providerProfiles.ListProviderProfiles(ctx, input.TenantID)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (s Service) GetProviderProfile(ctx context.Context, input GetProviderProfileInput) (domain.ProviderProfile, error) {
	if err := s.ensureTenantConfigure(ctx, input.Principal, input.TenantID); err != nil {
		return domain.ProviderProfile{}, err
	}
	profile, found, err := s.providerProfiles.ProviderProfileByID(ctx, input.TenantID, input.ProfileID)
	if err != nil {
		return domain.ProviderProfile{}, err
	}
	if !found {
		return domain.ProviderProfile{}, apperrors.ErrNotFound
	}
	s.recordProviderProfileEvent(ctx, ports.EventProviderProfileViewed, input.Principal, profile)
	return profile, nil
}

func (s Service) EnableProviderProfile(ctx context.Context, input ProviderProfileLifecycleInput) (domain.ProviderProfile, error) {
	return s.updateLifecycle(ctx, input, audit.ActionProviderProfileEnabled, ports.EventProviderProfileEnabled, func(profile domain.ProviderProfile) (domain.ProviderProfile, bool) {
		return profile.Enable(s.clock.Now())
	})
}

func (s Service) DisableProviderProfile(ctx context.Context, input ProviderProfileLifecycleInput) (domain.ProviderProfile, error) {
	return s.updateLifecycle(ctx, input, audit.ActionProviderProfileDisabled, ports.EventProviderProfileDisabled, func(profile domain.ProviderProfile) (domain.ProviderProfile, bool) {
		return profile.Disable(s.clock.Now())
	})
}

func (s Service) ArchiveProviderProfile(ctx context.Context, input ProviderProfileLifecycleInput) (domain.ProviderProfile, error) {
	return s.updateLifecycle(ctx, input, audit.ActionProviderProfileArchived, ports.EventProviderProfileArchived, func(profile domain.ProviderProfile) (domain.ProviderProfile, bool) {
		return profile.Archive(s.clock.Now())
	})
}

func (s Service) updateLifecycle(ctx context.Context, input ProviderProfileLifecycleInput, action audit.Action, eventName ports.EventName, transition func(domain.ProviderProfile) (domain.ProviderProfile, bool)) (domain.ProviderProfile, error) {
	if err := s.ensureTenantConfigure(ctx, input.Principal, input.TenantID); err != nil {
		return domain.ProviderProfile{}, err
	}
	current, found, err := s.providerProfiles.ProviderProfileByID(ctx, input.TenantID, input.ProfileID)
	if err != nil {
		return domain.ProviderProfile{}, err
	}
	if !found {
		return domain.ProviderProfile{}, apperrors.ErrNotFound
	}
	updated, ok := transition(current)
	if !ok {
		return domain.ProviderProfile{}, apperrors.ErrPrecondition
	}
	auditRecord, err := s.auditRecord(providerProfileAuditInput{
		Principal: input.Principal,
		Source:    input.Source,
		RequestID: input.RequestID,
		TenantID:  input.TenantID,
		Profile:   updated,
		Action:    action,
	})
	if err != nil {
		return domain.ProviderProfile{}, err
	}
	if err := s.providerProfileUnitOfWork.UpdateProviderProfile(ctx, updated, auditRecord); err != nil {
		return domain.ProviderProfile{}, err
	}
	s.recordProviderProfileEvent(ctx, eventName, input.Principal, updated)
	return updated, nil
}

func rejectsCredentialMaterial(raw []byte) bool {
	if len(bytes.TrimSpace(raw)) == 0 {
		return false
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return false
	}
	for key := range decoded {
		normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
		if slices.Contains([]string{"apikey", "apisecret", "token", "accesstoken", "bearertoken", "credential", "credentials", "secret", "password"}, normalized) {
			return true
		}
	}
	return false
}

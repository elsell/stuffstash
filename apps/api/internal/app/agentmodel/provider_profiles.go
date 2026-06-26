package agentmodel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

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
	PromptTemplate     string
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

type ReplaceProviderProfileCredentialInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
	ProfileID domain.ProviderProfileID
	Purpose   string
	Raw       []byte
}

type TestProviderProfileInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
	ProfileID domain.ProviderProfileID
}

var credentialMaterialMarkers = []string{"apikey", "apisecret", "token", "accesstoken", "bearertoken", "credential", "credentials", "secret", "password"}

func (s Service) CreateProviderProfile(ctx context.Context, input CreateProviderProfileInput) (domain.ProviderProfile, error) {
	if err := s.ensureTenantConfigure(ctx, input.Principal, input.TenantID); err != nil {
		return domain.ProviderProfile{}, err
	}
	if rejectsCredentialMaterial(input.RuntimeOptionsJSON) || rejectsCredentialMaterial(input.CapabilityJSON) || rejectsCredentialText(input.PromptTemplate) {
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
		PromptTemplate:     input.PromptTemplate,
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

func (s Service) ReplaceProviderProfileCredential(ctx context.Context, input ReplaceProviderProfileCredentialInput) (domain.ProviderProfile, error) {
	if err := s.ensureTenantConfigure(ctx, input.Principal, input.TenantID); err != nil {
		return domain.ProviderProfile{}, err
	}
	if s.providerCredentialSealer == nil {
		return domain.ProviderProfile{}, apperrors.ErrPrecondition
	}
	purpose, ok := ports.NewProviderCredentialPurpose(input.Purpose)
	if !ok || len(bytes.TrimSpace(input.Raw)) == 0 {
		return domain.ProviderProfile{}, apperrors.ErrValidation
	}
	current, found, err := s.providerProfiles.ProviderProfileByID(ctx, input.TenantID, input.ProfileID)
	if err != nil {
		return domain.ProviderProfile{}, err
	}
	if !found {
		return domain.ProviderProfile{}, apperrors.ErrNotFound
	}
	updated, ok := current.WithCredentialConfigured(s.clock.Now())
	if !ok {
		return domain.ProviderProfile{}, apperrors.ErrPrecondition
	}
	scope := ports.ProviderCredentialScope{
		TenantID:          input.TenantID,
		ProviderProfileID: current.ID.String(),
		Capability:        ports.ProviderCapability(current.Capability.String()),
		ProviderKind:      ports.ProviderKind(current.ProviderKind.String()),
		Purpose:           purpose,
	}
	sealed, err := s.providerCredentialSealer.SealProviderCredential(ctx, scope, input.Raw)
	if err != nil {
		return domain.ProviderProfile{}, apperrors.ErrValidation
	}
	credentialID := strings.TrimSpace(s.ids.NewID())
	if credentialID == "" {
		return domain.ProviderProfile{}, apperrors.ErrValidation
	}
	credential := ports.ProviderCredentialRecord{
		ID:        credentialID,
		Scope:     scope,
		Sealed:    sealed,
		CreatedAt: updated.UpdatedAt,
		UpdatedAt: updated.UpdatedAt,
	}
	auditRecord, err := s.auditRecord(providerProfileAuditInput{
		Principal: input.Principal,
		Source:    input.Source,
		RequestID: input.RequestID,
		TenantID:  input.TenantID,
		Profile:   updated,
		Action:    audit.ActionProviderProfileCredentialReplaced,
	})
	if err != nil {
		return domain.ProviderProfile{}, err
	}
	if err := s.providerProfileUnitOfWork.ReplaceProviderProfileCredential(ctx, updated, credential, auditRecord); err != nil {
		return domain.ProviderProfile{}, err
	}
	s.recordProviderProfileEvent(ctx, ports.EventProviderProfileCredentialReplaced, input.Principal, updated)
	return updated, nil
}

func (s Service) TestProviderProfile(ctx context.Context, input TestProviderProfileInput) (ports.ProviderProfileTestResult, error) {
	if err := s.ensureTenantConfigure(ctx, input.Principal, input.TenantID); err != nil {
		return ports.ProviderProfileTestResult{}, err
	}
	if s.providerCredentials == nil || s.providerCredentialSealer == nil || s.providerProfileTester == nil {
		return ports.ProviderProfileTestResult{}, apperrors.ErrPrecondition
	}
	current, found, err := s.providerProfiles.ProviderProfileByID(ctx, input.TenantID, input.ProfileID)
	if err != nil {
		return ports.ProviderProfileTestResult{}, err
	}
	if !found {
		return ports.ProviderProfileTestResult{}, apperrors.ErrNotFound
	}
	if current.LifecycleState == domain.ProviderProfileArchived || current.CredentialStatus != domain.CredentialStatusConfigured {
		return ports.ProviderProfileTestResult{}, apperrors.ErrPrecondition
	}
	purpose, raw, err := s.activeProviderCredential(ctx, input.TenantID, current)
	if err != nil {
		return ports.ProviderProfileTestResult{}, err
	}
	now := s.clock.Now()
	result, err := s.providerProfileTester.TestProviderProfile(ctx, ports.ProviderProfileTestInput{
		Profile:           current,
		CredentialPurpose: purpose,
		Credential:        raw,
		TestedAt:          now,
	})
	result = safeProviderProfileTestResult(current, result, err, now)
	profileForAudit := current
	if result.Status == ports.ProviderProfileTestStatusSucceeded {
		updated, ok := current.WithLastTested(now)
		if !ok {
			return ports.ProviderProfileTestResult{}, apperrors.ErrPrecondition
		}
		profileForAudit = updated
	}
	auditRecord, err := s.auditRecord(providerProfileAuditInput{
		Principal: input.Principal,
		Source:    input.Source,
		RequestID: input.RequestID,
		TenantID:  input.TenantID,
		Profile:   profileForAudit,
		Action:    audit.ActionProviderProfileTested,
	})
	if err != nil {
		return ports.ProviderProfileTestResult{}, err
	}
	if err := s.providerProfileUnitOfWork.UpdateProviderProfile(ctx, profileForAudit, auditRecord); err != nil {
		return ports.ProviderProfileTestResult{}, err
	}
	s.recordProviderProfileEvent(ctx, ports.EventProviderProfileTested, input.Principal, profileForAudit)
	return result, nil
}

func safeProviderProfileTestResult(profile domain.ProviderProfile, result ports.ProviderProfileTestResult, testErr error, testedAt time.Time) ports.ProviderProfileTestResult {
	status := result.Status
	message := strings.TrimSpace(result.Message)
	if testErr != nil || status != ports.ProviderProfileTestStatusSucceeded {
		status = ports.ProviderProfileTestStatusFailed
		message = "Provider profile test failed safely. Check the profile configuration and credential."
	} else if message == "" {
		message = "Provider profile test succeeded."
	}
	return ports.ProviderProfileTestResult{
		ProfileID:    profile.ID.String(),
		Capability:   profile.Capability.String(),
		ProviderKind: profile.ProviderKind.String(),
		Status:       status,
		Message:      message,
		TestedAt:     testedAt,
	}
}

func (s Service) activeProviderCredential(ctx context.Context, tenantID tenant.ID, profile domain.ProviderProfile) (ports.ProviderCredentialPurpose, []byte, error) {
	for _, purpose := range providerCredentialPurposes(profile) {
		scope := ports.ProviderCredentialScope{
			TenantID:          tenantID,
			ProviderProfileID: profile.ID.String(),
			Capability:        ports.ProviderCapability(profile.Capability.String()),
			ProviderKind:      ports.ProviderKind(profile.ProviderKind.String()),
			Purpose:           purpose,
		}
		record, found, err := s.providerCredentials.ActiveProviderCredential(ctx, scope)
		if err != nil {
			return "", nil, err
		}
		if !found {
			continue
		}
		raw, err := s.providerCredentialSealer.UnsealProviderCredential(ctx, scope, record.Sealed)
		if err != nil || len(bytes.TrimSpace(raw)) == 0 {
			return "", nil, apperrors.ErrPrecondition
		}
		return purpose, raw, nil
	}
	return "", nil, apperrors.ErrPrecondition
}

func providerCredentialPurposes(profile domain.ProviderProfile) []ports.ProviderCredentialPurpose {
	if profile.ProviderKind == domain.ProviderKindGemini {
		if profile.Capability == domain.ProviderCapabilityTextToSpeech {
			return []ports.ProviderCredentialPurpose{ports.ProviderCredentialPurposeOAuthBearer}
		}
		return []ports.ProviderCredentialPurpose{ports.ProviderCredentialPurposeAPIKey, ports.ProviderCredentialPurposeOAuthBearer}
	}
	return []ports.ProviderCredentialPurpose{ports.ProviderCredentialPurposeAPIKey, ports.ProviderCredentialPurposeOAuthBearer}
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
		if slices.Contains(credentialMaterialMarkers, normalized) {
			return true
		}
	}
	return false
}

func rejectsCredentialText(value string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(value, "_", ""), "-", ""))
	normalized = strings.ReplaceAll(normalized, " ", "")
	for _, marker := range credentialMaterialMarkers {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

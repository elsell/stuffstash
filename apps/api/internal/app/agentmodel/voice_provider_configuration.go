package agentmodel

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const (
	VoiceProviderReadinessReady          = "ready"
	VoiceProviderReadinessNeedsAttention = "needs_attention"

	VoiceProviderSelectionExplicit = "explicit"
	VoiceProviderSelectionImplicit = "implicit"
	VoiceProviderSelectionMissing  = "missing"

	VoiceProviderSlotReady               = "ready"
	VoiceProviderSlotMissing             = "missing"
	VoiceProviderSlotDisabled            = "disabled"
	VoiceProviderSlotArchived            = "archived"
	VoiceProviderSlotCredentialMissing   = "credential_missing"
	VoiceProviderSlotUntested            = "untested"
	VoiceProviderSlotDuplicateCandidates = "duplicate_candidates"
	VoiceProviderSlotInvalidSelection    = "invalid_selection"
)

type VoiceProviderConfigurationInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
}

type UpdateVoiceProviderConfigurationInput struct {
	Principal                  identity.Principal
	Source                     audit.Source
	RequestID                  string
	TenantID                   tenant.ID
	SpeechToTextProfileID      string
	LanguageInferenceProfileID string
	TextToSpeechProfileID      string
}

type VoiceProviderConfigurationView struct {
	TenantID   string
	Readiness  string
	UpdatedAt  string
	Slots      []VoiceProviderSlotView
	ProfileIDs VoiceProviderConfigurationProfileIDs
}

type VoiceProviderConfigurationProfileIDs struct {
	SpeechToText      string
	LanguageInference string
	TextToSpeech      string
}

type VoiceProviderSlotView struct {
	Capability        string
	Label             string
	SelectedProfile   *ProviderProfileSummary
	SelectedProfileID string
	SelectionSource   string
	Readiness         string
	Issues            []string
	RecommendedAction string
	DuplicateProfiles []ProviderProfileSummary
}

type ProviderProfileSummary struct {
	ID                string
	Capability        string
	ProviderKind      string
	DisplayName       string
	ModelName         string
	CredentialStatus  string
	CredentialPurpose string
	LifecycleState    string
	LastTestedAt      string
}

func (s Service) GetVoiceProviderConfiguration(ctx context.Context, input VoiceProviderConfigurationInput) (VoiceProviderConfigurationView, error) {
	if err := s.ensureTenantConfigure(ctx, input.Principal, input.TenantID); err != nil {
		return VoiceProviderConfigurationView{}, err
	}
	return s.voiceProviderConfigurationView(ctx, input.TenantID)
}

func (s Service) UpdateVoiceProviderConfiguration(ctx context.Context, input UpdateVoiceProviderConfigurationInput) (VoiceProviderConfigurationView, error) {
	if err := s.ensureTenantConfigure(ctx, input.Principal, input.TenantID); err != nil {
		return VoiceProviderConfigurationView{}, err
	}
	if s.voiceProviderConfigs == nil {
		return VoiceProviderConfigurationView{}, apperrors.ErrPrecondition
	}
	profiles, err := s.providerProfiles.ListProviderProfiles(ctx, input.TenantID)
	if err != nil {
		return VoiceProviderConfigurationView{}, err
	}
	if err := validateVoiceProviderSelection(profiles, domain.ProviderCapabilitySpeechToText, input.SpeechToTextProfileID); err != nil {
		return VoiceProviderConfigurationView{}, err
	}
	if err := validateVoiceProviderSelection(profiles, domain.ProviderCapabilityLanguageInference, input.LanguageInferenceProfileID); err != nil {
		return VoiceProviderConfigurationView{}, err
	}
	if err := validateVoiceProviderSelection(profiles, domain.ProviderCapabilityTextToSpeech, input.TextToSpeechProfileID); err != nil {
		return VoiceProviderConfigurationView{}, err
	}
	now := s.clock.Now()
	existing, found, err := s.voiceProviderConfigs.VoiceProviderConfiguration(ctx, input.TenantID)
	if err != nil {
		return VoiceProviderConfigurationView{}, err
	}
	createdAt := now
	if found && !existing.CreatedAt.IsZero() {
		createdAt = existing.CreatedAt
	}
	record := ports.VoiceProviderConfigurationRecord{
		TenantID:                   input.TenantID,
		SpeechToTextProfileID:      strings.TrimSpace(input.SpeechToTextProfileID),
		LanguageInferenceProfileID: strings.TrimSpace(input.LanguageInferenceProfileID),
		TextToSpeechProfileID:      strings.TrimSpace(input.TextToSpeechProfileID),
		CreatedAt:                  createdAt,
		UpdatedAt:                  now,
	}
	auditRecord, err := appsupport.NewAuditRecord(s.ids, s.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: inventory.InventoryID(""),
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionVoiceProviderConfigurationUpdated,
		TargetType:  audit.TargetTenant,
		TargetID:    input.TenantID.String(),
		Metadata: map[string]string{
			"speech_to_text_profile_id":     record.SpeechToTextProfileID,
			"language_inference_profile_id": record.LanguageInferenceProfileID,
			"text_to_speech_profile_id":     record.TextToSpeechProfileID,
		},
	})
	if err != nil {
		return VoiceProviderConfigurationView{}, err
	}
	if err := s.voiceProviderConfigs.SaveVoiceProviderConfiguration(ctx, record, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return VoiceProviderConfigurationView{}, apperrors.ErrConflict
		}
		return VoiceProviderConfigurationView{}, err
	}
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventVoiceProviderConfigurationUpdated,
		Message: string(ports.EventVoiceProviderConfigurationUpdated),
		Fields: map[string]string{
			"tenant_id":                     input.TenantID.String(),
			"principal_id":                  input.Principal.ID.String(),
			"speech_to_text_profile_id":     record.SpeechToTextProfileID,
			"language_inference_profile_id": record.LanguageInferenceProfileID,
			"text_to_speech_profile_id":     record.TextToSpeechProfileID,
		},
	})
	return s.voiceProviderConfigurationViewFrom(profiles, record, true), nil
}

func validateVoiceProviderSelection(profiles []domain.ProviderProfile, capability domain.ProviderCapability, profileID string) error {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return nil
	}
	for _, profile := range profiles {
		if profile.ID.String() == profileID {
			if profile.Capability != capability || profile.LifecycleState == domain.ProviderProfileArchived {
				return apperrors.ErrValidation
			}
			return nil
		}
	}
	return apperrors.ErrValidation
}

func (s Service) voiceProviderConfigurationView(ctx context.Context, tenantID tenant.ID) (VoiceProviderConfigurationView, error) {
	profiles, err := s.providerProfiles.ListProviderProfiles(ctx, tenantID)
	if err != nil {
		return VoiceProviderConfigurationView{}, err
	}
	record := ports.VoiceProviderConfigurationRecord{TenantID: tenantID}
	explicit := false
	if s.voiceProviderConfigs != nil {
		existing, found, err := s.voiceProviderConfigs.VoiceProviderConfiguration(ctx, tenantID)
		if err != nil {
			return VoiceProviderConfigurationView{}, err
		}
		if found {
			record = existing
			explicit = true
		}
	}
	return s.voiceProviderConfigurationViewFrom(profiles, record, explicit), nil
}

func (s Service) voiceProviderConfigurationViewFrom(profiles []domain.ProviderProfile, record ports.VoiceProviderConfigurationRecord, explicit bool) VoiceProviderConfigurationView {
	slots := []VoiceProviderSlotView{
		voiceProviderSlotView(domain.ProviderCapabilitySpeechToText, "Speech input", record.SpeechToTextProfileID, explicit, profiles),
		voiceProviderSlotView(domain.ProviderCapabilityLanguageInference, "Agent brain", record.LanguageInferenceProfileID, explicit, profiles),
		voiceProviderSlotView(domain.ProviderCapabilityTextToSpeech, "Spoken output", record.TextToSpeechProfileID, explicit, profiles),
	}
	readiness := VoiceProviderReadinessReady
	for _, slot := range slots {
		if slot.Readiness != VoiceProviderSlotReady {
			readiness = VoiceProviderReadinessNeedsAttention
			break
		}
	}
	return VoiceProviderConfigurationView{
		TenantID:  record.TenantID.String(),
		Readiness: readiness,
		UpdatedAt: safeTime(record.UpdatedAt),
		Slots:     slots,
		ProfileIDs: VoiceProviderConfigurationProfileIDs{
			SpeechToText:      slots[0].SelectedProfileID,
			LanguageInference: slots[1].SelectedProfileID,
			TextToSpeech:      slots[2].SelectedProfileID,
		},
	}
}

func voiceProviderSlotView(capability domain.ProviderCapability, label string, selectedID string, explicit bool, profiles []domain.ProviderProfile) VoiceProviderSlotView {
	candidates := profilesForCapability(profiles, capability)
	selected, source := selectVoiceProviderProfile(candidates, selectedID, explicit)
	slot := VoiceProviderSlotView{
		Capability:        capability.String(),
		Label:             label,
		SelectionSource:   source,
		Readiness:         VoiceProviderSlotMissing,
		RecommendedAction: "add_profile",
	}
	eligible := readyVoiceProviderCandidates(candidates)
	if len(eligible) > 1 {
		slot.DuplicateProfiles = providerProfileSummaries(eligible)
	}
	if source == VoiceProviderSelectionImplicit && len(eligible) > 1 {
		slot.Readiness = VoiceProviderSlotDuplicateCandidates
		slot.Issues = []string{"Multiple ready profiles exist. Choose one explicitly."}
		slot.RecommendedAction = "choose_profile"
		if len(eligible) > 0 {
			summary := providerProfileSummary(eligible[0])
			slot.SelectedProfile = &summary
			slot.SelectedProfileID = summary.ID
		}
		return slot
	}
	if selected == nil {
		if strings.TrimSpace(selectedID) != "" {
			slot.Readiness = VoiceProviderSlotInvalidSelection
			slot.Issues = []string{"The selected profile is missing or no longer matches this slot."}
			slot.RecommendedAction = "choose_profile"
		} else {
			slot.Issues = []string{"No profile is selected for this voice slot."}
		}
		return slot
	}
	summary := providerProfileSummary(*selected)
	slot.SelectedProfile = &summary
	slot.SelectedProfileID = summary.ID
	slot.Readiness, slot.Issues, slot.RecommendedAction = readinessForProviderProfile(*selected)
	return slot
}

func selectVoiceProviderProfile(candidates []domain.ProviderProfile, selectedID string, explicit bool) (*domain.ProviderProfile, string) {
	selectedID = strings.TrimSpace(selectedID)
	if selectedID != "" {
		for _, profile := range candidates {
			if profile.ID.String() == selectedID {
				copy := profile
				return &copy, VoiceProviderSelectionExplicit
			}
		}
		return nil, VoiceProviderSelectionExplicit
	}
	eligible := readyVoiceProviderCandidates(candidates)
	if len(eligible) > 0 {
		copy := eligible[0]
		return &copy, VoiceProviderSelectionImplicit
	}
	diagnosticCandidates := nonArchivedProviderProfiles(candidates)
	if len(diagnosticCandidates) > 0 {
		copy := diagnosticCandidates[0]
		return &copy, VoiceProviderSelectionImplicit
	}
	if explicit {
		return nil, VoiceProviderSelectionExplicit
	}
	return nil, VoiceProviderSelectionMissing
}

func readinessForProviderProfile(profile domain.ProviderProfile) (string, []string, string) {
	switch {
	case profile.LifecycleState == domain.ProviderProfileDisabled:
		return VoiceProviderSlotDisabled, []string{"The selected profile is disabled."}, "enable_profile"
	case profile.LifecycleState == domain.ProviderProfileArchived:
		return VoiceProviderSlotArchived, []string{"The selected profile is archived."}, "choose_profile"
	case profile.CredentialStatus != domain.CredentialStatusConfigured:
		return VoiceProviderSlotCredentialMissing, []string{"The selected profile needs a credential."}, "replace_credential"
	case profile.LastTestedAt == nil:
		return VoiceProviderSlotUntested, []string{"The selected profile needs a successful test."}, "test_profile"
	default:
		return VoiceProviderSlotReady, nil, "none"
	}
}

func profilesForCapability(profiles []domain.ProviderProfile, capability domain.ProviderCapability) []domain.ProviderProfile {
	matches := []domain.ProviderProfile{}
	for _, profile := range profiles {
		if profile.Capability == capability {
			matches = append(matches, profile)
		}
	}
	sort.Slice(matches, func(left, right int) bool {
		if matches[left].CreatedAt.Equal(matches[right].CreatedAt) {
			return matches[left].ID.String() < matches[right].ID.String()
		}
		return matches[left].CreatedAt.Before(matches[right].CreatedAt)
	})
	return matches
}

func nonArchivedProviderProfiles(profiles []domain.ProviderProfile) []domain.ProviderProfile {
	filtered := []domain.ProviderProfile{}
	for _, profile := range profiles {
		if profile.LifecycleState != domain.ProviderProfileArchived {
			filtered = append(filtered, profile)
		}
	}
	return filtered
}

func readyVoiceProviderCandidates(profiles []domain.ProviderProfile) []domain.ProviderProfile {
	ready := []domain.ProviderProfile{}
	for _, profile := range profiles {
		if profile.LifecycleState == domain.ProviderProfileEnabled && profile.CredentialStatus == domain.CredentialStatusConfigured && profile.LastTestedAt != nil {
			ready = append(ready, profile)
		}
	}
	return ready
}

func providerProfileSummaries(profiles []domain.ProviderProfile) []ProviderProfileSummary {
	summaries := make([]ProviderProfileSummary, 0, len(profiles))
	for _, profile := range profiles {
		summaries = append(summaries, providerProfileSummary(profile))
	}
	return summaries
}

func providerProfileSummary(profile domain.ProviderProfile) ProviderProfileSummary {
	return ProviderProfileSummary{
		ID:                profile.ID.String(),
		Capability:        profile.Capability.String(),
		ProviderKind:      profile.ProviderKind.String(),
		DisplayName:       profile.DisplayName.String(),
		ModelName:         profile.ModelName.String(),
		CredentialStatus:  profile.CredentialStatus.String(),
		CredentialPurpose: providerProfileSummaryCredentialPurpose(profile),
		LifecycleState:    profile.LifecycleState.String(),
		LastTestedAt:      safeTimePtr(profile.LastTestedAt),
	}
}

func providerProfileSummaryCredentialPurpose(profile domain.ProviderProfile) string {
	options := map[string]any{}
	_ = json.Unmarshal([]byte(profile.RuntimeOptionsJSON.String()), &options)
	credentialType, _ := options["credentialType"].(string)
	switch strings.TrimSpace(credentialType) {
	case string(ports.ProviderCredentialPurposeServerADC):
		return string(ports.ProviderCredentialPurposeServerADC)
	case string(ports.ProviderCredentialPurposeAPIKey):
		return string(ports.ProviderCredentialPurposeAPIKey)
	case string(ports.ProviderCredentialPurposeOAuthBearer):
		return string(ports.ProviderCredentialPurposeOAuthBearer)
	}
	if profile.ProviderKind == domain.ProviderKindGemini {
		if profile.Capability == domain.ProviderCapabilityTextToSpeech {
			return string(ports.ProviderCredentialPurposeOAuthBearer)
		}
		return string(ports.ProviderCredentialPurposeAPIKey)
	}
	return ""
}

func safeTimePtr(value *time.Time) string {
	if value == nil {
		return ""
	}
	return safeTime(*value)
}

func safeTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

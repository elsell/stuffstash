package gormstore

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

type providerProfileModel struct {
	ID                 string      `gorm:"primaryKey;size:64"`
	TenantID           string      `gorm:"not null;size:26;index"`
	Tenant             tenantModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	Capability         string      `gorm:"not null;size:64;index;check:chk_provider_profiles_capability,capability IN ('speech_to_text','language_inference','text_to_speech')"`
	ProviderKind       string      `gorm:"not null;size:64;index;check:chk_provider_profiles_provider_kind,provider_kind IN ('gemini','openai_compatible','local_http')"`
	DisplayName        string      `gorm:"not null;size:120"`
	EndpointURL        string      `gorm:"not null;default:'';size:2048"`
	ModelName          string      `gorm:"not null;default:'';size:256"`
	RuntimeOptionsJSON []byte      `gorm:"not null"`
	CapabilityJSON     []byte      `gorm:"not null"`
	CredentialStatus   string      `gorm:"not null;size:64;index;check:chk_provider_profiles_credential_status,credential_status IN ('missing','configured')"`
	LifecycleState     string      `gorm:"not null;size:64;index;check:chk_provider_profiles_lifecycle_state,lifecycle_state IN ('enabled','disabled','archived')"`
	LastTestedAt       *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (providerProfileModel) TableName() string {
	return "provider_profiles"
}

func (m providerProfileModel) toDomain() (agentmodel.ProviderProfile, bool) {
	return agentmodel.NewProviderProfile(agentmodel.ProviderProfileInput{
		ID:                 agentmodel.ProviderProfileID(m.ID),
		TenantID:           agentmodel.TenantID(m.TenantID),
		Capability:         agentmodel.ProviderCapability(m.Capability),
		ProviderKind:       agentmodel.ProviderKind(m.ProviderKind),
		DisplayName:        agentmodel.DisplayName(m.DisplayName),
		EndpointURL:        agentmodel.EndpointURL(m.EndpointURL),
		ModelName:          agentmodel.ModelName(m.ModelName),
		RuntimeOptionsJSON: append([]byte{}, m.RuntimeOptionsJSON...),
		CapabilityJSON:     append([]byte{}, m.CapabilityJSON...),
		CredentialStatus:   agentmodel.CredentialStatus(m.CredentialStatus),
		LifecycleState:     agentmodel.ProviderProfileLifecycleState(m.LifecycleState),
		LastTestedAt:       m.LastTestedAt,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	})
}

func providerProfileModelFromDomain(profile agentmodel.ProviderProfile) providerProfileModel {
	return providerProfileModel{
		ID:                 profile.ID.String(),
		TenantID:           profile.TenantID.String(),
		Capability:         profile.Capability.String(),
		ProviderKind:       profile.ProviderKind.String(),
		DisplayName:        profile.DisplayName.String(),
		EndpointURL:        profile.EndpointURL.String(),
		ModelName:          profile.ModelName.String(),
		RuntimeOptionsJSON: []byte(profile.RuntimeOptionsJSON.String()),
		CapabilityJSON:     []byte(profile.CapabilityJSON.String()),
		CredentialStatus:   profile.CredentialStatus.String(),
		LifecycleState:     profile.LifecycleState.String(),
		LastTestedAt:       profile.LastTestedAt,
		CreatedAt:          profile.CreatedAt,
		UpdatedAt:          profile.UpdatedAt,
	}
}

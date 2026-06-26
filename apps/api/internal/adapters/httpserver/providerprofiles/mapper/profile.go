package mapper

import (
	"encoding/json"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/providerprofiles/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func ProviderProfileToResponse(profile agentmodel.ProviderProfile) dto.ProviderProfileResponse {
	var lastTestedAt *string
	if profile.LastTestedAt != nil {
		value := profile.LastTestedAt.UTC().Format(time.RFC3339Nano)
		lastTestedAt = &value
	}
	return dto.ProviderProfileResponse{
		ID:                 profile.ID.String(),
		TenantID:           profile.TenantID.String(),
		Capability:         profile.Capability.String(),
		ProviderKind:       profile.ProviderKind.String(),
		DisplayName:        profile.DisplayName.String(),
		EndpointURL:        profile.EndpointURL.String(),
		ModelName:          profile.ModelName.String(),
		RuntimeOptions:     jsonObject(profile.RuntimeOptionsJSON.String()),
		CapabilityMetadata: jsonObject(profile.CapabilityJSON.String()),
		PromptTemplate:     profile.PromptTemplate.String(),
		CredentialStatus:   profile.CredentialStatus.String(),
		LifecycleState:     profile.LifecycleState.String(),
		CreatedAt:          profile.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:          profile.UpdatedAt.UTC().Format(time.RFC3339Nano),
		LastTestedAt:       lastTestedAt,
	}
}

func ProviderProfilesToResponse(profiles []agentmodel.ProviderProfile) []dto.ProviderProfileResponse {
	data := make([]dto.ProviderProfileResponse, 0, len(profiles))
	for _, profile := range profiles {
		data = append(data, ProviderProfileToResponse(profile))
	}
	return data
}

func ProviderProfileTestToResponse(result ports.ProviderProfileTestResult) dto.TestProviderProfileResponse {
	return dto.TestProviderProfileResponse{
		ProviderProfileID: result.ProfileID,
		Capability:        result.Capability,
		ProviderKind:      result.ProviderKind,
		Status:            string(result.Status),
		Message:           result.Message,
		TestedAt:          result.TestedAt.UTC().Format(time.RFC3339Nano),
	}
}

func jsonObject(raw string) map[string]any {
	result := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &result)
	return result
}

package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type CreateProviderProfileInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Body          CreateProviderProfileBody
}

type CreateProviderProfileBody struct {
	Capability         string         `json:"capability" doc:"Provider capability"`
	ProviderKind       string         `json:"providerKind" doc:"Provider adapter kind"`
	DisplayName        string         `json:"displayName" maxLength:"120" doc:"User-facing provider profile name"`
	EndpointURL        string         `json:"endpointUrl,omitempty" maxLength:"2048" doc:"Provider endpoint URL when required"`
	ModelName          string         `json:"modelName,omitempty" maxLength:"256" doc:"Provider model or deployment name"`
	RuntimeOptions     map[string]any `json:"runtimeOptions,omitempty" doc:"Non-secret runtime options"`
	CapabilityMetadata map[string]any `json:"capabilityMetadata,omitempty" doc:"Safe provider capability metadata"`
	Enable             bool           `json:"enable,omitempty" doc:"Create the profile enabled"`
}

type ProviderProfileOutput struct {
	Body shared.SuccessEnvelope[ProviderProfileResponse]
}

type ListProviderProfilesInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
}

type ListProviderProfilesOutput struct {
	Body shared.SuccessEnvelope[[]ProviderProfileResponse]
}

type GetProviderProfileInput struct {
	Authorization     string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID         string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID          string `path:"tenantId" doc:"Tenant ID"`
	ProviderProfileID string `path:"providerProfileId" doc:"Provider profile ID"`
}

type ProviderProfileLifecycleInput struct {
	Authorization     string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID         string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID          string `path:"tenantId" doc:"Tenant ID"`
	ProviderProfileID string `path:"providerProfileId" doc:"Provider profile ID"`
}

type ReplaceProviderProfileCredentialInput struct {
	Authorization     string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID         string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID          string `path:"tenantId" doc:"Tenant ID"`
	ProviderProfileID string `path:"providerProfileId" doc:"Provider profile ID"`
	Body              ReplaceProviderProfileCredentialBody
}

type ReplaceProviderProfileCredentialBody struct {
	Purpose    string `json:"purpose" doc:"Credential purpose"`
	Credential string `json:"credential" doc:"Raw provider credential for this request only"`
}

type ProviderProfileResponse struct {
	ID                 string         `json:"id"`
	TenantID           string         `json:"tenantId"`
	Capability         string         `json:"capability"`
	ProviderKind       string         `json:"providerKind"`
	DisplayName        string         `json:"displayName"`
	EndpointURL        string         `json:"endpointUrl"`
	ModelName          string         `json:"modelName"`
	RuntimeOptions     map[string]any `json:"runtimeOptions"`
	CapabilityMetadata map[string]any `json:"capabilityMetadata"`
	CredentialStatus   string         `json:"credentialStatus"`
	LifecycleState     string         `json:"lifecycleState"`
	CreatedAt          string         `json:"createdAt"`
	UpdatedAt          string         `json:"updatedAt"`
	LastTestedAt       *string        `json:"lastTestedAt,omitempty"`
}

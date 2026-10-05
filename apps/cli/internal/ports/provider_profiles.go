package ports

import "context"

type ProviderProfile struct {
	Capability         string                 `json:"capability"`
	CapabilityMetadata map[string]interface{} `json:"capabilityMetadata"`
	CreatedAt          string                 `json:"createdAt"`
	CredentialStatus   string                 `json:"credentialStatus"`
	DisplayName        string                 `json:"displayName"`
	EndpointURL        string                 `json:"endpointUrl"`
	ID                 string                 `json:"id"`
	LastTestedAt       *string                `json:"lastTestedAt,omitempty"`
	LifecycleState     string                 `json:"lifecycleState"`
	ModelName          string                 `json:"modelName"`
	PromptTemplate     *string                `json:"promptTemplate,omitempty"`
	ProviderKind       string                 `json:"providerKind"`
	RuntimeOptions     map[string]interface{} `json:"runtimeOptions"`
	TenantID           string                 `json:"tenantId"`
	UpdatedAt          string                 `json:"updatedAt"`
}
type ProviderProfilesAPI interface {
	ProviderProfiles(context.Context, string) (Result[[]ProviderProfile], error)
	ProviderProfile(context.Context, string, string) (Result[ProviderProfile], error)
}

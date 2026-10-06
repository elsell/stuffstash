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
type ProviderTest struct {
	ProfileID    string `json:"providerProfileId"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	ProviderKind string `json:"providerKind"`
	Capability   string `json:"capability"`
	TestedAt     string `json:"testedAt"`
}
type ProviderProfilesAPI interface {
	EnableProvider(context.Context, string, string) (Result[ProviderProfile], error)
	DisableProvider(context.Context, string, string) (Result[ProviderProfile], error)
	ArchiveProvider(context.Context, string, string) (Result[ProviderProfile], error)
	TestProvider(context.Context, string, string) (Result[ProviderTest], error)
	ProviderProfiles(context.Context, string) (Result[[]ProviderProfile], error)
	ProviderProfile(context.Context, string, string) (Result[ProviderProfile], error)
}

package ports

import "context"

type VoiceProviderProfiles struct {
	LanguageInference *string `json:"languageInference,omitempty"`
	SpeechToText      *string `json:"speechToText,omitempty"`
	TextToSpeech      *string `json:"textToSpeech,omitempty"`
}
type VoiceProviderSummary struct {
	Capability        string  `json:"capability"`
	CredentialPurpose *string `json:"credentialPurpose,omitempty"`
	CredentialStatus  string  `json:"credentialStatus"`
	DisplayName       string  `json:"displayName"`
	ID                string  `json:"id"`
	LastTestedAt      *string `json:"lastTestedAt,omitempty"`
	LifecycleState    string  `json:"lifecycleState"`
	ModelName         string  `json:"modelName"`
	ProviderKind      string  `json:"providerKind"`
}
type VoiceProviderSlot struct {
	Capability        string                 `json:"capability"`
	DuplicateProfiles []VoiceProviderSummary `json:"duplicateProfiles"`
	Issues            []string               `json:"issues"`
	Label             string                 `json:"label"`
	Readiness         string                 `json:"readiness"`
	RecommendedAction string                 `json:"recommendedAction"`
	SelectedProfile   *VoiceProviderSummary  `json:"selectedProfile,omitempty"`
	SelectedProfileID *string                `json:"selectedProfileId,omitempty"`
	SelectionSource   string                 `json:"selectionSource"`
}
type VoiceProviderConfiguration struct {
	ProfileIDs VoiceProviderProfiles `json:"profileIds"`
	Readiness  string                `json:"readiness"`
	Slots      []VoiceProviderSlot   `json:"slots"`
	TenantID   string                `json:"tenantId"`
	UpdatedAt  *string               `json:"updatedAt,omitempty"`
}
type VoiceProviderAPI interface {
	UpdateVoiceProviderConfiguration(context.Context, string, []byte) (Result[*VoiceProviderConfiguration], error)
	VoiceProviderConfiguration(context.Context, string) (Result[*VoiceProviderConfiguration], error)
}

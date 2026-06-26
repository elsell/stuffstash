package agentmodel

import (
	"encoding/json"
	"strings"
	"time"
)

type ProviderProfileID string

func NewProviderProfileID(value string) (ProviderProfileID, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 64 {
		return "", false
	}
	return ProviderProfileID(value), true
}

func (id ProviderProfileID) String() string {
	return string(id)
}

type TenantID string

func (id TenantID) String() string {
	return string(id)
}

type ProviderCapability string

const (
	ProviderCapabilitySpeechToText      ProviderCapability = "speech_to_text"
	ProviderCapabilityLanguageInference ProviderCapability = "language_inference"
	ProviderCapabilityTextToSpeech      ProviderCapability = "text_to_speech"
)

func NewProviderCapability(value string) (ProviderCapability, bool) {
	switch ProviderCapability(strings.TrimSpace(value)) {
	case ProviderCapabilitySpeechToText:
		return ProviderCapabilitySpeechToText, true
	case ProviderCapabilityLanguageInference:
		return ProviderCapabilityLanguageInference, true
	case ProviderCapabilityTextToSpeech:
		return ProviderCapabilityTextToSpeech, true
	default:
		return "", false
	}
}

func (c ProviderCapability) String() string {
	return string(c)
}

type ProviderKind string

const (
	ProviderKindGemini           ProviderKind = "gemini"
	ProviderKindOpenAICompatible ProviderKind = "openai_compatible"
	ProviderKindLocalHTTP        ProviderKind = "local_http"
)

func NewProviderKind(value string) (ProviderKind, bool) {
	switch ProviderKind(strings.TrimSpace(value)) {
	case ProviderKindGemini:
		return ProviderKindGemini, true
	case ProviderKindOpenAICompatible:
		return ProviderKindOpenAICompatible, true
	case ProviderKindLocalHTTP:
		return ProviderKindLocalHTTP, true
	default:
		return "", false
	}
}

func (k ProviderKind) String() string {
	return string(k)
}

type DisplayName string

func NewDisplayName(value string) (DisplayName, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 120 {
		return "", false
	}
	return DisplayName(value), true
}

func (n DisplayName) String() string {
	return string(n)
}

type EndpointURL string

func NewEndpointURL(value string) (EndpointURL, bool) {
	value = strings.TrimSpace(value)
	if len(value) > 2048 {
		return "", false
	}
	return EndpointURL(value), true
}

func (u EndpointURL) String() string {
	return string(u)
}

type ModelName string

func NewModelName(value string) (ModelName, bool) {
	value = strings.TrimSpace(value)
	if len(value) > 256 {
		return "", false
	}
	return ModelName(value), true
}

func (n ModelName) String() string {
	return string(n)
}

type PromptTemplate string

func NewPromptTemplate(value string) (PromptTemplate, bool) {
	value = strings.TrimSpace(value)
	if len(value) > 8192 {
		return "", false
	}
	return PromptTemplate(value), true
}

func (t PromptTemplate) String() string {
	return string(t)
}

type JSONObject string

func NewJSONObject(raw []byte) (JSONObject, bool) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		trimmed = "{}"
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return "", false
	}
	return JSONObject(trimmed), true
}

func (o JSONObject) String() string {
	if strings.TrimSpace(string(o)) == "" {
		return "{}"
	}
	return string(o)
}

type CredentialStatus string

const (
	CredentialStatusMissing    CredentialStatus = "missing"
	CredentialStatusConfigured CredentialStatus = "configured"
)

func NewCredentialStatus(value string) (CredentialStatus, bool) {
	if strings.TrimSpace(value) == "" {
		return CredentialStatusMissing, true
	}
	switch CredentialStatus(strings.TrimSpace(value)) {
	case CredentialStatusMissing:
		return CredentialStatusMissing, true
	case CredentialStatusConfigured:
		return CredentialStatusConfigured, true
	default:
		return "", false
	}
}

func (s CredentialStatus) String() string {
	return string(s)
}

type ProviderProfileLifecycleState string

const (
	ProviderProfileEnabled  ProviderProfileLifecycleState = "enabled"
	ProviderProfileDisabled ProviderProfileLifecycleState = "disabled"
	ProviderProfileArchived ProviderProfileLifecycleState = "archived"
)

func NewProviderProfileLifecycleState(value string) (ProviderProfileLifecycleState, bool) {
	if strings.TrimSpace(value) == "" {
		return ProviderProfileDisabled, true
	}
	switch ProviderProfileLifecycleState(strings.TrimSpace(value)) {
	case ProviderProfileEnabled:
		return ProviderProfileEnabled, true
	case ProviderProfileDisabled:
		return ProviderProfileDisabled, true
	case ProviderProfileArchived:
		return ProviderProfileArchived, true
	default:
		return "", false
	}
}

func (s ProviderProfileLifecycleState) String() string {
	return string(s)
}

type ProviderProfile struct {
	ID                 ProviderProfileID
	TenantID           TenantID
	Capability         ProviderCapability
	ProviderKind       ProviderKind
	DisplayName        DisplayName
	EndpointURL        EndpointURL
	ModelName          ModelName
	RuntimeOptionsJSON JSONObject
	CapabilityJSON     JSONObject
	PromptTemplate     PromptTemplate
	CredentialStatus   CredentialStatus
	LifecycleState     ProviderProfileLifecycleState
	LastTestedAt       *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ProviderProfileInput struct {
	ID                 ProviderProfileID
	TenantID           TenantID
	Capability         ProviderCapability
	ProviderKind       ProviderKind
	DisplayName        DisplayName
	EndpointURL        EndpointURL
	ModelName          ModelName
	RuntimeOptionsJSON []byte
	CapabilityJSON     []byte
	PromptTemplate     string
	CredentialStatus   CredentialStatus
	LifecycleState     ProviderProfileLifecycleState
	LastTestedAt       *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewProviderProfile(input ProviderProfileInput) (ProviderProfile, bool) {
	if _, ok := NewProviderProfileID(input.ID.String()); !ok {
		return ProviderProfile{}, false
	}
	if strings.TrimSpace(input.TenantID.String()) == "" {
		return ProviderProfile{}, false
	}
	capability, ok := NewProviderCapability(input.Capability.String())
	if !ok {
		return ProviderProfile{}, false
	}
	providerKind, ok := NewProviderKind(input.ProviderKind.String())
	if !ok {
		return ProviderProfile{}, false
	}
	displayName, ok := NewDisplayName(input.DisplayName.String())
	if !ok {
		return ProviderProfile{}, false
	}
	endpointURL, ok := NewEndpointURL(input.EndpointURL.String())
	if !ok {
		return ProviderProfile{}, false
	}
	modelName, ok := NewModelName(input.ModelName.String())
	if !ok {
		return ProviderProfile{}, false
	}
	runtimeOptions, ok := NewJSONObject(input.RuntimeOptionsJSON)
	if !ok {
		return ProviderProfile{}, false
	}
	capabilityJSON, ok := NewJSONObject(input.CapabilityJSON)
	if !ok {
		return ProviderProfile{}, false
	}
	promptTemplate, ok := NewPromptTemplate(input.PromptTemplate)
	if !ok {
		return ProviderProfile{}, false
	}
	if capability != ProviderCapabilityLanguageInference && promptTemplate.String() != "" {
		return ProviderProfile{}, false
	}
	credentialStatus, ok := NewCredentialStatus(input.CredentialStatus.String())
	if !ok {
		return ProviderProfile{}, false
	}
	lifecycleState, ok := NewProviderProfileLifecycleState(input.LifecycleState.String())
	if !ok {
		return ProviderProfile{}, false
	}
	if input.CreatedAt.IsZero() || input.UpdatedAt.IsZero() {
		return ProviderProfile{}, false
	}
	return ProviderProfile{
		ID:                 input.ID,
		TenantID:           input.TenantID,
		Capability:         capability,
		ProviderKind:       providerKind,
		DisplayName:        displayName,
		EndpointURL:        endpointURL,
		ModelName:          modelName,
		RuntimeOptionsJSON: runtimeOptions,
		CapabilityJSON:     capabilityJSON,
		PromptTemplate:     promptTemplate,
		CredentialStatus:   credentialStatus,
		LifecycleState:     lifecycleState,
		LastTestedAt:       input.LastTestedAt,
		CreatedAt:          input.CreatedAt,
		UpdatedAt:          input.UpdatedAt,
	}, true
}

func (p ProviderProfile) Enable(now time.Time) (ProviderProfile, bool) {
	if now.IsZero() || p.LifecycleState == ProviderProfileArchived {
		return ProviderProfile{}, false
	}
	p.LifecycleState = ProviderProfileEnabled
	p.UpdatedAt = now
	return p, true
}

func (p ProviderProfile) Disable(now time.Time) (ProviderProfile, bool) {
	if now.IsZero() || p.LifecycleState == ProviderProfileArchived {
		return ProviderProfile{}, false
	}
	p.LifecycleState = ProviderProfileDisabled
	p.UpdatedAt = now
	return p, true
}

func (p ProviderProfile) Archive(now time.Time) (ProviderProfile, bool) {
	if now.IsZero() || p.LifecycleState == ProviderProfileArchived {
		return ProviderProfile{}, false
	}
	p.LifecycleState = ProviderProfileArchived
	p.UpdatedAt = now
	return p, true
}

func (p ProviderProfile) WithCredentialConfigured(now time.Time) (ProviderProfile, bool) {
	if now.IsZero() || p.LifecycleState == ProviderProfileArchived {
		return ProviderProfile{}, false
	}
	p.CredentialStatus = CredentialStatusConfigured
	p.UpdatedAt = now
	return p, true
}

func (p ProviderProfile) WithLastTested(now time.Time) (ProviderProfile, bool) {
	if now.IsZero() || p.LifecycleState == ProviderProfileArchived {
		return ProviderProfile{}, false
	}
	p.LastTestedAt = &now
	p.UpdatedAt = now
	return p, true
}

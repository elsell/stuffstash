package ports

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

var ErrInvalidProviderCredential = errors.New("invalid provider credential")

const ProviderCredentialAlgorithmAES256GCM = "AES-256-GCM"
const ProviderCredentialAESGCMNonceBytes = 12

type ProviderCapability string

const (
	ProviderCapabilitySpeechToText      ProviderCapability = "speech_to_text"
	ProviderCapabilityLanguageInference ProviderCapability = "language_inference"
	ProviderCapabilityTextToSpeech      ProviderCapability = "text_to_speech"
)

type ProviderKind string

const (
	ProviderKindGemini           ProviderKind = "gemini"
	ProviderKindOpenAICompatible ProviderKind = "openai_compatible"
	ProviderKindLocalHTTP        ProviderKind = "local_http"
)

type ProviderCredentialPurpose string

const (
	ProviderCredentialPurposeAPIKey      ProviderCredentialPurpose = "api_key"
	ProviderCredentialPurposeOAuthBearer ProviderCredentialPurpose = "oauth_bearer"
)

type ProviderCredentialScope struct {
	TenantID          tenant.ID
	ProviderProfileID string
	Capability        ProviderCapability
	ProviderKind      ProviderKind
	Purpose           ProviderCredentialPurpose
}

type SealedProviderCredential struct {
	KeyID      string
	Algorithm  string
	Nonce      []byte
	Ciphertext []byte
}

type ProviderCredentialRecord struct {
	ID           string
	Scope        ProviderCredentialScope
	Sealed       SealedProviderCredential
	CreatedAt    time.Time
	UpdatedAt    time.Time
	SupersededAt *time.Time
}

type ProviderCredentialSealer interface {
	SealProviderCredential(ctx context.Context, scope ProviderCredentialScope, raw []byte) (SealedProviderCredential, error)
	UnsealProviderCredential(ctx context.Context, scope ProviderCredentialScope, sealed SealedProviderCredential) ([]byte, error)
}

type ProviderCredentialRepository interface {
	ReplaceProviderCredential(ctx context.Context, credential ProviderCredentialRecord) error
	ActiveProviderCredential(ctx context.Context, scope ProviderCredentialScope) (ProviderCredentialRecord, bool, error)
	ActiveProviderCredentialsExist(ctx context.Context) (bool, error)
	SupersedeActiveProviderCredential(ctx context.Context, scope ProviderCredentialScope, supersededAt time.Time) error
}

package ports

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

var ErrInvalidProviderInput = errors.New("invalid provider input")

type SpeechToTextProvider interface {
	Transcribe(ctx context.Context, input SpeechToTextInput) (SpeechToTextResult, error)
}

type SpeechToTextProviderProbe interface {
	ProbeSpeechToText(ctx context.Context) error
}

type SpeechToTextInput struct {
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Principal   identity.Principal
	AudioFormat RealtimeAudioFormat
	AudioChunks [][]byte
}

type SpeechToTextResult struct {
	Transcript string
}

type LanguageInferenceProvider interface {
	NextTurn(ctx context.Context, input LanguageInferenceInput) (LanguageInferenceTurn, error)
}

type VoiceResponseGenerator interface {
	GenerateResponse(ctx context.Context, input VoiceResponseGenerationInput) (VoiceResponseGenerationResult, error)
}

type RealtimeLanguageProvider interface {
	LanguageInferenceProvider
	VoiceResponseGenerator
}

type LanguageInferenceProviderProbe interface {
	ProbeLanguageInference(ctx context.Context) error
}

type LanguageInferenceInput struct {
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	Principal         identity.Principal
	Transcript        string
	ConversationTurns []AgentConversationTurn
	PromptTemplate    string
	PreviousTurns     int
	Investigation     *agentmodel.InvestigationInput
}

type AgentConversationRole string

const (
	AgentConversationRoleUser      AgentConversationRole = "user"
	AgentConversationRoleAssistant AgentConversationRole = "assistant"
)

type AgentConversationTurn struct {
	Role AgentConversationRole
	Kind string
	Text string
}

type LanguageInferenceTurn struct {
	Investigation *agentmodel.InvestigationStep
}

type VoiceResponseGenerationInput struct {
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Principal   identity.Principal
	Brief       agentmodel.GroundedVoiceResponseBrief
}

type VoiceResponseGenerationResult struct {
	SpokenResponse  string `json:"spokenResponse"`
	DisplayResponse string `json:"displayResponse"`
}

type AgentToolCall struct {
	ID        string
	Name      string
	Arguments map[string]any
}

type AgentToolResult struct {
	CallID  string
	Name    string
	Call    AgentToolCall
	Content string
}

type StructuredAgentResponseKind string

const (
	StructuredAgentResponseKindAnswer            StructuredAgentResponseKind = "answer"
	StructuredAgentResponseKindClarification     StructuredAgentResponseKind = "clarification"
	StructuredAgentResponseKindUnsupportedAction StructuredAgentResponseKind = "unsupported_action"
	StructuredAgentResponseKindSafeFailure       StructuredAgentResponseKind = "safe_failure"
)

type StructuredAgentResponse struct {
	ResponseID      string
	SessionID       string
	TenantID        tenant.ID
	InventoryID     inventory.InventoryID
	Source          string
	Kind            StructuredAgentResponseKind
	SpokenResponse  string
	DisplayResponse string
	Artifacts       []StructuredAgentResponseArtifact
	ToolCallIDs     []string
}

type StructuredAgentResponseArtifactType string

const StructuredAgentResponseArtifactAssetReference StructuredAgentResponseArtifactType = "asset_reference"

type StructuredAgentResponseArtifact struct {
	Type      StructuredAgentResponseArtifactType
	AssetID   asset.ID
	Title     string
	AssetKind asset.Kind
	Context   string
}

type TextToSpeechProvider interface {
	Synthesize(ctx context.Context, input TextToSpeechInput) (TextToSpeechResult, error)
}

type TextToSpeechProviderProbe interface {
	ProbeTextToSpeech(ctx context.Context) error
}

type TextToSpeechInput struct {
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Principal   identity.Principal
	Text        string
	MimeTypes   []string
}

type TextToSpeechResult struct {
	MimeType string
	Chunks   [][]byte
}

type RealtimeAudioFormat struct {
	MimeType   string
	SampleRate int
	Channels   int
}

type RealtimeVoiceProviderResolutionInput struct {
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Principal   identity.Principal
}

type RealtimeVoiceProviderSet struct {
	SpeechToTextProfileID      string
	LanguageInferenceProfileID string
	TextToSpeechProfileID      string
	LanguagePromptTemplate     string
	SpeechToText               SpeechToTextProvider
	LanguageInference          LanguageInferenceProvider
	ResponseGenerator          VoiceResponseGenerator
	TextToSpeech               TextToSpeechProvider
}

type RealtimeVoiceProviderResolver interface {
	ResolveRealtimeVoiceProviders(ctx context.Context, input RealtimeVoiceProviderResolutionInput) (RealtimeVoiceProviderSet, error)
}

type RealtimeSessionState string

const (
	RealtimeSessionStateStarted   RealtimeSessionState = "started"
	RealtimeSessionStateCompleted RealtimeSessionState = "completed"
	RealtimeSessionStateFailed    RealtimeSessionState = "failed"
	RealtimeSessionStateCancelled RealtimeSessionState = "cancelled"
)

type RealtimeSessionRecord struct {
	ID                         string
	TenantID                   tenant.ID
	InventoryID                inventory.InventoryID
	PrincipalID                identity.PrincipalID
	Source                     string
	State                      RealtimeSessionState
	SpeechToTextProfileID      string
	LanguageInferenceProfileID string
	TextToSpeechProfileID      string
	StartedAt                  time.Time
	LastActivityAt             time.Time
	EndedAt                    time.Time
	SafeFailureCode            string
}

type RealtimeSessionOutcome struct {
	State           RealtimeSessionState
	At              time.Time
	SafeFailureCode string
}

type RealtimeSessionRepository interface {
	SaveRealtimeSession(ctx context.Context, record RealtimeSessionRecord) error
	UpdateRealtimeSessionOutcome(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, sessionID string, outcome RealtimeSessionOutcome) error
	RealtimeSessionByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, sessionID string) (RealtimeSessionRecord, bool, error)
}

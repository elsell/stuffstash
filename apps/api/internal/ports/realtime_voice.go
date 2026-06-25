package ports

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

var ErrInvalidProviderInput = errors.New("invalid provider input")

type SpeechToTextProvider interface {
	Transcribe(ctx context.Context, input SpeechToTextInput) (SpeechToTextResult, error)
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

type LanguageInferenceInput struct {
	TenantID      tenant.ID
	InventoryID   inventory.InventoryID
	Principal     identity.Principal
	Transcript    string
	Tools         []AgentToolDescriptor
	ToolResults   []AgentToolResult
	PreviousTurns int
}

type AgentToolDescriptor struct {
	Name        string
	Label       string
	Description string
	ReadOnly    bool
}

type LanguageInferenceTurn struct {
	ToolCalls []AgentToolCall
	Final     *StructuredAgentResponse
}

type AgentToolCall struct {
	ID        string
	Name      string
	Arguments map[string]any
}

type AgentToolResult struct {
	CallID  string
	Name    string
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
	ToolCallIDs     []string
}

type TextToSpeechProvider interface {
	Synthesize(ctx context.Context, input TextToSpeechInput) (TextToSpeechResult, error)
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

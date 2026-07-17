package app

import (
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const (
	RealtimeVoiceSourceMobile = "mobile_voice"

	RealtimeVoiceEventTranscriptFinal             = "transcript.final"
	RealtimeVoiceEventAgentProgress               = "agent.progress"
	RealtimeVoiceEventAgentDiagnostic             = "agent.diagnostic"
	RealtimeVoiceEventToolCallStarted             = "tool.call.started"
	RealtimeVoiceEventToolCallCompleted           = "tool.call.completed"
	RealtimeVoiceEventToolCallFailed              = "tool.call.failed"
	RealtimeVoiceEventActionPlanProposed          = "action.plan.proposed"
	RealtimeVoiceEventActionPlanApproved          = "action.plan.approved"
	RealtimeVoiceEventActionPlanCancelled         = "action.plan.cancelled"
	RealtimeVoiceEventActionPlanExecuted          = "action.plan.executed"
	RealtimeVoiceEventActionPlanFailed            = "action.plan.failed"
	RealtimeVoiceEventAssistantResponseStarted    = "assistant.response.started"
	RealtimeVoiceEventAssistantResponseCompleted  = "assistant.response.completed"
	RealtimeVoiceEventTextToSpeechAudioStarted    = "tts.audio.started"
	RealtimeVoiceEventTextToSpeechAudioChunk      = "tts.audio.chunk"
	RealtimeVoiceEventTextToSpeechAudioCompleted  = "tts.audio.completed"
	RealtimeVoiceEventSessionCompleted            = "session.completed"
	RealtimeVoiceToolSearchAuthorizedAssets       = "search_authorized_assets"
	RealtimeVoiceToolGetAssetDetail               = "get_asset_detail"
	RealtimeVoiceToolListAuthorizedAssets         = "list_authorized_assets"
	RealtimeVoiceToolListAssetAuditHistory        = "list_asset_audit_history"
	RealtimeVoiceToolListCheckedOutAssets         = "list_checked_out_assets"
	RealtimeVoiceToolListAssetCheckoutHistory     = "list_asset_checkout_history"
	realtimeVoiceSearchAuthorizedAssetsPublicName = "Search inventory"
	realtimeVoiceGetAssetDetailPublicName         = "Inspect item"
	realtimeVoiceListAuthorizedAssetsPublicName   = "List inventory"
	realtimeVoiceListAssetAuditHistoryPublicName  = "Check history"
	realtimeVoiceListCheckedOutAssetsPublicName   = "List checked out"
	realtimeVoiceListCheckoutHistoryPublicName    = "Checkout history"
	realtimeVoiceFailureSpeechToText              = "speech_to_text_failed"
	realtimeVoiceFailureLanguageInference         = "language_inference_failed"
	realtimeVoiceFailureTextToSpeech              = "text_to_speech_failed"
	realtimeVoiceToolTurnBudget                   = 6
	realtimeVoiceProgressUnderstanding            = "understanding"
	realtimeVoiceProgressExploring                = "exploring"
	realtimeVoiceProgressPlanning                 = "planning"
	realtimeVoiceProgressReviewing                = "reviewing"
	realtimeVoiceProgressAnswering                = "answering"
	realtimeVoiceProgressRecovering               = "recovering"
)

type RealtimeVoiceSessionInput struct {
	Principal            identity.Principal
	TenantID             tenant.ID
	InventoryID          inventory.InventoryID
	Source               string
	InputAudio           ports.RealtimeAudioFormat
	OutputAudio          RealtimeVoiceOutputAudio
	DeveloperDiagnostics bool
}

type RealtimeVoiceOutputAudio struct {
	MimeTypes []string
}

type RealtimeVoiceSession struct {
	ID                         string
	TenantID                   tenant.ID
	InventoryID                inventory.InventoryID
	Principal                  identity.Principal
	Source                     string
	InputAudio                 ports.RealtimeAudioFormat
	OutputAudio                RealtimeVoiceOutputAudio
	SpeechToTextProfileID      string
	LanguageInferenceProfileID string
	TextToSpeechProfileID      string
	LanguagePromptTemplate     string
	DeveloperDiagnostics       bool
	speechToText               ports.SpeechToTextProvider
	languageInference          ports.LanguageInferenceProvider
	textToSpeech               ports.TextToSpeechProvider
}

type RealtimeVoiceQueryInput struct {
	Session                    RealtimeVoiceSession
	AudioChunks                [][]byte
	ContinueAfterClarification bool
	ConversationTurns          []ports.AgentConversationTurn
}

type RealtimeVoiceEvent struct {
	Type           string
	SessionID      string
	ToolCallID     string
	ToolLabel      string
	Status         string
	Code           string
	Message        string
	Text           string
	Detail         string
	Response       *ports.StructuredAgentResponse
	ActionPlan     *RealtimeVoiceActionPlanProposal
	PlanID         string
	CommandResults []RealtimeVoiceActionPlanCommandResult
	Audio          []byte
	AudioMime      string
	ChunkID        string
	FinalChunk     bool
}

type RealtimeVoiceEventSink func(RealtimeVoiceEvent) error

type RealtimeVoiceActionPlanProposal struct {
	PlanID              string
	ConfirmationSummary string
	Commands            []RealtimeVoiceActionPlanCommand
	Risks               []string
}

type RealtimeVoiceActionPlanCommand struct {
	ID              string
	Kind            string
	Summary         string
	Operation       string
	Title           string
	AssetKind       string
	ParentAssetID   string
	ParentTitle     string
	ParentKind      string
	ParentCommandID string
}

type RealtimeVoiceActionPlanCommandResult struct {
	CommandID string
	AssetID   string
	Operation string
	AssetKind string
}

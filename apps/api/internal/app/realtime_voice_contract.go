package app

import (
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const (
	RealtimeVoiceSourceMobile                     = agentmodelapp.RealtimeVoiceSourceMobile
	RealtimeVoiceSourceWebText                    = agentmodelapp.RealtimeVoiceSourceWebText
	RealtimeVoiceEventTranscriptFinal             = agentmodelapp.RealtimeVoiceEventTranscriptFinal
	RealtimeVoiceEventAgentProgress               = agentmodelapp.RealtimeVoiceEventAgentProgress
	RealtimeVoiceEventAgentDiagnostic             = agentmodelapp.RealtimeVoiceEventAgentDiagnostic
	RealtimeVoiceEventToolCallStarted             = agentmodelapp.RealtimeVoiceEventToolCallStarted
	RealtimeVoiceEventToolCallCompleted           = agentmodelapp.RealtimeVoiceEventToolCallCompleted
	RealtimeVoiceEventToolCallFailed              = agentmodelapp.RealtimeVoiceEventToolCallFailed
	RealtimeVoiceEventActionPlanProposed          = agentmodelapp.RealtimeVoiceEventActionPlanProposed
	RealtimeVoiceEventActionPlanApproved          = agentmodelapp.RealtimeVoiceEventActionPlanApproved
	RealtimeVoiceEventActionPlanCancelled         = agentmodelapp.RealtimeVoiceEventActionPlanCancelled
	RealtimeVoiceEventActionPlanExecuted          = agentmodelapp.RealtimeVoiceEventActionPlanExecuted
	RealtimeVoiceEventActionPlanFailed            = agentmodelapp.RealtimeVoiceEventActionPlanFailed
	RealtimeVoiceEventAssistantResponseStarted    = agentmodelapp.RealtimeVoiceEventAssistantResponseStarted
	RealtimeVoiceEventAssistantResponseCompleted  = agentmodelapp.RealtimeVoiceEventAssistantResponseCompleted
	RealtimeVoiceEventTextToSpeechAudioStarted    = agentmodelapp.RealtimeVoiceEventTextToSpeechAudioStarted
	RealtimeVoiceEventTextToSpeechAudioChunk      = agentmodelapp.RealtimeVoiceEventTextToSpeechAudioChunk
	RealtimeVoiceEventTextToSpeechAudioCompleted  = agentmodelapp.RealtimeVoiceEventTextToSpeechAudioCompleted
	RealtimeVoiceEventSessionCompleted            = agentmodelapp.RealtimeVoiceEventSessionCompleted
	RealtimeVoiceToolSearchAuthorizedAssets       = agentmodelapp.RealtimeVoiceToolSearchAuthorizedAssets
	RealtimeVoiceToolGetInventoryVocabulary       = agentmodelapp.RealtimeVoiceToolGetInventoryVocabulary
	RealtimeVoiceToolGetAssetDetail               = agentmodelapp.RealtimeVoiceToolGetAssetDetail
	RealtimeVoiceToolListAuthorizedAssets         = agentmodelapp.RealtimeVoiceToolListAuthorizedAssets
	RealtimeVoiceToolListAssetAuditHistory        = agentmodelapp.RealtimeVoiceToolListAssetAuditHistory
	RealtimeVoiceToolListCheckedOutAssets         = agentmodelapp.RealtimeVoiceToolListCheckedOutAssets
	RealtimeVoiceToolListAssetCheckoutHistory     = agentmodelapp.RealtimeVoiceToolListAssetCheckoutHistory
	realtimeVoiceSearchAuthorizedAssetsPublicName = agentmodelapp.RealtimeVoiceSearchAuthorizedAssetsPublicName
	realtimeVoiceGetAssetDetailPublicName         = agentmodelapp.RealtimeVoiceGetAssetDetailPublicName
	realtimeVoiceListAuthorizedAssetsPublicName   = agentmodelapp.RealtimeVoiceListAuthorizedAssetsPublicName
	realtimeVoiceListAssetAuditHistoryPublicName  = agentmodelapp.RealtimeVoiceListAssetAuditHistoryPublicName
	realtimeVoiceListCheckedOutAssetsPublicName   = agentmodelapp.RealtimeVoiceListCheckedOutAssetsPublicName
	realtimeVoiceListCheckoutHistoryPublicName    = agentmodelapp.RealtimeVoiceListCheckoutHistoryPublicName
	realtimeVoiceFailureSpeechToText              = agentmodelapp.RealtimeVoiceFailureSpeechToText
	realtimeVoiceFailureLanguageInference         = agentmodelapp.RealtimeVoiceFailureLanguageInference
	realtimeVoiceFailureTextToSpeech              = agentmodelapp.RealtimeVoiceFailureTextToSpeech
	realtimeVoiceToolTurnBudget                   = agentmodelapp.RealtimeVoiceToolTurnBudget
	realtimeVoiceProgressUnderstanding            = agentmodelapp.RealtimeVoiceProgressUnderstanding
	realtimeVoiceProgressExploring                = agentmodelapp.RealtimeVoiceProgressExploring
	realtimeVoiceProgressPlanning                 = agentmodelapp.RealtimeVoiceProgressPlanning
	realtimeVoiceProgressReviewing                = agentmodelapp.RealtimeVoiceProgressReviewing
	realtimeVoiceProgressAnswering                = agentmodelapp.RealtimeVoiceProgressAnswering
	realtimeVoiceProgressRecovering               = agentmodelapp.RealtimeVoiceProgressRecovering
)

type RealtimeVoiceSessionInput struct {
	ConversationContinuity bool
	Principal              identity.Principal
	TenantID               tenant.ID
	InventoryID            inventory.InventoryID
	Source                 string
	InputAudio             ports.RealtimeAudioFormat
	OutputAudio            RealtimeVoiceOutputAudio
	DeveloperDiagnostics   bool
}

type RealtimeVoiceOutputAudio struct {
	MimeTypes []string
}

type RealtimeVoiceSession struct {
	silentReply                bool
	conversationMemory         *agentmodelapp.ConversationMemory
	conversationModel          ports.ConversationModel
	ConversationContinuity     bool
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
	WorkflowRevisionID         string
	LanguagePromptTemplate     string
	DeveloperDiagnostics       bool
	workflow                   *agentmodelapp.PreparedWorkflow
	speechToText               ports.SpeechToTextProvider
	textToSpeech               ports.TextToSpeechProvider
}

const MaxRealtimeTextCharacters = agentmodelapp.MaxRealtimeTextCharacters

type RealtimeVoiceQueryInput struct {
	Text                       string
	Session                    RealtimeVoiceSession
	AudioChunks                [][]byte
	ContinueAfterClarification bool
	ConversationTurns          []ports.AgentConversationTurn
}

type RealtimeVoiceEvent = agentmodelapp.RealtimeVoiceEvent
type RealtimeVoiceEventSink = agentmodelapp.RealtimeVoiceEventSink
type RealtimeVoiceActionPlanProposal = agentmodelapp.RealtimeVoiceActionPlanProposal
type RealtimeVoiceActionPlanExpiration = agentmodelapp.RealtimeVoiceActionPlanExpiration
type RealtimeVoiceActionPlanCommand = agentmodelapp.RealtimeVoiceActionPlanCommand
type RealtimeVoiceActionPlanCommandResult = agentmodelapp.RealtimeVoiceActionPlanCommandResult

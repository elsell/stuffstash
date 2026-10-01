package agentmodel

import "github.com/stuffstash/stuff-stash/internal/ports"

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

type RealtimeVoiceActionPlanExpiration struct {
	Date      string
	Precision string
}

type RealtimeVoiceActionPlanCommand struct {
	Changes           []string
	ExpirationCleared bool
	Expiration        *RealtimeVoiceActionPlanExpiration
	ID                string
	Kind              string
	Summary           string
	Operation         string
	Title             string
	AssetKind         string
	ParentAssetID     string
	ParentTitle       string
	ParentKind        string
	ParentCommandID   string
}

type RealtimeVoiceActionPlanCommandResult struct {
	CommandID string
	AssetID   string
	Operation string
	AssetKind string
}

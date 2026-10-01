package agentmodel

import (
	"context"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Session inputs are wired from the prepared realtime session, never from model output.
type RealtimeConversationSession struct {
	RealtimeReadToolScope
	ID                     string
	Source                 string
	LanguagePromptTemplate string
	Memory                 *ConversationMemory
	Model                  ports.ConversationModel
	Workflow               *PreparedWorkflow
	ContextBytes           int
}

type RealtimeConversationQueries interface {
	GetAsset(context.Context, assetapp.GetAssetInput) (asset.Asset, error)
	EnsureActiveInventoryAccess(context.Context, identity.Principal, tenant.ID, inventory.InventoryID, ports.InventoryPermission) error
	EnsureRealtimeVoiceAccess(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error
}

type RealtimeConversationDependencies struct {
	Queries          RealtimeConversationQueries
	Reads            RealtimeReadTools
	Plans            ActionPlanService
	CustomAssetTypes ports.CustomAssetTypeRepository
	CustomFields     ports.CustomFieldDefinitionRepository
}

type RealtimeConversationService struct {
	RealtimeConversationQueries
	reads            RealtimeReadTools
	plans            ActionPlanService
	customAssetTypes ports.CustomAssetTypeRepository
	customFields     ports.CustomFieldDefinitionRepository
}

type RealtimeConversationCompletion func(context.Context, ports.StructuredAgentResponse, []string, []ports.AgentToolResult) error
type RealtimeConversationFailure func(int, int, []ports.AgentToolResult, error) error

func NewRealtimeConversationService(deps RealtimeConversationDependencies) RealtimeConversationService {
	return RealtimeConversationService{RealtimeConversationQueries: deps.Queries, reads: deps.Reads, plans: deps.Plans, customAssetTypes: deps.CustomAssetTypes, customFields: deps.CustomFields}
}

func (a RealtimeConversationService) NewTools(session RealtimeConversationSession, emit RealtimeVoiceEventSink) *RealtimeConversationTools {
	return &RealtimeConversationTools{application: a, session: session, emit: emit, visible: map[string]struct{}{}, items: map[string]RealtimeVoiceAssetToolItem{}, stale: map[string]bool{}}
}

func (e *RealtimeConversationTools) Proposal() *RealtimeVoiceActionPlanProposal { return e.proposal }

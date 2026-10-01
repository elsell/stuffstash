package agentmodel

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const RealtimeVoiceSourceMobile = "mobile_voice"
const RealtimeVoiceSourceWebText = "web_text"

type RealtimeSessionDependencies struct {
	Workflows    ConversationWorkflowService
	Providers    ports.RealtimeVoiceProviderResolver
	Access       ports.RealtimeInventoryAccess
	Authorizer   ports.Authorizer
	Sessions     ports.RealtimeSessionRepository
	Clock        ports.Clock
	IDs          ports.IDGenerator
	ContextBytes int
}

type RealtimeSessionService struct{ deps RealtimeSessionDependencies }

func NewRealtimeSessionService(deps RealtimeSessionDependencies) RealtimeSessionService {
	return RealtimeSessionService{deps: deps}
}

type RealtimeSessionStartInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Source      string
	InputAudio  ports.RealtimeAudioFormat
}

// PreparedRealtimeSession contains one already-selected runtime and scoped memory.
// The facade must carry these identities forward, not resolve them a second time.
type PreparedRealtimeSession struct {
	ID                 string
	Providers          ports.RealtimeVoiceProviderSet
	Workflow           *PreparedWorkflow
	WorkflowRevisionID string
	Memory             *ConversationMemory
}

func (s RealtimeSessionService) Start(ctx context.Context, input RealtimeSessionStartInput) (PreparedRealtimeSession, error) {
	if s.deps.Authorizer == nil || s.deps.Access == nil || s.deps.Providers == nil || s.deps.Sessions == nil || s.deps.Clock == nil {
		return PreparedRealtimeSession{}, apperrors.ErrInvalidInput
	}
	if input.Source != RealtimeVoiceSourceMobile && input.Source != RealtimeVoiceSourceWebText {
		return PreparedRealtimeSession{}, apperrors.ErrInvalidInput
	}
	if input.InputAudio.MimeType != "audio/mp4" || input.InputAudio.Channels != 1 {
		return PreparedRealtimeSession{}, apperrors.ErrInvalidInput
	}
	if err := s.EnsureAccess(ctx, input.Principal, input.TenantID, input.InventoryID); err != nil {
		return PreparedRealtimeSession{}, err
	}
	selectedWorkflow, err := s.deps.Workflows.Selected(ctx, input.Principal, input.TenantID)
	if err != nil {
		return PreparedRealtimeSession{}, err
	}
	providers, err := s.deps.Providers.ResolveRealtimeVoiceProviders(ctx, ports.RealtimeVoiceProviderResolutionInput{
		SkipDefaultLanguage: selectedWorkflow != nil && !selectedWorkflow.NeedsDefaultLanguage(),
		TenantID:            input.TenantID,
		InventoryID:         input.InventoryID,
		Principal:           input.Principal,
	})
	if err != nil {
		return PreparedRealtimeSession{}, err
	}

	var workflow *PreparedWorkflow
	if selectedWorkflow != nil {
		workflowResolver, _ := s.deps.Providers.(ports.WorkflowLanguageProviderResolver)
		workflow, err = selectedWorkflow.Prepare(ctx, providers, workflowResolver)
		if err != nil {
			return PreparedRealtimeSession{}, err
		}
		providers.ConversationModel = workflow.ConversationModel()
		providers.LanguageInferenceProfileID = workflow.ConversationProfileID()
		providers.LanguagePromptTemplate = ""
	}
	if providers.SpeechToText == nil || providers.TextToSpeech == nil || providers.ConversationModel == nil {
		return PreparedRealtimeSession{}, apperrors.ErrInvalidInput
	}

	sessionID := s.newID()
	if strings.TrimSpace(sessionID) == "" {
		return PreparedRealtimeSession{}, apperrors.ErrInvalidInput
	}

	prepared := PreparedRealtimeSession{ID: sessionID, Providers: providers, Workflow: workflow}
	if workflow != nil {
		prepared.WorkflowRevisionID = string(workflow.Revision().Snapshot().ID)
	}
	prepared.Memory = NewConversationMemory(ConversationScope{SessionID: sessionID, PrincipalID: input.Principal.ID, TenantID: input.TenantID, InventoryID: input.InventoryID}, s.deps.ContextBytes)
	now := s.deps.Clock.Now()
	if err := s.deps.Sessions.SaveRealtimeSession(ctx, ports.RealtimeSessionRecord{
		ID: sessionID, TenantID: input.TenantID, InventoryID: input.InventoryID,
		PrincipalID: input.Principal.ID, Source: input.Source, State: ports.RealtimeSessionStateStarted,
		SpeechToTextProfileID:      providers.SpeechToTextProfileID,
		LanguageInferenceProfileID: providers.LanguageInferenceProfileID,
		TextToSpeechProfileID:      providers.TextToSpeechProfileID,
		StartedAt:                  now, LastActivityAt: now,
	}); err != nil {
		return PreparedRealtimeSession{}, err
	}
	return prepared, nil
}

func (s RealtimeSessionService) EnsureAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	if err := s.deps.Authorizer.CheckTenant(ctx, principal, ports.TenantPermissionView, tenantID); err != nil {
		s.deps.Access.RecordAuthorizationDenied(ctx, principal, tenantID)
		return err
	}
	return s.deps.Access.EnsureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, ports.InventoryPermissionView)
}

func (s RealtimeSessionService) UpdateOutcome(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, sessionID string, state ports.RealtimeSessionState, safeFailureCode string) error {
	if s.deps.Sessions == nil || strings.TrimSpace(sessionID) == "" {
		return apperrors.ErrInvalidInput
	}
	return s.deps.Sessions.UpdateRealtimeSessionOutcome(ctx, tenantID, inventoryID, sessionID, ports.RealtimeSessionOutcome{State: state, At: s.deps.Clock.Now(), SafeFailureCode: strings.TrimSpace(safeFailureCode)})
}

func (s RealtimeSessionService) newID() string {
	if s.deps.IDs == nil {
		return ""
	}
	return s.deps.IDs.NewID()
}

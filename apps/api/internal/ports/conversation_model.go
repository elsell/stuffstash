package ports

import (
	"context"
	"encoding/json"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type ConversationRole string

const (
	ConversationRoleUser      ConversationRole = "user"
	ConversationRoleAssistant ConversationRole = "assistant"
	ConversationRoleTool      ConversationRole = "tool"
)

type ConversationMessage struct {
	ProviderState []byte `json:"-"`
	Role          ConversationRole
	Text          string
	ToolCalls     []AgentToolCall
	ToolResults   []AgentToolResult
}

type ConversationToolDefinition struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

type ConversationModelInput struct {
	Principal    identity.Principal
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	Instructions string
	Messages     []ConversationMessage
	Tools        []ConversationToolDefinition
}

type ConversationAnswer struct {
	Spoken   string
	Display  string
	AssetIDs []string
}

type ConversationModelTurn struct {
	ProviderState []byte `json:"-"`
	Text          string
	ToolCalls     []AgentToolCall
	Answer        *ConversationAnswer
}

type ConversationModel interface {
	Converse(context.Context, ConversationModelInput) (ConversationModelTurn, error)
}

// The executor is bound to an authorized session by the application. A pause is
// produced only by a validated proposal; the model cannot authorize execution.
type ConversationToolOutcome struct {
	Answer         *ConversationAnswer
	Result         AgentToolResult
	ApprovalPlanID string
}

type ConversationToolExecutor interface {
	ExecuteConversationTool(context.Context, AgentToolCall) (ConversationToolOutcome, error)
}

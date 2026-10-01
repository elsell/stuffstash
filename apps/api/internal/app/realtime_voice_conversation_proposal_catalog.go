package app

import (
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeConversationProposalTool() ports.ConversationToolDefinition {
	return agentmodelapp.RealtimeConversationProposalTool()
}

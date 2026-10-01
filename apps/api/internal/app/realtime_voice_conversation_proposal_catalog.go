package app

import (
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeConversationProposalTool() ports.ConversationToolDefinition {
	return tools.ConversationProposal()
}

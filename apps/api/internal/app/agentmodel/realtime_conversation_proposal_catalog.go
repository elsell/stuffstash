package agentmodel

import (
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func RealtimeConversationProposalTool() ports.ConversationToolDefinition {
	return tools.ConversationProposal()
}

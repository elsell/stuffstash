package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type scriptedRealtimeLanguageInference struct {
	turns                 []ports.LanguageInferenceTurn
	errs                  []error
	seenTools             [][]ports.AgentToolDescriptor
	seenToolResults       [][]ports.AgentToolResult
	seenFinalOnly         []bool
	seenPlanOnly          []bool
	seenRequireToolCall   []bool
	seenTranscripts       []string
	seenConversationTurns [][]ports.AgentConversationTurn
	seenInvestigations    []*agentmodel.InvestigationInput
}

func (s *scriptedRealtimeLanguageInference) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	s.seenTools = append(s.seenTools, append([]ports.AgentToolDescriptor{}, input.Tools...))
	s.seenToolResults = append(s.seenToolResults, append([]ports.AgentToolResult{}, input.ToolResults...))
	s.seenFinalOnly = append(s.seenFinalOnly, input.FinalOnly)
	s.seenPlanOnly = append(s.seenPlanOnly, input.PlanOnly)
	s.seenRequireToolCall = append(s.seenRequireToolCall, input.RequireToolCall)
	s.seenTranscripts = append(s.seenTranscripts, input.Transcript)
	s.seenConversationTurns = append(s.seenConversationTurns, append([]ports.AgentConversationTurn{}, input.ConversationTurns...))
	s.seenInvestigations = append(s.seenInvestigations, input.Investigation)
	if len(s.errs) > 0 {
		err := s.errs[0]
		s.errs = s.errs[1:]
		if err != nil {
			return ports.LanguageInferenceTurn{}, err
		}
	}
	if len(s.turns) == 0 {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	turn := s.turns[0]
	s.turns = s.turns[1:]
	return turn, nil
}

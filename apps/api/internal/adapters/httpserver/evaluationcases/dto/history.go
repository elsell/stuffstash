package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type HistoryInput struct {
	GetInput
	Limit  int    `query:"limit" minimum:"1" maximum:"100"`
	Cursor string `query:"cursor"`
}
type HistoryOutput struct {
	Body shared.SuccessEnvelope[[]EvaluationCaseRevision]
}

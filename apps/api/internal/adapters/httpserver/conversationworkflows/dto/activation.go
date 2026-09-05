package dto

import rundto "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationruns/dto"

type WorkflowSelection struct {
	WorkflowID string `json:"workflowId" minLength:"1"`
	RevisionID string `json:"revisionId" minLength:"1"`
}
type WorkflowActivationBody struct {
	RevisionID string                              `json:"revisionId" minLength:"1"`
	RunID      string                              `json:"runId" minLength:"1"`
	Cases      []rundto.EvaluationRunCaseReference `json:"cases" minItems:"1" maxItems:"100"`
	Expected   *WorkflowSelection                  `json:"expected,omitempty"`
}
type ActivateInput struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
	TenantID      string `path:"tenantId"`
	WorkflowID    string `path:"workflowId"`
	Body          WorkflowActivationBody
}

package agentmodel

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidEvaluationCaseRevision = errors.New("invalid evaluation case revision")

type EvaluationCaseID string
type EvaluationCaseRevisionID string
type EvaluationCaseAuthorID string

type EvaluationCaseRevisionInput struct {
	ID         EvaluationCaseRevisionID
	CaseID     EvaluationCaseID
	TenantID   TenantID
	AuthorID   EvaluationCaseAuthorID
	Number     int
	Definition EvaluationCaseDefinition
	CreatedAt  time.Time
}
type EvaluationCaseRevision struct{ snapshot EvaluationCaseRevisionInput }

func NewEvaluationCaseRevision(input EvaluationCaseRevisionInput) (EvaluationCaseRevision, error) {
	if !workflowIdentifierValid(string(input.ID)) || !workflowIdentifierValid(string(input.CaseID)) || strings.TrimSpace(string(input.TenantID)) == "" || !workflowAuthorValid(string(input.AuthorID)) || input.Number < 1 || input.CreatedAt.IsZero() {
		return EvaluationCaseRevision{}, ErrInvalidEvaluationCaseRevision
	}
	definition, err := NewEvaluationCaseDefinition(input.Definition.Settings())
	if err != nil {
		return EvaluationCaseRevision{}, ErrInvalidEvaluationCaseRevision
	}
	input.Definition = definition
	return EvaluationCaseRevision{snapshot: input}, nil
}
func (revision EvaluationCaseRevision) Snapshot() EvaluationCaseRevisionInput {
	return revision.snapshot
}

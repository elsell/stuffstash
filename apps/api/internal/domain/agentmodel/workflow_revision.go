package agentmodel

import (
	"errors"
	"strings"
	"time"
	"unicode"
)

var ErrInvalidWorkflowRevision = errors.New("invalid conversation workflow revision")

type WorkflowID string

type WorkflowRevisionID string

type WorkflowAuthorID string

type WorkflowRevisionInput struct {
	ID         WorkflowRevisionID
	WorkflowID WorkflowID
	TenantID   TenantID
	AuthorID   WorkflowAuthorID
	Number     int
	Definition WorkflowDefinition
	Limits     WorkflowLimits
	CreatedAt  time.Time
}

// A revision is immutable. Activation belongs to the workflow repository and
// never changes this snapshot or the revision held by an in-flight turn.
type WorkflowRevision struct {
	snapshot WorkflowRevisionInput
}

func NewWorkflowRevision(input WorkflowRevisionInput) (WorkflowRevision, error) {
	for _, id := range []string{string(input.ID), string(input.WorkflowID)} {
		if !workflowIdentifierValid(id) {
			return WorkflowRevision{}, ErrInvalidWorkflowRevision
		}
	}
	if !workflowAuthorValid(string(input.AuthorID)) || strings.TrimSpace(string(input.TenantID)) == "" || input.Number < 1 || input.CreatedAt.IsZero() {
		return WorkflowRevision{}, ErrInvalidWorkflowRevision
	}
	definition, err := NewWorkflowDefinition(input.Definition.Settings(), input.Limits)
	if err != nil {
		return WorkflowRevision{}, ErrInvalidWorkflowRevision
	}
	input.Definition = definition
	return WorkflowRevision{snapshot: input}, nil
}

func (revision WorkflowRevision) Snapshot() WorkflowRevisionInput {
	return revision.snapshot
}

func workflowIdentifierValid(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for _, c := range value {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' {
			continue
		}
		return false
	}
	return true
}

func workflowAuthorValid(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

package app

import (
	"bytes"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
)

type workflowDefinitionInput struct {
	Name              string  `json:"name"`
	Instructions      *string `json:"instructions,omitempty"`
	ProviderProfileID *string `json:"providerProfileId,omitempty"`
	Budget            *struct {
		ElapsedSeconds *int64 `json:"elapsedSeconds"`
		FollowUpTurns  *int64 `json:"followUpTurns"`
		ModelCalls     *int64 `json:"modelCalls"`
		ToolCalls      *int64 `json:"toolCalls"`
	} `json:"budget"`
}
type workflowCreationInput struct {
	Schema     *string                  `json:"$schema,omitempty"`
	Definition *workflowDefinitionInput `json:"definition"`
}
type workflowRevisionInput struct {
	workflowCreationInput
	ExpectedRevision int64 `json:"expectedRevision"`
}
type workflowActivationInput struct {
	Schema     *string                  `json:"$schema,omitempty"`
	RevisionID string                   `json:"revisionId"`
	RunID      string                   `json:"runId"`
	Cases      json.RawMessage          `json:"cases"`
	Expected   *ports.WorkflowSelection `json:"expected,omitempty"`
}

func isWorkflowWrite(o Options) bool {
	c := o.Command
	return len(c) == 2 && c[0] == "workflows" && c[1] == "create" || len(c) == 3 && c[0] == "workflows" && c[1] == "activate" && c[2] != "" || len(c) == 4 && c[0] == "workflows" && c[1] == "revisions" && c[2] == "create" && c[3] != ""
}
func decodeWorkflowInput(body []byte, target any) error {
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if !json.Valid(body) || d.Decode(target) != nil {
		return ports.Failure("usage", "The workflow input has unknown fields or incorrect field types. Use the request shape shown by this command's --help.")
	}
	return nil
}
func prepareWorkflowInput(o Options) (Options, error) {
	if o.InputPath == "" {
		return o, ports.Failure("usage", "Supply structured workflow JSON with --input FILE or --input - for stdin. Use this command's --help for the request fields.")
	}
	if o.Command[1] == "activate" {
		var v workflowActivationInput
		if err := decodeWorkflowInput(o.RequestBody, &v); err != nil {
			return o, err
		}
		if strings.TrimSpace(v.RevisionID) == "" || strings.TrimSpace(v.RunID) == "" || len(v.Cases) == 0 {
			return o, ports.Failure("usage", "Activation input requires revisionId, runId and cases, with optional expected selection.")
		}
		var cases []struct {
			CaseID     string `json:"caseId"`
			RevisionID string `json:"revisionId"`
		}
		if err := decodeWorkflowInput(v.Cases, &cases); err != nil {
			return o, err
		}
		for _, c := range cases {
			if strings.TrimSpace(c.CaseID) == "" || strings.TrimSpace(c.RevisionID) == "" {
				return o, ports.Failure("usage", "Each activation case requires caseId and revisionId.")
			}
		}
		if v.Expected != nil && (strings.TrimSpace(v.Expected.WorkflowId) == "" || strings.TrimSpace(v.Expected.RevisionId) == "") {
			return o, ports.Failure("usage", "Expected selection requires workflowId and revisionId, or null when no selection is expected.")
		}
		return o, nil
	}
	var definition *workflowDefinitionInput
	if o.Command[1] == "create" {
		var v workflowCreationInput
		if err := decodeWorkflowInput(o.RequestBody, &v); err != nil {
			return o, err
		}
		definition = v.Definition
	} else {
		var v workflowRevisionInput
		if err := decodeWorkflowInput(o.RequestBody, &v); err != nil {
			return o, err
		}
		definition = v.Definition
		if v.ExpectedRevision <= 0 {
			return o, ports.Failure("usage", "Revision input requires a positive integer expectedRevision. Inspect the workflow before retrying a conflict.")
		}
	}
	if definition == nil || strings.TrimSpace(definition.Name) == "" || definition.Budget == nil {
		return o, ports.Failure("usage", "Input requires definition with name and budget. Supply all four budget fields shown by --help.")
	}
	b := definition.Budget
	if b.ElapsedSeconds == nil || b.FollowUpTurns == nil || b.ModelCalls == nil || b.ToolCalls == nil {
		return o, ports.Failure("usage", "Budget requires integer elapsedSeconds, followUpTurns, modelCalls and toolCalls.")
	}
	return o, nil
}

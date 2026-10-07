package app

import (
	"bytes"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
)

type evaluationDefinitionInput struct {
	Title     string `json:"title"`
	Utterance string `json:"utterance"`
	Assets    []struct {
		ID          string   `json:"id"`
		Title       string   `json:"title"`
		Kind        string   `json:"kind"`
		Description *string  `json:"description,omitempty"`
		ParentID    *string  `json:"parentId,omitempty"`
		TagNames    []string `json:"tagNames,omitempty"`
	} `json:"assets,omitempty"`
	Expectations *struct {
		Kind             string                         `json:"kind"`
		ReferencedAssets []string                       `json:"referencedAssets,omitempty"`
		Locations        []ports.EvaluationCaseLocation `json:"locations,omitempty"`
		Proposals        []struct {
			Operation     string  `json:"operation"`
			TargetID      *string `json:"targetId,omitempty"`
			DestinationID *string `json:"destinationId,omitempty"`
			NewTitle      *string `json:"newTitle,omitempty"`
			NewKind       *string `json:"newKind,omitempty"`
			Details       *string `json:"details,omitempty"`
		} `json:"proposals,omitempty"`
		ForbiddenOperations []string `json:"forbiddenOperations,omitempty"`
	} `json:"expectations"`
}
type evaluationCreationInput struct {
	Schema     *string                    `json:"$schema,omitempty"`
	Definition *evaluationDefinitionInput `json:"definition"`
}
type evaluationRevisionInput struct {
	evaluationCreationInput
	ExpectedRevision int64 `json:"expectedRevision"`
}
type evaluationQueueInput struct {
	Schema     *string `json:"$schema,omitempty"`
	WorkflowID string  `json:"workflowId"`
	RevisionID string  `json:"revisionId"`
	Cases      []struct {
		CaseID     string `json:"caseId"`
		RevisionID string `json:"revisionId"`
	} `json:"cases"`
}

func isEvaluationWrite(o Options) bool {
	c := o.Command
	return len(c) == 3 && c[0] == "evaluation" && (c[1] == "cases" || c[1] == "runs") && c[2] == "create" || len(c) == 4 && c[0] == "evaluation" && c[1] == "revisions" && c[2] == "create" && c[3] != ""
}
func decodeEvaluationWrite(body []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if !json.Valid(body) || decoder.Decode(target) != nil {
		return ports.Failure("usage", "Evaluation input has unknown fields or incorrect field types. Use the request shape shown by this command's --help.")
	}
	return nil
}
func prepareEvaluationWrite(o Options) (Options, error) {
	if o.InputPath == "" {
		return o, ports.Failure("usage", "Supply evaluation JSON with --input FILE or --input - for stdin. Use this command's --help for request fields.")
	}
	if o.Command[1] == "runs" {
		var v evaluationQueueInput
		if err := decodeEvaluationWrite(o.RequestBody, &v); err != nil {
			return o, err
		}
		if strings.TrimSpace(v.WorkflowID) == "" || strings.TrimSpace(v.RevisionID) == "" || len(v.Cases) < 1 || len(v.Cases) > 100 {
			return o, ports.Failure("usage", "Supply workflowId, revisionId, and 1 to 100 case references in the run input.")
		}
		for _, c := range v.Cases {
			if strings.TrimSpace(c.CaseID) == "" || strings.TrimSpace(c.RevisionID) == "" {
				return o, ports.Failure("usage", "Supply caseId and revisionId for each run case.")
			}
		}
		return o, nil
	}
	var definition *evaluationDefinitionInput
	if o.Command[1] == "cases" {
		var v evaluationCreationInput
		if err := decodeEvaluationWrite(o.RequestBody, &v); err != nil {
			return o, err
		}
		definition = v.Definition
	} else {
		var v evaluationRevisionInput
		if err := decodeEvaluationWrite(o.RequestBody, &v); err != nil {
			return o, err
		}
		definition = v.Definition
		if v.ExpectedRevision <= 0 {
			return o, ports.Failure("usage", "Supply an integer greater than zero for expectedRevision. Examine the case before you try again.")
		}
	}
	if definition == nil || strings.TrimSpace(definition.Title) == "" || strings.TrimSpace(definition.Utterance) == "" || definition.Expectations == nil || strings.TrimSpace(definition.Expectations.Kind) == "" {
		return o, ports.Failure("usage", "Supply title, utterance, and expectations with kind in the definition.")
	}
	for _, a := range definition.Assets {
		if strings.TrimSpace(a.ID) == "" || strings.TrimSpace(a.Title) == "" || strings.TrimSpace(a.Kind) == "" {
			return o, ports.Failure("usage", "Supply id, title, and kind for each fixture asset.")
		}
	}
	for _, l := range definition.Expectations.Locations {
		if strings.TrimSpace(l.AssetID) == "" || strings.TrimSpace(l.AncestorID) == "" {
			return o, ports.Failure("usage", "Supply assetId and ancestorId for each location expectation.")
		}
	}
	for _, p := range definition.Expectations.Proposals {
		if strings.TrimSpace(p.Operation) == "" {
			return o, ports.Failure("usage", "Supply operation for each proposal expectation.")
		}
	}
	return o, nil
}

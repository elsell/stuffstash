package presentation

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func (o Output) evaluationRevision(v ports.EvaluationCaseRevision) error {
	definition, err := json.Marshal(v.Definition)
	if err != nil {
		return err
	}
	return o.details([][2]string{{"ID", v.ID}, {"Case", v.CaseID}, {"Revision", strconv.FormatInt(v.Number, 10)}, {"Author", v.AuthorID}, {"Created", v.CreatedAt}, {"Title", v.Definition.Title}, {"Utterance", v.Definition.Utterance}, {"Definition", string(definition)}})
}
func (o Output) evaluationRun(v ports.EvaluationRun) error {
	fields := [][2]string{{"ID", v.ID}, {"Workflow", v.WorkflowID}, {"Revision", v.RevisionID}, {"State", v.State}, {"Coverage", v.Coverage}, {"Author", v.AuthorID}, {"Version", strconv.FormatInt(v.Version, 10)}, {"Total cases", strconv.FormatInt(v.TotalCases, 10)}, {"Completed cases", strconv.FormatInt(v.CompletedCases, 10)}, {"Passed cases", strconv.FormatInt(v.PassedCases, 10)}, {"Created", v.CreatedAt}, {"Updated", v.UpdatedAt}}
	for _, entry := range []struct {
		label string
		value any
	}{{"Started", v.StartedAt}, {"Finished", v.FinishedAt}, {"Cases", v.Cases}, {"Providers", v.Providers}, {"Results", v.Results}} {
		data, err := json.Marshal(entry.value)
		if err != nil {
			return err
		}
		fields = append(fields, [2]string{entry.label, string(data)})
	}
	if len(v.FailureCode) > 0 {
		data, err := json.Marshal(v.FailureCode)
		if err != nil {
			return err
		}
		fields = append(fields, [2]string{"Failure code", string(data)})
	}
	return o.details(fields)
}
func (o Output) evaluationCases(v ports.Result[[]ports.EvaluationCaseHead]) error {
	for _, item := range v.Data {
		if err := o.details([][2]string{{"ID", item.ID}, {"Title", item.Title}, {"Latest revision", strconv.FormatInt(item.LatestRevision, 10)}, {"Latest revision ID", item.LatestRevisionID}, {"Created", item.CreatedAt}, {"Updated", item.UpdatedAt}}); err != nil {
			return err
		}
	}
	return o.pagination(v.Pagination)
}
func (o Output) evaluationRuns(v ports.Result[[]ports.EvaluationRunHead]) error {
	for _, item := range v.Data {
		if err := o.details([][2]string{{"ID", item.ID}, {"Workflow", item.WorkflowID}, {"Revision", item.RevisionID}, {"State", item.State}, {"Version", strconv.FormatInt(item.Version, 10)}, {"Total cases", strconv.FormatInt(item.TotalCases, 10)}, {"Completed cases", strconv.FormatInt(item.CompletedCases, 10)}, {"Passed cases", strconv.FormatInt(item.PassedCases, 10)}, {"Created", item.CreatedAt}, {"Updated", item.UpdatedAt}}); err != nil {
			return err
		}
	}
	return o.pagination(v.Pagination)
}

package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isEvaluationCommand(o Options) bool { return len(o.Command) > 0 && o.Command[0] == "evaluation" }
func validateEvaluation(o Options, scope bool) error {
	valid := false
	if len(o.Command) >= 3 {
		switch o.Command[1] {
		case "cases", "runs":
			valid = o.Command[2] == "list" && len(o.Command) == 3 || o.Command[2] == "show" && len(o.Command) == 4 && o.Command[3] != ""
		case "revisions":
			valid = o.Command[2] == "list" && len(o.Command) == 4 && o.Command[3] != "" || o.Command[2] == "show" && len(o.Command) == 5 && o.Command[3] != "" && o.Command[4] != ""
		}
	}
	if isEvaluationCancellation(o) && len(o.Command) == 4 && o.Command[3] != "" {
		valid = true
	}
	if isEvaluationWrite(o) {
		valid = true
	}
	if !valid {
		return ports.Failure("usage", "Run evaluation --help for the case, run, and revision commands and their IDs.")
	}
	if scope && o.Scope.Tenant == "" {
		return ports.Failure("usage", "Supply --tenant, or select a saved household context.")
	}
	return nil
}
func (r Runner) evaluationCommand(ctx context.Context, o Options, token string) error {
	if r.EvaluationAPI == nil {
		return ports.Failure("configuration", "Evaluation commands are not available. Update the CLI and try again.")
	}
	api, err := r.EvaluationAPI(o.Server, token)
	if err != nil {
		return err
	}
	if isEvaluationWrite(o) {
		return r.writeEvaluation(ctx, o, api)
	}
	if isEvaluationCancellation(o) {
		return r.cancelEvaluationRun(ctx, o, api)
	}
	var result any
	switch o.Command[1] {
	case "cases":
		if o.Command[2] == "list" {
			result, err = listPages(ctx, o, func(page ports.Page) (ports.Result[[]ports.EvaluationCaseHead], error) {
				return api.EvaluationCases(ctx, o.Scope.Tenant, page)
			})
		} else {
			result, err = api.EvaluationCase(ctx, o.Scope.Tenant, o.Command[3])
		}
	case "revisions":
		if o.Command[2] == "list" {
			result, err = listPages(ctx, o, func(page ports.Page) (ports.Result[[]ports.EvaluationCaseRevision], error) {
				return api.EvaluationRevisions(ctx, o.Scope.Tenant, o.Command[3], page)
			})
		} else {
			result, err = api.EvaluationRevision(ctx, o.Scope.Tenant, o.Command[3], o.Command[4])
		}
	case "runs":
		if o.Command[2] == "list" {
			result, err = listPages(ctx, o, func(page ports.Page) (ports.Result[[]ports.EvaluationRunHead], error) {
				return api.EvaluationRuns(ctx, o.Scope.Tenant, page)
			})
		} else {
			result, err = api.EvaluationRun(ctx, o.Scope.Tenant, o.Command[3])
		}
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.evaluation.read.completed")
	return r.Output.Result(result)
}

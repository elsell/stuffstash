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
	if !valid {
		return ports.Failure("usage", "Use evaluation cases|runs list or show ID, evaluation runs cancel RUN_ID, or evaluation revisions list CASE_ID or show CASE_ID REVISION_ID.")
	}
	if scope && o.Scope.Tenant == "" {
		return ports.Failure("usage", "Supply --tenant, or choose a saved household context.")
	}
	return nil
}
func (r Runner) evaluationCommand(ctx context.Context, o Options, token string) error {
	if r.EvaluationAPI == nil {
		return ports.Failure("configuration", "Evaluation inspection is not available. Update the CLI and try again.")
	}
	api, err := r.EvaluationAPI(o.Server, token)
	if err != nil {
		return err
	}
	if isEvaluationCancellation(o) {
		return r.cancelEvaluationRun(ctx, o, api)
	}
	var result any
	switch o.Command[1] {
	case "cases":
		if o.Command[2] == "list" {
			result, err = api.EvaluationCases(ctx, o.Scope.Tenant, o.Page)
		} else {
			result, err = api.EvaluationCase(ctx, o.Scope.Tenant, o.Command[3])
		}
	case "revisions":
		if o.Command[2] == "list" {
			result, err = api.EvaluationRevisions(ctx, o.Scope.Tenant, o.Command[3], o.Page)
		} else {
			result, err = api.EvaluationRevision(ctx, o.Scope.Tenant, o.Command[3], o.Command[4])
		}
	case "runs":
		if o.Command[2] == "list" {
			result, err = api.EvaluationRuns(ctx, o.Scope.Tenant, o.Page)
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

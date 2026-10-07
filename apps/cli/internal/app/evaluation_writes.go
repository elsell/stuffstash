package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func (r Runner) writeEvaluation(ctx context.Context, o Options, api ports.EvaluationAPI) error {
	action, detail, target := "case.create", "Create an evaluation case and its first revision.", "new case"
	notice := "Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant)
	switch o.Command[1] {
	case "revisions":
		var input evaluationRevisionInput
		if err := decodeEvaluationWrite(o.RequestBody, &input); err != nil {
			return err
		}
		action, detail, target = "revision.create", "Append an evaluation case revision using the supplied expectedRevision.", o.Command[3]
		notice += ". Case: " + strconv.Quote(target) + ". Expected revision: " + strconv.FormatInt(input.ExpectedRevision, 10)
	case "runs":
		var input evaluationQueueInput
		if err := decodeEvaluationWrite(o.RequestBody, &input); err != nil {
			return err
		}
		action, detail = "run.queue", "Queue a background text-only evaluation that may call configured model providers."
		notice += ". Workflow: " + strconv.Quote(input.WorkflowID) + ". Revision: " + strconv.Quote(input.RevisionID) + "; " + strconv.Itoa(len(input.Cases)) + " case(s)"
	default:
		notice += ". Case: " + strconv.Quote(target)
	}
	if err := r.Output.Notice(notice + ". " + detail); err != nil {
		return err
	}
	if err := r.confirmAction(ctx, o, "Evaluation: "+action, "Apply evaluation change", detail); err != nil {
		return err
	}
	var result any
	var err error
	switch o.Command[1] {
	case "cases":
		result, err = api.CreateEvaluationCase(ctx, o.Scope.Tenant, o.RequestBody)
	case "revisions":
		result, err = api.CreateEvaluationRevision(ctx, o.Scope.Tenant, target, o.RequestBody)
	case "runs":
		result, err = api.CreateEvaluationRun(ctx, o.Scope.Tenant, o.RequestBody)
	}
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "conflict":
				return ports.Failure("conflict", "The server rejected the evaluation change. Run evaluation cases show CASE_ID or evaluation runs show RUN_ID. Examine the revisions before you try again.")
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "The evaluation change result is unknown. Examine evaluation cases list or evaluation runs list before you try again.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.evaluation."+action+".completed")
	return r.Output.Result(result)
}

package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func (r Runner) writeWorkflow(ctx context.Context, o Options, api ports.WorkflowsAPI) error {
	action, target, detail := "create", "new workflow", "Create a workflow revision without changing the selected workflow."
	if o.Command[1] == "revisions" {
		action = "revision.create"
		target = o.Command[3]
		detail = "Append a revision using the supplied expectedRevision. The selected workflow does not change."
	}
	if o.Command[1] == "activate" {
		action = "activate"
		target = o.Command[2]
		detail = "Change the selected workflow to this revision. The server must accept the supplied evaluation evidence and expected selection."
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Workflow: " + strconv.Quote(target) + ". " + detail); err != nil {
		return err
	}
	if err := r.confirmAction(ctx, o, "Workflow: "+action, "Apply workflow change", detail); err != nil {
		return err
	}
	var result ports.Result[ports.WorkflowRevision]
	var err error
	switch action {
	case "create":
		result, err = api.CreateWorkflow(ctx, o.Scope.Tenant, o.RequestBody)
	case "revision.create":
		result, err = api.CreateWorkflowRevision(ctx, o.Scope.Tenant, target, o.RequestBody)
	case "activate":
		result, err = api.ActivateWorkflow(ctx, o.Scope.Tenant, target, o.RequestBody)
	}
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "conflict":
				return ports.Failure("conflict", "The server rejected the workflow change. Examine workflows show and workflows selection show, then review the supplied revision or evaluation evidence before you try again.")
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "The workflow change result is unknown. Examine workflows list, show and selection show before you try again.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.workflow."+action+".completed")
	return r.Output.Result(result)
}

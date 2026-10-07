package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isWorkflowCommand(o Options) bool { return len(o.Command) > 0 && o.Command[0] == "workflows" }
func workflowList(o Options) bool {
	c := o.Command
	return len(c) == 2 && c[1] == "list" || len(c) == 4 && c[1] == "revisions" && c[2] == "list"
}
func validateWorkflows(o Options, scope bool) error {
	c := o.Command
	valid := isWorkflowWrite(o) || workflowList(o) || len(c) == 3 && (c[1] == "show" && c[2] != "" || c[1] == "selection" && c[2] == "show") || len(c) == 5 && c[1] == "revisions" && c[2] == "show" && c[3] != "" && c[4] != ""
	if !valid || len(c) == 4 && c[3] == "" {
		return ports.Failure("usage", "Use workflows list, show WORKFLOW_ID, revisions list WORKFLOW_ID, revisions show WORKFLOW_ID REVISION_ID, selection show, create --input FILE, revisions create WORKFLOW_ID --input FILE, or activate WORKFLOW_ID --input FILE.")
	}
	if scope && o.Scope.Tenant == "" {
		return ports.Failure("usage", "Supply --tenant, or select a saved household context.")
	}
	return nil
}
func (r Runner) workflowCommand(ctx context.Context, o Options, token string) error {
	if r.WorkflowsAPI == nil {
		return ports.Failure("configuration", "Workflow commands are not available. Update the CLI and try again.")
	}
	api, err := r.WorkflowsAPI(o.Server, token)
	if err != nil {
		return err
	}
	if isWorkflowWrite(o) {
		return r.writeWorkflow(ctx, o, api)
	}
	var result any
	switch o.Command[1] {
	case "list":
		result, err = api.Workflows(ctx, o.Scope.Tenant, o.Page)
	case "show":
		result, err = api.Workflow(ctx, o.Scope.Tenant, o.Command[2])
	case "selection":
		result, err = api.WorkflowSelection(ctx, o.Scope.Tenant)
	case "revisions":
		if o.Command[2] == "list" {
			result, err = api.WorkflowRevisions(ctx, o.Scope.Tenant, o.Command[3], o.Page)
		} else {
			result, err = api.WorkflowRevision(ctx, o.Scope.Tenant, o.Command[3], o.Command[4])
		}
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.workflow.read.completed")
	return r.Output.Result(result)
}

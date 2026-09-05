package agentmodel

import (
	"context"
	"errors"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestWorkflowQueriesAuthorizeBeforeDependencies(t *testing.T) {
	service := NewWorkflowQueryService(WorkflowQueryDependencies{Authorizer: denyTenantAuthorizer{}})
	access := EvaluationRunAccess{Principal: testPrincipal(), TenantID: "home"}
	for _, call := range []func() error{
		func() error {
			_, err := service.Get(context.Background(), GetWorkflowInput{EvaluationRunAccess: access})
			return err
		},
		func() error {
			_, err := service.List(context.Background(), ListWorkflowsInput{EvaluationRunAccess: access})
			return err
		},
		func() error {
			_, err := service.History(context.Background(), ListWorkflowRevisionsInput{EvaluationRunAccess: access})
			return err
		},
		func() error { _, err := service.Selection(context.Background(), access); return err },
	} {
		if err := call(); !errors.Is(err, ports.ErrForbidden) {
			t.Fatalf("authorization: %v", err)
		}
	}
}
func TestWorkflowQueriesPreserveRevisionsScopeAndAudit(t *testing.T) {
	commands, queue, store := evaluationCommandSetup(t)
	ctx := context.Background()
	service := NewWorkflowQueryService(WorkflowQueryDependencies{Authorizer: commands.Authorizer, Repository: store, Discovery: store, Audit: store, IDs: commands.IDs, Clock: commands.Clock})
	access := queue.EvaluationRunAccess
	original, err := service.Get(ctx, GetWorkflowInput{EvaluationRunAccess: access, WorkflowID: queue.WorkflowID})
	if err != nil {
		t.Fatal(err)
	}
	value := original.Snapshot()
	value.ID = "second"
	value.Number = 2
	definition := value.Definition.Settings()
	definition.Name = "New name"
	value.Definition, err = model.NewWorkflowDefinition(definition, value.Limits)
	if err != nil {
		t.Fatal(err)
	}
	revision, err := model.NewWorkflowRevision(value)
	if err != nil {
		t.Fatal(err)
	}
	record, ok := audit.NewRecord("second-audit", audit.TenantID(access.TenantID), "", "owner", audit.ActionConversationWorkflowRevisionCreated, audit.SourceAPI, audit.TargetConversationWorkflow, string(queue.WorkflowID), commands.Clock.Now(), "", nil)
	if !ok {
		t.Fatal("audit invalid")
	}
	if err := store.AppendWorkflowRevision(ctx, revision, 1, record); err != nil {
		t.Fatal(err)
	}
	latest, err := service.Get(ctx, GetWorkflowInput{EvaluationRunAccess: access, WorkflowID: queue.WorkflowID})
	if err != nil || latest.Snapshot().ID != "second" {
		t.Fatalf("latest revision: %v", err)
	}
	prior, err := service.Get(ctx, GetWorkflowInput{EvaluationRunAccess: access, WorkflowID: queue.WorkflowID, RevisionID: queue.RevisionID})
	if err != nil || prior.Snapshot().ID != queue.RevisionID {
		t.Fatal("immutable read")
	}
	list, err := service.List(ctx, ListWorkflowsInput{EvaluationRunAccess: access, Limit: 1})
	if err != nil || len(list.Items) != 1 || list.Items[0].Name != "New name" || list.NextCursor != nil {
		t.Fatalf("heads: %+v %v", list, err)
	}
	history, err := service.History(ctx, ListWorkflowRevisionsInput{EvaluationRunAccess: access, WorkflowID: queue.WorkflowID, Limit: 1})
	if err != nil || len(history.Items) != 1 || history.NextCursor == nil {
		t.Fatalf("history: %+v %v", history, err)
	}
	next, err := service.History(ctx, ListWorkflowRevisionsInput{EvaluationRunAccess: access, WorkflowID: queue.WorkflowID, Limit: 1, Cursor: *history.NextCursor})
	if err != nil || len(next.Items) != 1 || next.Items[0].Snapshot().ID != "second" || next.NextCursor != nil {
		t.Fatal("history cursor")
	}
	wrong := access
	wrong.TenantID = "outside"
	if _, err := service.Get(ctx, GetWorkflowInput{EvaluationRunAccess: wrong, WorkflowID: queue.WorkflowID, RevisionID: queue.RevisionID}); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("cross-tenant read: %v", err)
	}
	if _, err := service.History(ctx, ListWorkflowRevisionsInput{EvaluationRunAccess: access, WorkflowID: "other", Cursor: *history.NextCursor}); !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("cross-workflow cursor: %v", err)
	}
	invalid := appsupport.EncodePageCursor("conversation_workflows", "outside", "anything")
	if _, err := service.List(ctx, ListWorkflowsInput{EvaluationRunAccess: access, Cursor: *invalid}); !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("cross-tenant cursor: %v", err)
	}
	selected, err := service.Selection(ctx, access)
	if err != nil || selected != nil {
		t.Fatal("default selection")
	}
	records, err := store.ListTenantAuditRecords(ctx, access.TenantID, ports.AuditRecordPageRequest{Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	reads := 0
	for _, record := range records {
		if record.Action == audit.ActionConversationWorkflowViewed || record.Action == audit.ActionConversationWorkflowListed {
			reads++
		}
	}
	if reads != 7 {
		t.Fatalf("expected seven successful read audits, got %d", reads)
	}
	broken := NewWorkflowQueryService(WorkflowQueryDependencies{Authorizer: commands.Authorizer, Repository: store, Discovery: store, Audit: workflowQueryFailingAudit{AuditRepository: store}, IDs: commands.IDs, Clock: commands.Clock})
	if _, err := broken.Get(ctx, GetWorkflowInput{EvaluationRunAccess: access, WorkflowID: queue.WorkflowID}); err == nil {
		t.Fatal("read escaped audit failure")
	}
	if _, err := broken.List(ctx, ListWorkflowsInput{EvaluationRunAccess: access}); err == nil {
		t.Fatal("list escaped audit failure")
	}
	if _, err := broken.Selection(ctx, access); err == nil {
		t.Fatal("selection escaped audit failure")
	}
}

type workflowQueryFailingAudit struct{ ports.AuditRepository }

func (workflowQueryFailingAudit) SaveAuditRecord(context.Context, audit.Record) error {
	return errors.New("audit unavailable")
}

func TestWorkflowQueryFullPages(t *testing.T) {
	deps, input, store := evaluationCommandSetup(t)
	ctx := context.Background()
	service := NewWorkflowQueryService(WorkflowQueryDependencies{Authorizer: deps.Authorizer, Repository: store, Discovery: store, Audit: store, IDs: deps.IDs, Clock: deps.Clock})
	original, _, err := store.WorkflowRevision(ctx, input.TenantID, input.WorkflowID, input.RevisionID)
	if err != nil {
		t.Fatal(err)
	}
	appendRevision := func(workflow model.WorkflowID, number int) {
		value := original.Snapshot()
		value.WorkflowID = workflow
		value.Number = number
		value.ID = model.WorkflowRevisionID(fmt.Sprintf("%s-%03d", workflow, number))
		revision, err := model.NewWorkflowRevision(value)
		if err != nil {
			t.Fatal(err)
		}
		record, ok := audit.NewRecord(audit.ID(value.ID), audit.TenantID(input.TenantID), "", "owner", audit.ActionConversationWorkflowRevisionCreated, audit.SourceAPI, audit.TargetConversationWorkflow, string(workflow), value.CreatedAt, "", nil)
		if !ok {
			t.Fatal("fixture audit invalid")
		}
		if err := store.AppendWorkflowRevision(ctx, revision, number-1, record); err != nil {
			t.Fatal(err)
		}
	}
	for i := 1; i <= 99; i++ {
		appendRevision(model.WorkflowID(fmt.Sprintf("w%03d", i)), 1)
		appendRevision(input.WorkflowID, i+1)
	}
	heads, err := service.List(ctx, ListWorkflowsInput{EvaluationRunAccess: input.EvaluationRunAccess, Limit: 100})
	if err != nil || len(heads.Items) != 100 || heads.NextCursor != nil {
		t.Fatal("exactly full head page")
	}
	history, err := service.History(ctx, ListWorkflowRevisionsInput{EvaluationRunAccess: input.EvaluationRunAccess, WorkflowID: input.WorkflowID, Limit: 100})
	if err != nil || len(history.Items) != 100 || history.NextCursor != nil {
		t.Fatal("exactly full history page")
	}
	appendRevision("w100", 1)
	appendRevision(input.WorkflowID, 101)
	heads, err = service.List(ctx, ListWorkflowsInput{EvaluationRunAccess: input.EvaluationRunAccess, Limit: 100})
	if err != nil || len(heads.Items) != 100 || heads.NextCursor == nil {
		t.Fatal("missing head continuation")
	}
	history, err = service.History(ctx, ListWorkflowRevisionsInput{EvaluationRunAccess: input.EvaluationRunAccess, WorkflowID: input.WorkflowID, Limit: 100})
	if err != nil || len(history.Items) != 100 || history.NextCursor == nil {
		t.Fatal("missing history continuation")
	}
	tail, err := service.History(ctx, ListWorkflowRevisionsInput{EvaluationRunAccess: input.EvaluationRunAccess, WorkflowID: input.WorkflowID, Limit: 100, Cursor: *history.NextCursor})
	if err != nil || len(tail.Items) != 1 || tail.Items[0].Snapshot().Number != 101 || tail.NextCursor != nil {
		t.Fatal("history continuation lost revision")
	}
}

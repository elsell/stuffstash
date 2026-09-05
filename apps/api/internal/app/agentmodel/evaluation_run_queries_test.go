package agentmodel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	fixture "github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

func TestEvaluationRunQueriesAuthorizeScopeAndAudit(t *testing.T) {
	commands, input, store := evaluationCommandSetup(t)
	ctx := context.Background()
	queued, err := NewEvaluationRunCommandService(commands).Queue(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveEvaluationRun(ctx, fixture.Run(t, "second"), 0, fixture.Record(t, "second-created", "second", audit.ActionConversationEvaluationRunCreated)); err != nil {
		t.Fatal(err)
	}
	deps := EvaluationRunQueryDependencies{Authorizer: commands.Authorizer, Runs: store, Audit: store, IDs: commands.IDs, Clock: commands.Clock}
	service := NewEvaluationRunQueryService(deps)
	query := GetEvaluationRunInput{EvaluationRunAccess: input.EvaluationRunAccess, RunID: queued.Snapshot().Input.ID}
	read, err := service.Get(ctx, query)
	if err != nil || read.Snapshot().Input.ID != query.RunID {
		t.Fatal("run read failed")
	}
	page, err := service.List(ctx, ListEvaluationRunsInput{EvaluationRunAccess: input.EvaluationRunAccess, Limit: 1})
	if err != nil || len(page.Items) != 1 || page.NextCursor == nil {
		t.Fatal("run pagination failed")
	}
	next, err := service.List(ctx, ListEvaluationRunsInput{EvaluationRunAccess: input.EvaluationRunAccess, Limit: 1, Cursor: *page.NextCursor})
	if err != nil || len(next.Items) != 1 || next.NextCursor != nil || next.Items[0].ID == page.Items[0].ID {
		t.Fatal("run pagination repeated or omitted rows")
	}
	wrong := input.EvaluationRunAccess
	wrong.TenantID = "outside"
	if _, err := service.List(ctx, ListEvaluationRunsInput{EvaluationRunAccess: wrong, Cursor: *page.NextCursor}); !errors.Is(err, apperrors.ErrValidation) {
		t.Fatal("cross-tenant cursor accepted")
	}
	if _, err := service.Get(ctx, GetEvaluationRunInput{EvaluationRunAccess: wrong, RunID: query.RunID}); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatal("cross-tenant run exposed")
	}
	records, err := store.ListTenantAuditRecords(ctx, input.TenantID, ports.AuditRecordPageRequest{Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	viewed, listed := 0, 0
	for _, record := range records {
		if record.Action == "conversation_evaluation_run.viewed" {
			viewed++
		}
		if record.Action == "conversation_evaluation_run.listed" {
			listed++
		}
	}
	if viewed != 1 || listed != 2 {
		t.Fatal("read audit missing")
	}
	deps.Audit = unavailableEvaluationAudit{AuditRepository: store}
	unavailable := NewEvaluationRunQueryService(deps)
	if _, err := unavailable.Get(ctx, query); err == nil {
		t.Fatal("run exposed without audit")
	}
	if _, err := unavailable.List(ctx, ListEvaluationRunsInput{EvaluationRunAccess: input.EvaluationRunAccess}); err == nil {
		t.Fatal("runs listed without audit")
	}
	denied := NewEvaluationRunQueryService(EvaluationRunQueryDependencies{Authorizer: denyTenantAuthorizer{}})
	if _, err := denied.Get(ctx, query); !errors.Is(err, ports.ErrForbidden) {
		t.Fatal("run read accessed dependencies before permission")
	}
	if _, err := denied.List(ctx, ListEvaluationRunsInput{EvaluationRunAccess: input.EvaluationRunAccess}); !errors.Is(err, ports.ErrForbidden) {
		t.Fatal("run list accessed dependencies before permission")
	}
}

func TestEvaluationRunQueriesDistinguishFullFinalPageFromMoreResults(t *testing.T) {
	commands, input, store := evaluationCommandSetup(t)
	ctx := context.Background()
	for index := 0; index < 100; index++ {
		id := fmt.Sprintf("page-%03d", index)
		if err := store.SaveEvaluationRun(ctx, fixture.Run(t, id), 0, fixture.Record(t, "created-"+id, id, audit.ActionConversationEvaluationRunCreated)); err != nil {
			t.Fatal(err)
		}
	}
	service := NewEvaluationRunQueryService(EvaluationRunQueryDependencies{Authorizer: commands.Authorizer, Runs: store, Audit: store, IDs: commands.IDs, Clock: commands.Clock})
	query := ListEvaluationRunsInput{EvaluationRunAccess: input.EvaluationRunAccess, Limit: 100}
	page, err := service.List(ctx, query)
	if err != nil || len(page.Items) != 100 || page.NextCursor != nil {
		t.Fatal("exactly full final page reported more rows")
	}
	if err := store.SaveEvaluationRun(ctx, fixture.Run(t, "page-100"), 0, fixture.Record(t, "created-page-100", "page-100", audit.ActionConversationEvaluationRunCreated)); err != nil {
		t.Fatal(err)
	}
	page, err = service.List(ctx, query)
	if err != nil || len(page.Items) != 100 || page.NextCursor == nil {
		t.Fatal("additional row missed at limit")
	}
	query.Cursor = *page.NextCursor
	tail, err := service.List(ctx, query)
	if err != nil || len(tail.Items) != 1 || tail.NextCursor != nil || tail.Items[0].ID != "page-100" {
		t.Fatal("final row missing")
	}
}

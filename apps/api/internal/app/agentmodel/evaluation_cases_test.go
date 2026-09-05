package agentmodel

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestEvaluationCaseServiceAuthorizesAndAuditsReads(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	name, _ := tenant.NewName("Home")
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: "tenant-home", Name: name}); err != nil {
		t.Fatal(err)
	}
	ids := &workflowSequenceIDs{}
	deps := EvaluationCaseDependencies{Authorizer: allowTenantConfigureAuthorizer{}, Repository: store, Audit: store, IDs: ids, Clock: fixedClock{}, DefaultPageLimit: 1, MaxPageLimit: 100}
	service := NewEvaluationCaseService(deps)
	input := SaveEvaluationCaseInput{EvaluationCaseAccess: EvaluationCaseAccess{Principal: testPrincipal(), TenantID: "tenant-home", Source: audit.SourceAPI}, Definition: domain.EvaluationCaseDefinitionInput{Title: "Clothes", Utterance: "Find clothes", Expectations: domain.EvaluationExpectations{Kind: domain.EvaluationOutcomeAnswer}}}
	first, err := service.SaveRevision(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SaveRevision(ctx, input); err != nil {
		t.Fatal(err)
	}
	query := GetEvaluationCaseInput{EvaluationCaseAccess: input.EvaluationCaseAccess, CaseID: first.Snapshot().CaseID}
	current, err := service.Get(ctx, query)
	if err != nil || current.Snapshot().ID != first.Snapshot().ID {
		t.Fatalf("current case: %+v %v", current, err)
	}
	page, err := service.List(ctx, ListEvaluationCasesInput{EvaluationCaseAccess: input.EvaluationCaseAccess, Limit: 1})
	if err != nil || len(page.Items) != 1 || page.NextCursor == nil {
		t.Fatalf("page: %+v %v", page, err)
	}
	wrong := input.EvaluationCaseAccess
	wrong.TenantID = "other"
	if _, err := service.List(ctx, ListEvaluationCasesInput{EvaluationCaseAccess: wrong, Cursor: *page.NextCursor}); !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("cross-tenant cursor accepted: %v", err)
	}
	records, err := store.ListTenantAuditRecords(ctx, "tenant-home", ports.AuditRecordPageRequest{Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	actions := map[audit.Action]int{}
	for _, record := range records {
		actions[record.Action]++
		for _, value := range record.Metadata {
			if value == input.Definition.Utterance {
				t.Fatal("utterance leaked into audit")
			}
		}
	}
	if actions[audit.ActionConversationEvaluationCaseRevisionCreated] != 2 || actions[audit.ActionConversationEvaluationCaseViewed] != 1 || actions[audit.ActionConversationEvaluationCaseListed] != 1 {
		t.Fatalf("read audit missing: %v", actions)
	}
	for count := 2; count < 100; count++ {
		if _, err := service.SaveRevision(ctx, input); err != nil {
			t.Fatal(err)
		}
	}
	full, err := service.List(ctx, ListEvaluationCasesInput{EvaluationCaseAccess: input.EvaluationCaseAccess, Limit: 100})
	if err != nil || len(full.Items) != 100 || full.NextCursor != nil {
		t.Fatalf("exactly full terminal page: %+v %v", full, err)
	}
	if _, err := service.SaveRevision(ctx, input); err != nil {
		t.Fatal(err)
	}
	full, err = service.List(ctx, ListEvaluationCasesInput{EvaluationCaseAccess: input.EvaluationCaseAccess, Limit: 100})
	if err != nil || len(full.Items) != 100 || full.NextCursor == nil {
		t.Fatalf("full page with remaining case: %+v %v", full, err)
	}
	tail, err := service.List(ctx, ListEvaluationCasesInput{EvaluationCaseAccess: input.EvaluationCaseAccess, Limit: 100, Cursor: *full.NextCursor})
	if err != nil || len(tail.Items) != 1 || tail.NextCursor != nil {
		t.Fatalf("last case: %+v %v", tail, err)
	}
	for _, previous := range full.Items {
		if previous.ID == tail.Items[0].ID {
			t.Fatal("case repeated across pages")
		}
	}
	deps.Authorizer = denyTenantAuthorizer{}
	denied := NewEvaluationCaseService(deps)
	if _, err := denied.SaveRevision(ctx, input); !errors.Is(err, ports.ErrForbidden) {
		t.Fatal("denied save accepted")
	}
	if _, err := denied.Get(ctx, query); !errors.Is(err, ports.ErrForbidden) {
		t.Fatal("denied read accepted")
	}
	if _, err := denied.List(ctx, ListEvaluationCasesInput{EvaluationCaseAccess: input.EvaluationCaseAccess}); !errors.Is(err, ports.ErrForbidden) {
		t.Fatal("denied list accepted")
	}
	deps.Authorizer = allowTenantConfigureAuthorizer{}
	deps.Audit = unavailableEvaluationAudit{AuditRepository: store}
	unavailable := NewEvaluationCaseService(deps)
	if _, err := unavailable.Get(ctx, query); err == nil {
		t.Fatal("read succeeded without audit")
	}
	if _, err := unavailable.List(ctx, ListEvaluationCasesInput{EvaluationCaseAccess: input.EvaluationCaseAccess}); err == nil {
		t.Fatal("list succeeded without audit")
	}
}

type unavailableEvaluationAudit struct{ ports.AuditRepository }

func (unavailableEvaluationAudit) SaveAuditRecord(context.Context, audit.Record) error {
	return errors.New("audit unavailable")
}

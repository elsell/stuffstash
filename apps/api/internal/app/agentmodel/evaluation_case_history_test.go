package agentmodel

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestEvaluationCaseHistoryAuthorizesAuditsAndCapsPages(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	name, _ := tenant.NewName("Home")
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: "home", Name: name}); err != nil {
		t.Fatal(err)
	}
	deps := EvaluationCaseDependencies{Authorizer: allowTenantConfigureAuthorizer{}, Repository: store, Audit: store, IDs: &workflowSequenceIDs{}, Clock: fixedClock{}, DefaultPageLimit: 20, MaxPageLimit: 100}
	service := NewEvaluationCaseService(deps)
	input := SaveEvaluationCaseInput{EvaluationCaseAccess: EvaluationCaseAccess{Principal: testPrincipal(), TenantID: "home", Source: audit.SourceAPI}, Definition: domain.EvaluationCaseDefinitionInput{Title: "Clothes", Utterance: "Find baby clothes", Expectations: domain.EvaluationExpectations{Kind: domain.EvaluationOutcomeAnswer}}}
	first, err := service.SaveRevision(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	query := ListEvaluationCaseRevisionsInput{EvaluationCaseAccess: input.EvaluationCaseAccess, CaseID: first.Snapshot().CaseID, Limit: 100}
	if _, err := service.History(ctx, query); err != nil {
		t.Fatal(err)
	}
	records, err := store.ListTenantAuditRecords(ctx, "home", ports.AuditRecordPageRequest{Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, record := range records {
		if record.Action == audit.ActionConversationEvaluationCaseListed && record.TargetID == string(query.CaseID) && record.Metadata["returned_count"] == "1" {
			found = true
		}
	}
	if !found {
		t.Fatal("case history read audit absent")
	}
	input.CaseID = query.CaseID
	for number := 2; number <= 100; number++ {
		input.ExpectedRevision = number - 1
		if _, err := service.SaveRevision(ctx, input); err != nil {
			t.Fatal(err)
		}
	}
	page, err := service.History(ctx, query)
	if err != nil || len(page.Items) != 100 || page.NextCursor != nil {
		t.Fatalf("exact full page: %+v %v", page, err)
	}
	input.ExpectedRevision = 100
	if _, err := service.SaveRevision(ctx, input); err != nil {
		t.Fatal(err)
	}
	page, err = service.History(ctx, query)
	if err != nil || len(page.Items) != 100 || page.NextCursor == nil {
		t.Fatalf("overflow page: %+v %v", page, err)
	}
	query.Cursor = *page.NextCursor
	last, err := service.History(ctx, query)
	if err != nil || len(last.Items) != 1 || last.Items[0].Snapshot().Number != 101 || last.NextCursor != nil {
		t.Fatalf("last page: %+v %v", last, err)
	}
	wrong := query
	wrong.TenantID = "other"
	if _, err := service.History(ctx, wrong); !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("cross tenant cursor: %v", err)
	}
	deps.Authorizer = denyTenantAuthorizer{}
	if _, err := NewEvaluationCaseService(deps).History(ctx, query); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("unauthorized history: %v", err)
	}
	deps.Authorizer = allowTenantConfigureAuthorizer{}
	deps.Audit = unavailableEvaluationAudit{}
	if _, err := NewEvaluationCaseService(deps).History(ctx, query); err == nil {
		t.Fatal("history returned without read audit")
	}
}

package gormstore

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func storedEvaluationRevision(t *testing.T, id string, number int) domain.EvaluationCaseRevision {
	t.Helper()
	definition, err := domain.NewEvaluationCaseDefinition(domain.EvaluationCaseDefinitionInput{Title: "Find clothes", Utterance: "Check out my baby clothes for Sam.",
		Assets: []domain.EvaluationFixtureAsset{
			{ID: "room", Title: "Attic", Kind: domain.EvaluationFixtureLocation},
			{ID: "box", Title: "Blue box", Kind: domain.EvaluationFixtureContainer, ParentID: "room"},
			{ID: "clothes", Title: "3 to 6 months", Description: "Winter clothes", Kind: domain.EvaluationFixtureItem, ParentID: "box", TagNames: []string{"baby", "clothes"}},
		}, Expectations: domain.EvaluationExpectations{Kind: domain.EvaluationOutcomeProposal, ReferencedAssets: []string{"clothes"}, Locations: []domain.EvaluationLocationExpectation{{AssetID: "clothes", AncestorID: "room"}}, Proposals: []domain.EvaluationProposal{{Operation: domain.OperationCheckout, TargetID: "clothes", Details: "For Sam"}}, ForbiddenOperations: []domain.Operation{domain.OperationArchive}}})
	if err != nil {
		t.Fatal(err)
	}
	revision, err := domain.NewEvaluationCaseRevision(domain.EvaluationCaseRevisionInput{ID: domain.EvaluationCaseRevisionID(id), CaseID: "clothes", TenantID: "evaluation-home", AuthorID: "owner", Number: number, Definition: definition, CreatedAt: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	return revision
}
func TestEvaluationCaseRepositoryScopeHistoryAndAtomicAudit(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	verifyEvaluationCaseRepository(t, ctx, store)
}
func verifyEvaluationCaseRepository(t *testing.T, ctx context.Context, store Store) {
	t.Helper()
	saveTenant(t, ctx, store, "evaluation-home", "Home")
	first := storedEvaluationRevision(t, "revision-one", 1)
	record := auditRecord(t, "case-audit-one", "evaluation-home", "", audit.ActionConversationEvaluationCaseRevisionCreated, "conversation_evaluation_case")
	if err := store.AppendEvaluationCaseRevision(ctx, first, 0, record); err != nil {
		t.Fatal(err)
	}
	if _, found, err := store.EvaluationCaseRevision(ctx, "other", "clothes", "revision-one"); err != nil || found {
		t.Fatal("cross tenant read")
	}
	if _, found, err := store.EvaluationCaseRevision(ctx, "evaluation-home", "other", "revision-one"); err != nil || found {
		t.Fatal("cross case read")
	}
	second := storedEvaluationRevision(t, "revision-two", 2)
	secondSnapshot := second.Snapshot()
	updated := secondSnapshot.Definition.Settings()
	updated.Title = "Edited clothes case"
	var updateErr error
	secondSnapshot.Definition, updateErr = domain.NewEvaluationCaseDefinition(updated)
	if updateErr != nil {
		t.Fatal(updateErr)
	}
	second, updateErr = domain.NewEvaluationCaseRevision(secondSnapshot)
	if updateErr != nil {
		t.Fatal(updateErr)
	}
	if err := store.AppendEvaluationCaseRevision(ctx, second, 1, record); err == nil {
		t.Fatal("duplicate audit accepted")
	}
	head, found, err := store.EvaluationCaseHead(ctx, "evaluation-home", "clothes")
	if err != nil || !found || head.LatestRevision != 1 {
		t.Fatalf("audit failure changed head: %+v %v", head, err)
	}
	if _, found, err := store.EvaluationCaseRevision(ctx, "evaluation-home", "clothes", "revision-two"); err != nil || found {
		t.Fatal("audit failure persisted revision")
	}
	record = auditRecord(t, "case-audit-two", "evaluation-home", "", audit.ActionConversationEvaluationCaseRevisionCreated, "conversation_evaluation_case")
	if err := store.AppendEvaluationCaseRevision(ctx, second, 1, record); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendEvaluationCaseRevision(ctx, storedEvaluationRevision(t, "stale", 2), 1, auditRecord(t, "case-stale", "evaluation-home", "", audit.ActionConversationEvaluationCaseRevisionCreated, "conversation_evaluation_case")); !errors.Is(err, ports.ErrEvaluationCaseConflict) {
		t.Fatalf("stale save did not conflict: %v", err)
	}
	saved, found, err := store.EvaluationCaseRevision(ctx, "evaluation-home", "clothes", "revision-one")
	if err != nil || !found || saved.Snapshot().Number != 1 || !reflect.DeepEqual(saved.Snapshot().Definition.Settings(), first.Snapshot().Definition.Settings()) {
		t.Fatal("history changed")
	}
	rows, err := store.ListEvaluationCases(ctx, "evaluation-home", ports.EvaluationCasePageRequest{Limit: 1})
	if err != nil || len(rows) != 1 || rows[0].LatestRevision != 2 || rows[0].LatestRevisionID != "revision-two" || rows[0].Title != "Edited clothes case" {
		t.Fatalf("list: %+v %v", rows, err)
	}
	rows, err = store.ListEvaluationCases(ctx, "evaluation-home", ports.EvaluationCasePageRequest{AfterID: "clothes", Limit: 1})
	if err != nil || len(rows) != 0 {
		t.Fatal("cursor not exclusive")
	}
	rows, err = store.ListEvaluationCases(ctx, "other", ports.EvaluationCasePageRequest{Limit: 1})
	if err != nil || len(rows) != 0 {
		t.Fatal("list leaked tenant")
	}
	if err := store.AppendEvaluationCaseRevision(ctx, storedEvaluationRevision(t, "cross-audit", 3), 2, auditRecord(t, "case-cross", "other", "", audit.ActionConversationEvaluationCaseRevisionCreated, "conversation_evaluation_case")); err == nil {
		t.Fatal("cross tenant audit accepted")
	}
	for _, caseID := range []domain.EvaluationCaseID{"case-c", "case-a", "case-b"} {
		snapshot := first.Snapshot()
		snapshot.CaseID = caseID
		snapshot.ID = domain.EvaluationCaseRevisionID("revision-" + string(caseID))
		revision, err := domain.NewEvaluationCaseRevision(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.AppendEvaluationCaseRevision(ctx, revision, 0, auditRecord(t, "audit-"+string(caseID), "evaluation-home", "", audit.ActionConversationEvaluationCaseRevisionCreated, "conversation_evaluation_case")); err != nil {
			t.Fatal(err)
		}
	}
	ids := []domain.EvaluationCaseID{}
	cursor := domain.EvaluationCaseID("")
	for range 4 {
		page, err := store.ListEvaluationCases(ctx, "evaluation-home", ports.EvaluationCasePageRequest{AfterID: cursor, Limit: 2})
		if err != nil {
			t.Fatal(err)
		}
		if len(page) > 2 {
			t.Fatal("page exceeded limit")
		}
		if len(page) == 0 {
			break
		}
		for _, head := range page {
			ids = append(ids, head.ID)
		}
		cursor = page[len(page)-1].ID
	}
	if !slices.Equal(ids, []domain.EvaluationCaseID{"case-a", "case-b", "case-c", "clothes"}) {
		t.Fatalf("page traversal lost or repeated cases: %v", ids)
	}
	for _, limit := range []int{0, 101} {
		if _, err := store.ListEvaluationCases(ctx, "evaluation-home", ports.EvaluationCasePageRequest{Limit: limit}); !errors.Is(err, ports.ErrInvalidEvaluationCasePage) {
			t.Fatalf("invalid limit %d: %v", limit, err)
		}
	}

}

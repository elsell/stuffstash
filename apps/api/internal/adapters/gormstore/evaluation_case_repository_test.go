package gormstore

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func storedEvaluationRevision(t *testing.T, id string, number int) domain.EvaluationCaseRevision {
	t.Helper()
	definition, err := domain.NewEvaluationCaseDefinition(domain.EvaluationCaseDefinitionInput{Title: "Find clothes", Utterance: "Where are my clothes?", Expectations: domain.EvaluationExpectations{Kind: domain.EvaluationOutcomeAnswer}})
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
	saveTenant(t, ctx, store, "evaluation-home", "Home")
	first := storedEvaluationRevision(t, "revision-one", 1)
	record := auditRecord(t, "case-audit-one", "evaluation-home", "", audit.ActionConversationEvaluationCaseRevisionCreated)
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
	record = auditRecord(t, "case-audit-two", "evaluation-home", "", audit.ActionConversationEvaluationCaseRevisionCreated)
	if err := store.AppendEvaluationCaseRevision(ctx, second, 1, record); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendEvaluationCaseRevision(ctx, storedEvaluationRevision(t, "stale", 2), 1, auditRecord(t, "case-stale", "evaluation-home", "", audit.ActionConversationEvaluationCaseRevisionCreated)); !errors.Is(err, ports.ErrEvaluationCaseConflict) {
		t.Fatalf("stale save did not conflict: %v", err)
	}
	saved, found, err := store.EvaluationCaseRevision(ctx, "evaluation-home", "clothes", "revision-one")
	if err != nil || !found || saved.Snapshot().Number != 1 {
		t.Fatal("history changed")
	}
	rows, err := store.ListEvaluationCases(ctx, "evaluation-home", ports.EvaluationCasePageRequest{Limit: 1})
	if err != nil || len(rows) != 1 || rows[0].LatestRevision != 2 || rows[0].LatestRevisionID != "revision-two" {
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
	if err := store.AppendEvaluationCaseRevision(ctx, storedEvaluationRevision(t, "cross-audit", 3), 2, auditRecord(t, "case-cross", "other", "", audit.ActionConversationEvaluationCaseRevisionCreated)); err == nil {
		t.Fatal("cross tenant audit accepted")
	}
}

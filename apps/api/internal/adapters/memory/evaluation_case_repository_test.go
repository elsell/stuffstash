package memory

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestMemoryEvaluationCaseAtomicHistoryAndConcurrentSaves(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	name, _ := tenant.NewName("Home")
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: "home", Name: name}); err != nil {
		t.Fatal(err)
	}
	definition, err := domain.NewEvaluationCaseDefinition(domain.EvaluationCaseDefinitionInput{Title: "Clothes", Utterance: "Where are my clothes?", Expectations: domain.EvaluationExpectations{Kind: domain.EvaluationOutcomeAnswer}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	revision := func(id string, n int) domain.EvaluationCaseRevision {
		t.Helper()
		value, err := domain.NewEvaluationCaseRevision(domain.EvaluationCaseRevisionInput{ID: domain.EvaluationCaseRevisionID(id), CaseID: "case", TenantID: "home", AuthorID: "owner", Number: n, Definition: definition, CreatedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	record := func(id string) audit.Record {
		t.Helper()
		value, ok := audit.NewRecord(audit.ID(id), "home", "", "owner", audit.ActionConversationEvaluationCaseRevisionCreated, audit.SourceAPI, "conversation_evaluation_case", "case", now, "", nil)
		if !ok {
			t.Fatal("audit fixture")
		}
		return value
	}
	first := revision("one", 1)
	if err := store.AppendEvaluationCaseRevision(ctx, first, 0, record("one")); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendEvaluationCaseRevision(ctx, revision("two", 2), 1, record("one")); err == nil {
		t.Fatal("duplicate audit accepted")
	}
	if _, found, err := store.EvaluationCaseRevision(ctx, "home", "case", "two"); err != nil || found {
		t.Fatal("failed audit left revision")
	}
	revisions := []domain.EvaluationCaseRevision{revision("two-a", 2), revision("two-b", 2)}
	audits := []audit.Record{record("two-a"), record("two-b")}
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for index := range revisions {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results <- store.AppendEvaluationCaseRevision(ctx, revisions[index], 1, audits[index])
		}(index)
	}
	wg.Wait()
	close(results)
	wins, conflicts := 0, 0
	for err := range results {
		if err == nil {
			wins++
		} else if errors.Is(err, ports.ErrEvaluationCaseConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("CAS: %d wins %d conflicts", wins, conflicts)
	}
	if _, found, err := store.EvaluationCaseRevision(ctx, "other", "case", "one"); err != nil || found {
		t.Fatal("cross tenant revision")
	}
	if _, found, err := store.EvaluationCaseRevision(ctx, "home", "other", "one"); err != nil || found {
		t.Fatal("cross case revision")
	}
	history, err := store.ListEvaluationCaseRevisions(ctx, "home", "case", ports.EvaluationCaseRevisionPageRequest{Limit: 1})
	if err != nil || len(history) != 1 || history[0].Snapshot().Number != 1 {
		t.Fatalf("first history: %v %v", history, err)
	}
	history, err = store.ListEvaluationCaseRevisions(ctx, "home", "case", ports.EvaluationCaseRevisionPageRequest{AfterNumber: 1, Limit: 100})
	if err != nil || len(history) != 1 || history[0].Snapshot().Number != 2 {
		t.Fatalf("exclusive history: %v %v", history, err)
	}
	for _, scope := range []struct {
		tenant tenant.ID
		caseID domain.EvaluationCaseID
	}{{"other", "case"}, {"home", "missing"}} {
		rows, err := store.ListEvaluationCaseRevisions(ctx, scope.tenant, scope.caseID, ports.EvaluationCaseRevisionPageRequest{Limit: 100})
		if err != nil || len(rows) != 0 {
			t.Fatalf("history scope leaked: %v %v", rows, err)
		}
	}
	for _, page := range []ports.EvaluationCaseRevisionPageRequest{{Limit: 0}, {Limit: 101}, {Limit: 1, AfterNumber: -1}} {
		if _, err := store.ListEvaluationCaseRevisions(ctx, "home", "case", page); !errors.Is(err, ports.ErrInvalidEvaluationCasePage) {
			t.Fatalf("invalid history page: %v", err)
		}
	}
	rows, err := store.ListEvaluationCases(ctx, "home", ports.EvaluationCasePageRequest{Limit: 1})
	if err != nil || len(rows) != 1 || rows[0].LatestRevision != 2 {
		t.Fatalf("head list: %+v %v", rows, err)
	}
	rows, err = store.ListEvaluationCases(ctx, "home", ports.EvaluationCasePageRequest{AfterID: "case", Limit: 1})
	if err != nil || len(rows) != 0 {
		t.Fatal("cursor repeated case")
	}
	for _, id := range []domain.EvaluationCaseID{"case-z", "case-a"} {
		snapshot := first.Snapshot()
		snapshot.CaseID = id
		snapshot.ID = domain.EvaluationCaseRevisionID("revision-" + string(id))
		value, err := domain.NewEvaluationCaseRevision(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.AppendEvaluationCaseRevision(ctx, value, 0, record("audit-"+string(id))); err != nil {
			t.Fatal(err)
		}
	}
	ids := []domain.EvaluationCaseID{}
	cursor := domain.EvaluationCaseID("")
	for range 4 {
		rows, err := store.ListEvaluationCases(ctx, "home", ports.EvaluationCasePageRequest{AfterID: cursor, Limit: 1})
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) == 0 {
			break
		}
		if len(rows) != 1 {
			t.Fatal("unbounded page")
		}
		ids = append(ids, rows[0].ID)
		cursor = rows[0].ID
	}
	if !slices.Equal(ids, []domain.EvaluationCaseID{"case", "case-a", "case-z"}) {
		t.Fatalf("pagination: %v", ids)
	}
	for _, limit := range []int{0, 101} {
		if _, err := store.ListEvaluationCases(ctx, "home", ports.EvaluationCasePageRequest{Limit: limit}); !errors.Is(err, ports.ErrInvalidEvaluationCasePage) {
			t.Fatalf("invalid limit %d: %v", limit, err)
		}
	}

}

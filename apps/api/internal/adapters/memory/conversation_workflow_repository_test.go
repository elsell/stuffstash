package memory

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sync"
	"testing"
	"time"
)

func TestWorkflowMemoryAtomicScopeAndConflict(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	name, _ := tenant.NewName("Home")
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: "home", Name: name}); err != nil {
		t.Fatal(err)
	}
	first := persistedWorkflowRevision(t, "home", "one", 1)
	record := func(id string, action audit.Action) audit.Record {
		value, ok := audit.NewRecord(audit.ID(id), "home", "", "owner", action, audit.SourceAPI, "conversation_workflow", "workflow-one", first.Snapshot().CreatedAt, "", nil)
		if !ok {
			t.Fatal("invalid fixture audit")
		}
		return value
	}
	if err := store.AppendWorkflowRevision(ctx, first, 0, record("audit-one", audit.ActionConversationWorkflowRevisionCreated)); err != nil {
		t.Fatal(err)
	}
	if _, found, err := store.WorkflowRevision(ctx, "other", "workflow-one", "one"); err != nil || found {
		t.Fatalf("scope leak: %v %v", found, err)
	}
	if err := store.AppendWorkflowRevision(ctx, persistedWorkflowRevision(t, "home", "two", 2), 1, record("audit-one", audit.ActionConversationWorkflowRevisionCreated)); err == nil {
		t.Fatal("duplicate audit accepted")
	}
	head, _, _ := store.WorkflowHead(ctx, "home", "workflow-one")
	if head.LatestRevision != 1 {
		t.Fatal("failed append changed head")
	}
	revisions := []agentmodel.WorkflowRevision{persistedWorkflowRevision(t, "home", "two-a", 2), persistedWorkflowRevision(t, "home", "two-b", 2)}
	audits := []audit.Record{record("audit-a", audit.ActionConversationWorkflowRevisionCreated), record("audit-b", audit.ActionConversationWorkflowRevisionCreated)}
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := range revisions {
		wg.Add(1)
		go func(i int) { defer wg.Done(); results <- store.AppendWorkflowRevision(ctx, revisions[i], 1, audits[i]) }(i)
	}
	wg.Wait()
	close(results)
	wins, conflicts := 0, 0
	for err := range results {
		if err == nil {
			wins++
		} else if errors.Is(err, ports.ErrWorkflowConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("CAS: wins=%d conflicts=%d", wins, conflicts)
	}
	if err := store.ActivateWorkflowRevision(ctx, "home", "workflow-one", "one", ports.WorkflowSelectionReference{}, first.Snapshot().CreatedAt, record("active", audit.ActionConversationWorkflowActivated)); err != nil {
		t.Fatal(err)
	}
	if err := store.ActivateWorkflowRevision(ctx, "home", "workflow-one", "one", ports.WorkflowSelectionReference{}, first.Snapshot().CreatedAt, record("stale-active", audit.ActionConversationWorkflowActivated)); !errors.Is(err, ports.ErrWorkflowConflict) {
		t.Fatalf("stale activation: %v", err)
	}
	if err := store.ActivateWorkflowRevision(ctx, "other", "workflow-one", "one", ports.WorkflowSelectionReference{}, first.Snapshot().CreatedAt, record("cross", audit.ActionConversationWorkflowActivated)); err == nil {
		t.Fatal("cross-tenant audit accepted")
	}

	var winner agentmodel.WorkflowRevisionID
	for _, candidate := range revisions {
		id := candidate.Snapshot().ID
		if _, found, _ := store.WorkflowRevision(ctx, "home", "workflow-one", id); found {
			winner = id
		}
	}
	if winner == "" {
		t.Fatal("winning revision missing")
	}
	if err := store.ActivateWorkflowRevision(ctx, "home", "workflow-one", winner, ports.WorkflowSelectionReference{WorkflowID: "workflow-one", RevisionID: "one"}, first.Snapshot().CreatedAt, record("active", audit.ActionConversationWorkflowActivated)); err == nil {
		t.Fatal("duplicate activation audit accepted")
	}

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := store.AppendWorkflowRevision(cancelled, persistedWorkflowRevision(t, "home", "cancelled", 3), 2, record("cancelled-append", audit.ActionConversationWorkflowRevisionCreated)); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled append: %v", err)
	}
	if err := store.ActivateWorkflowRevision(cancelled, "home", "workflow-one", winner, ports.WorkflowSelectionReference{WorkflowID: "workflow-one", RevisionID: "one"}, first.Snapshot().CreatedAt, record("cancelled-active", audit.ActionConversationWorkflowActivated)); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled activation: %v", err)
	}
	if _, found, _ := store.WorkflowRevision(ctx, "home", "workflow-one", "cancelled"); found {
		t.Fatal("cancelled revision committed")
	}
	records, err := store.ListTenantAuditRecords(ctx, "home", ports.AuditRecordPageRequest{})
	if err != nil || len(records) != 3 {
		t.Fatalf("atomic audit count: %d %v", len(records), err)
	}
	original, found, err := store.WorkflowRevision(ctx, "home", "workflow-one", "one")
	if err != nil || !found || original.Snapshot().Number != 1 {
		t.Fatal("original revision lost")
	}
	head, _, _ = store.WorkflowHead(ctx, "home", "workflow-one")
	if head.LatestRevision != 2 || head.ActiveRevisionID != "one" {
		t.Fatalf("invalid head: %+v", head)
	}
}
func persistedWorkflowRevision(t *testing.T, tenantID, id string, number int) agentmodel.WorkflowRevision {
	t.Helper()
	limits := agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{ToolCalls: 4, ModelCalls: 20, ElapsedSeconds: 120, FollowUpTurns: 8}, MaxNameRunes: 120, MaxInstructionRunes: 4000}
	definition, err := agentmodel.NewWorkflowDefinition(agentmodel.WorkflowDefinitionInput{Name: "Home workflow", Budget: limits.Budget}, limits)
	if err != nil {
		t.Fatal(err)
	}
	revision, err := agentmodel.NewWorkflowRevision(agentmodel.WorkflowRevisionInput{ID: agentmodel.WorkflowRevisionID(id), WorkflowID: "workflow-one", TenantID: agentmodel.TenantID(tenantID), AuthorID: "owner-one", Number: number, Definition: definition, Limits: limits, CreatedAt: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	return revision
}

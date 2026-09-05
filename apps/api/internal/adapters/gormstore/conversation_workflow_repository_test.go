package gormstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestConversationWorkflowRepositorySnapshotsConflictsAndScope(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	verifyConversationWorkflowRepository(t, ctx, store)
}

func verifyConversationWorkflowRepository(t *testing.T, ctx context.Context, store Store) {
	t.Helper()
	home := tenant.ID("workflow-home")
	other := tenant.ID("workflow-other")
	saveTenant(t, ctx, store, home, "Home")
	saveTenant(t, ctx, store, other, "Other")
	first := persistedWorkflowRevision(t, string(home), "revision-one", 1)
	if err := store.AppendWorkflowRevision(ctx, first, 0, auditRecord(t, "workflow-audit-one", home, "", audit.ActionConversationWorkflowRevisionCreated)); err != nil {
		t.Fatal(err)
	}
	if _, found, err := store.WorkflowRevision(ctx, other, "workflow-one", "revision-one"); err != nil || found {
		t.Fatalf("cross-tenant revision: %v %v", found, err)
	}
	if _, found, err := store.WorkflowRevision(ctx, home, "workflow-other", "revision-one"); err != nil || found {
		t.Fatalf("cross-workflow revision: %v %v", found, err)
	}
	second := persistedWorkflowRevision(t, string(home), "revision-two", 2)
	if err := store.AppendWorkflowRevision(ctx, second, 0, auditRecord(t, "workflow-audit-stale", home, "", audit.ActionConversationWorkflowRevisionCreated)); !errors.Is(err, ports.ErrWorkflowConflict) {
		t.Fatalf("stale append must conflict: %v", err)
	}
	if err := store.AppendWorkflowRevision(ctx, second, 1, auditRecord(t, "workflow-audit-two", home, "", audit.ActionConversationWorkflowRevisionCreated)); err != nil {
		t.Fatal(err)
	}

	if err := store.AppendWorkflowRevision(ctx, persistedWorkflowRevision(t, string(home), "revision-stale-two", 2), 1, auditRecord(t, "workflow-real-stale", home, "", audit.ActionConversationWorkflowRevisionCreated)); !errors.Is(err, ports.ErrWorkflowConflict) {
		t.Fatalf("database head CAS did not conflict: %v", err)
	}
	if err := store.ActivateWorkflowRevision(ctx, home, "workflow-other", "revision-one", "", first.Snapshot().CreatedAt, auditRecord(t, "workflow-wrong-target", home, "", audit.ActionConversationWorkflowActivated)); !errors.Is(err, ports.ErrWorkflowNotFound) {
		t.Fatalf("cross-workflow activation: %v", err)
	}
	if err := store.ActivateWorkflowRevision(ctx, home, "workflow-one", "revision-one", "", first.Snapshot().CreatedAt, auditRecord(t, "workflow-wrong-audit", other, "", audit.ActionConversationWorkflowActivated)); err == nil {
		t.Fatal("cross-tenant activation audit accepted")
	}
	if err := store.ActivateWorkflowRevision(ctx, home, "workflow-one", "revision-one", "", first.Snapshot().CreatedAt, auditRecord(t, "workflow-active-one", home, "", audit.ActionConversationWorkflowActivated)); err != nil {
		t.Fatal(err)
	}
	if err := store.ActivateWorkflowRevision(ctx, home, "workflow-one", "revision-two", "", first.Snapshot().CreatedAt, auditRecord(t, "workflow-active-stale", home, "", audit.ActionConversationWorkflowActivated)); !errors.Is(err, ports.ErrWorkflowConflict) {
		t.Fatalf("stale activation must conflict: %v", err)
	}

	if err := store.ActivateWorkflowRevision(ctx, other, "workflow-one", "revision-one", "", first.Snapshot().CreatedAt, auditRecord(t, "workflow-cross-active", other, "", audit.ActionConversationWorkflowActivated)); !errors.Is(err, ports.ErrWorkflowNotFound) {
		t.Fatalf("cross-tenant activation: %v", err)
	}
	if err := store.ActivateWorkflowRevision(ctx, home, "workflow-one", "revision-two", "revision-one", first.Snapshot().CreatedAt, auditRecord(t, "workflow-active-one", home, "", audit.ActionConversationWorkflowActivated)); err == nil {
		t.Fatal("duplicate activation audit accepted")
	}
	if err := store.AppendWorkflowRevision(ctx, persistedWorkflowRevision(t, string(home), "revision-cross-audit", 3), 2, auditRecord(t, "workflow-cross-audit", other, "", audit.ActionConversationWorkflowRevisionCreated)); err == nil {
		t.Fatal("cross-tenant audit accepted")
	}
	head, found, err := store.WorkflowHead(ctx, home, "workflow-one")
	if err != nil || !found || head.LatestRevision != 2 || head.ActiveRevisionID != "revision-one" {
		t.Fatalf("bad workflow head: %+v %v", head, err)
	}
	got, found, err := store.WorkflowRevision(ctx, home, "workflow-one", "revision-one")
	if err != nil || !found || got.Snapshot().Number != 1 || got.Snapshot().Definition.Settings().Name != "Home workflow" {
		t.Fatalf("history changed: %+v %v", got, err)
	}
	// Reusing an audit ID must roll back the revision and head update together.
	third := persistedWorkflowRevision(t, string(home), "revision-three", 3)
	if err := store.AppendWorkflowRevision(ctx, third, 2, auditRecord(t, "workflow-audit-two", home, "", audit.ActionConversationWorkflowRevisionCreated)); err == nil {
		t.Fatal("duplicate audit accepted")
	}
	head, _, err = store.WorkflowHead(ctx, home, "workflow-one")
	if err != nil || head.LatestRevision != 2 {
		t.Fatalf("audit failure changed head: %+v %v", head, err)
	}
	if _, found, err := store.WorkflowRevision(ctx, home, "workflow-one", "revision-three"); err != nil || found {
		t.Fatalf("audit failure persisted revision: %v %v", found, err)
	}
}

func persistedWorkflowRevision(t *testing.T, tenantID, id string, number int) agentmodel.WorkflowRevision {
	t.Helper()
	limits := agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{EvidenceRounds: 4, ModelCalls: 20, ElapsedSeconds: 120, FollowUpTurns: 8}, MaxStepAttempts: 3, MaxNameRunes: 120, MaxInstructionRunes: 4000}
	definition, err := agentmodel.NewWorkflowDefinition(agentmodel.WorkflowDefinitionInput{Name: "Home workflow", Retrieval: agentmodel.WorkflowRetrievalPreciseFirst, Response: agentmodel.WorkflowResponseGroundedFallback, Budget: limits.Budget, Steps: []agentmodel.WorkflowStep{{Kind: agentmodel.WorkflowStepInterpret, Attempts: 1}, {Kind: agentmodel.WorkflowStepAssess, Attempts: 1}, {Kind: agentmodel.WorkflowStepRespond, Attempts: 1}}}, limits)
	if err != nil {
		t.Fatal(err)
	}
	revision, err := agentmodel.NewWorkflowRevision(agentmodel.WorkflowRevisionInput{ID: agentmodel.WorkflowRevisionID(id), WorkflowID: "workflow-one", TenantID: agentmodel.TenantID(tenantID), AuthorID: "owner-one", Number: number, Definition: definition, Limits: limits, CreatedAt: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	return revision
}

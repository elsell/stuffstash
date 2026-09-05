package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"os"
	"testing"
	"time"
)

func TestPostgresConversationWorkflowProductionMigrations(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		for _, model := range []any{&conversationWorkflowSelectionModel{}, &conversationWorkflowRevisionModel{}, &conversationWorkflowModel{}, &auditRecordModel{}, &tenantModel{}} {
			column := "tenant_id"
			if _, ok := model.(*tenantModel); ok {
				column = "id"
			}
			if err := db.Where(column+" IN ?", []string{"workflow-home", "workflow-other"}).Delete(model).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	verifyConversationWorkflowRepository(t, context.Background(), NewStore(db))
	store := NewStore(db)
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, id := range []string{"concurrent-one", "concurrent-two"} {
		revision := persistedWorkflowRevision(t, "workflow-home", id, 3)
		record := auditRecord(t, "audit-"+id, tenant.ID("workflow-home"), "", audit.ActionConversationWorkflowRevisionCreated, "conversation_workflow")
		go func() { <-start; results <- store.AppendWorkflowRevision(context.Background(), revision, 2, record) }()
	}
	close(start)
	wins, conflicts := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			wins++
		} else if errors.Is(err, ports.ErrWorkflowConflict) {
			conflicts++
		} else {
			t.Fatalf("unexpected concurrent result: %v", err)
		}
	}

	var revisions, audits int64
	if err := db.Model(&conversationWorkflowRevisionModel{}).Where(map[string]any{"tenant_id": "workflow-home", "workflow_id": "workflow-one", "number": 3}).Count(&revisions).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&auditRecordModel{}).Where("id IN ?", []string{"audit-concurrent-one", "audit-concurrent-two"}).Count(&audits).Error; err != nil {
		t.Fatal(err)
	}
	if revisions != 1 || audits != 1 {
		t.Fatalf("concurrent save leaked state: revisions=%d audits=%d", revisions, audits)
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("expected one append winner and one conflict: %d %d", wins, conflicts)
	}

	activationStart := make(chan struct{})
	activationResults := make(chan error, 2)
	expected := ports.WorkflowSelectionReference{WorkflowID: "workflow-other", RevisionID: "other-revision"}
	for _, target := range []ports.WorkflowSelectionReference{{WorkflowID: "workflow-one", RevisionID: "revision-one"}, {WorkflowID: "workflow-other", RevisionID: "other-draft"}} {
		record := auditRecord(t, "activation-"+string(target.WorkflowID), tenant.ID("workflow-home"), "", audit.ActionConversationWorkflowActivated, "conversation_workflow")
		go func() {
			<-activationStart
			activationResults <- store.ActivateWorkflowRevision(context.Background(), "workflow-home", target.WorkflowID, target.RevisionID, expected, time.Date(2026, 9, 5, 13, 0, 0, 0, time.UTC), record)
		}()
	}
	close(activationStart)
	wins, conflicts = 0, 0
	for range 2 {
		err := <-activationResults
		if err == nil {
			wins++
		} else if errors.Is(err, ports.ErrWorkflowConflict) {
			conflicts++
		} else {
			t.Errorf("activation failed unexpectedly: %v", err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("cross-workflow activation must have one winner: %d %d", wins, conflicts)
	}
	if err := db.Model(&auditRecordModel{}).Where("id IN ?", []string{"activation-workflow-one", "activation-workflow-other"}).Count(&audits).Error; err != nil {
		t.Fatal(err)
	}
	if audits != 1 {
		t.Fatalf("losing activation wrote audit: %d", audits)
	}

}

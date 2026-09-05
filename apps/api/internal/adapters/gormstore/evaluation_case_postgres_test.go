package gormstore

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm/clause"
)

func TestPostgresEvaluationCaseMigrationsAndConcurrentSaves(t *testing.T) {
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
		for _, model := range []any{&evaluationCaseRevisionModel{}, &evaluationCaseModel{}, &auditRecordModel{}, &tenantModel{}} {
			column := "tenant_id"
			if _, ok := model.(*tenantModel); ok {
				column = "id"
			}
			if err := db.Where(clause.Eq{Column: column, Value: "evaluation-home"}).Delete(model).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	ctx := context.Background()
	store := NewStore(db)
	verifyEvaluationCaseRepository(t, ctx, store)
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, id := range []string{"candidate-a", "candidate-b"} {
		revision := storedEvaluationRevision(t, id, 3)
		record := auditRecord(t, id, "evaluation-home", "", audit.ActionConversationEvaluationCaseRevisionCreated, "conversation_evaluation_case")
		go func() { <-start; results <- store.AppendEvaluationCaseRevision(ctx, revision, 2, record) }()
	}
	close(start)
	wins, conflicts := 0, 0
	for range 2 {
		err := <-results
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
	var revisions, audits int64
	if err := db.Model(&evaluationCaseRevisionModel{}).Where(map[string]any{"tenant_id": "evaluation-home", "case_id": "clothes", "number": 3}).Count(&revisions).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&auditRecordModel{}).Where(clause.IN{Column: "id", Values: []any{"candidate-a", "candidate-b"}}).Count(&audits).Error; err != nil {
		t.Fatal(err)
	}
	if revisions != 1 || audits != 1 {
		t.Fatalf("losing write left state: revisions=%d audits=%d", revisions, audits)
	}
}

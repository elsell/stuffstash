package workflowdiscovery

import (
	"context"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	fixture "github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
	"testing"
)

func Verify(t *testing.T, repo interface {
	ports.ConversationWorkflowRepository
	ports.WorkflowDiscoveryRepository
}) {
	t.Helper()
	ctx := context.Background()
	for _, id := range []string{"b", "a"} {
		for number := 1; number <= 2; number++ {
			input := fixture.Run(t, "template").Snapshot().Input.Workflow.Snapshot()
			input.WorkflowID = model.WorkflowID(id)
			input.TenantID = "discovery-home"
			input.Number = number
			suffix := "first"
			if number == 2 {
				suffix = "second"
			}
			input.ID = model.WorkflowRevisionID(id + "-" + suffix)
			definition := input.Definition.Settings()
			definition.Name = id + " " + suffix
			var err error
			input.Definition, err = model.NewWorkflowDefinition(definition, input.Limits)
			if err != nil {
				t.Fatal(err)
			}
			revision, err := model.NewWorkflowRevision(input)
			if err != nil {
				t.Fatal(err)
			}
			record, ok := audit.NewRecord(audit.ID(input.ID), "discovery-home", "", "owner", audit.ActionConversationWorkflowRevisionCreated, audit.SourceAPI, audit.TargetConversationWorkflow, id, input.CreatedAt, "", nil)
			if !ok {
				t.Fatal("invalid audit")
			}
			if err := repo.AppendWorkflowRevision(ctx, revision, number-1, record); err != nil {
				t.Fatal(err)
			}
		}
	}
	rows, err := repo.ListWorkflowHeads(ctx, "discovery-home", ports.WorkflowHeadPageRequest{Limit: 1})
	if err != nil || len(rows) != 1 || rows[0].ID != "a" || rows[0].Name != "a second" || rows[0].LatestRevisionID != "a-second" || rows[0].LatestRevision != 2 {
		t.Fatalf("first head: %+v %v", rows, err)
	}
	next, err := repo.ListWorkflowHeads(ctx, "discovery-home", ports.WorkflowHeadPageRequest{Limit: 1, AfterID: rows[0].ID})
	if err != nil || len(next) != 1 || next[0].ID != "b" {
		t.Fatalf("head cursor: %+v %v", next, err)
	}
	for _, scope := range []tenant.ID{"discovery-home", "other"} {
		revisions, err := repo.ListWorkflowRevisions(ctx, scope, "a", ports.WorkflowRevisionPageRequest{Limit: 1})
		if err != nil {
			t.Fatal(err)
		}
		if scope == "other" {
			if len(revisions) != 0 {
				t.Fatal("cross-tenant revisions exposed")
			}
			continue
		}
		if len(revisions) != 1 || revisions[0].Snapshot().ID != "a-first" {
			t.Fatal("revision ordering")
		}
		revisions, err = repo.ListWorkflowRevisions(ctx, scope, "a", ports.WorkflowRevisionPageRequest{Limit: 1, AfterNumber: 1})
		if err != nil || len(revisions) != 1 || revisions[0].Snapshot().ID != "a-second" {
			t.Fatalf("revision cursor: %v", err)
		}
	}

	empty, err := repo.ListWorkflowRevisions(ctx, "discovery-home", "missing", ports.WorkflowRevisionPageRequest{Limit: 1})
	if err != nil || len(empty) != 0 {
		t.Fatal("unknown workflow returned history")
	}
	if _, err := repo.ListWorkflowRevisions(ctx, "discovery-home", "a", ports.WorkflowRevisionPageRequest{Limit: 1, AfterNumber: -1}); err == nil {
		t.Fatal("negative revision cursor accepted")
	}
	exhausted, err := repo.ListWorkflowHeads(ctx, "discovery-home", ports.WorkflowHeadPageRequest{Limit: 1, AfterID: "b"})
	if err != nil || len(exhausted) != 0 {
		t.Fatal("exhausted page returned rows")
	}
	foreign, err := repo.ListWorkflowHeads(ctx, "other", ports.WorkflowHeadPageRequest{Limit: 100})
	if err != nil || len(foreign) != 0 {
		t.Fatal("cross-tenant heads exposed")
	}
	for _, limit := range []int{0, -1, 101} {
		if _, err := repo.ListWorkflowHeads(ctx, "discovery-home", ports.WorkflowHeadPageRequest{Limit: limit}); err == nil {
			t.Fatal("invalid head limit")
		}
		if _, err := repo.ListWorkflowRevisions(ctx, "discovery-home", "a", ports.WorkflowRevisionPageRequest{Limit: limit}); err == nil {
			t.Fatal("invalid revision limit")
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repo.ListWorkflowHeads(cancelled, "discovery-home", ports.WorkflowHeadPageRequest{Limit: 1}); err == nil {
		t.Fatal("cancelled head read")
	}
	if _, err := repo.ListWorkflowRevisions(cancelled, "discovery-home", "a", ports.WorkflowRevisionPageRequest{Limit: 1}); err == nil {
		t.Fatal("cancelled revision read")
	}
}

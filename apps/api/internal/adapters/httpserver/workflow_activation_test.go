package httpserver

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"net/http"
	"testing"
	"time"
)

func TestWorkflowActivationHTTPAuthorizationAndEvidence(t *testing.T) {
	application, store := newWorkflowHTTPTestRuntime(t)
	server := NewServer(":0", application)
	queue := evaluationRunRequest(t, server)
	queued := performRequest(server, http.MethodPost, "/tenants/home/conversation-evaluation-runs", "Bearer dev:owner", queue)
	if queued.Code != 201 {
		t.Fatalf("queue: %s", queued.Body.String())
	}
	var value struct{ Data struct{ ID string } }
	decodeBody(t, queued, &value)
	body := map[string]any{"revisionId": queue["revisionId"], "runId": value.Data.ID, "cases": queue["cases"]}
	path := "/tenants/home/conversation-workflows/" + queue["workflowId"].(string) + "/activation"
	for _, scenario := range []struct {
		token  string
		status int
	}{{"", 401}, {"Bearer malformed", 401}, {"Bearer dev:stranger", 403}, {"Bearer dev:outsider", 403}, {"Bearer dev:inventory-owner", 403}, {"Bearer dev:viewer", 403}} {
		response := performRequest(server, http.MethodPost, path, scenario.token, body)
		if response.Code != scenario.status {
			t.Fatalf("authorization %q: %d %s", scenario.token, response.Code, response.Body.String())
		}
	}
	pending := performRequest(server, http.MethodPost, path, "Bearer dev:owner", body)
	if pending.Code != 412 {
		t.Fatalf("pending evidence: %d %s", pending.Code, pending.Body.String())
	}
	completeHTTPWorkflowRun(t, store, value.Data.ID)
	foreign := performRequest(server, http.MethodPost, "/tenants/other-home/conversation-workflows/"+queue["workflowId"].(string)+"/activation", "Bearer dev:outsider", body)
	if foreign.Code != 404 {
		t.Fatalf("foreign workflow: %d %s", foreign.Code, foreign.Body.String())
	}
	response := performRequest(server, http.MethodPost, path, "Bearer dev:owner", body)
	if response.Code != 200 {
		t.Fatalf("activation: %d %s", response.Code, response.Body.String())
	}
	var selected struct{ Data struct{ ID string } }
	decodeBody(t, response, &selected)
	if selected.Data.ID != queue["revisionId"] {
		t.Fatal("wrong revision selected")
	}
	stale := performRequest(server, http.MethodPost, path, "Bearer dev:owner", body)
	if stale.Code != 409 {
		t.Fatalf("stale selection: %d %s", stale.Code, stale.Body.String())
	}
}

// Supply successful controlled outcomes through real domain transitions and
// repository CAS, retaining the distinction from live-provider quality evidence.
func completeHTTPWorkflowRun(t *testing.T, store *gormstore.Store, id string) {
	t.Helper()
	ctx := context.Background()
	run, found, err := store.EvaluationRun(ctx, "home", model.EvaluationRunID(id))
	if err != nil || !found {
		t.Fatalf("load run: %v", err)
	}
	ids := idgen.NewULIDGenerator()
	save := func(next model.EvaluationRun) {
		record, ok := audit.NewRecord(audit.ID(ids.NewID()), "home", "", "owner", audit.ActionConversationEvaluationRunProgressed, audit.SourceSystem, audit.TargetConversationEvaluationRun, id, time.Now(), "", nil)
		if !ok {
			t.Fatal("audit fixture invalid")
		}
		if err := store.SaveEvaluationRun(ctx, next, run.Snapshot().Version, record); err != nil {
			t.Fatal(err)
		}
		run = next
	}
	claimed, err := run.Claim("test-lease", time.Now(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	save(claimed)
	for i, revision := range run.Snapshot().Input.Cases {
		expected := revision.Snapshot().Definition.Settings().Expectations
		next, err := run.RecordCase("test-lease", i, model.EvaluationObservedOutcome{Kind: expected.Kind, ReferencedAssets: expected.ReferencedAssets, Locations: expected.Locations, Proposals: expected.Proposals}, 1, 0, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		save(next)
	}
}

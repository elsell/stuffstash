package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
)

func TestWorkflowActivationRejectsHistoricalRuntimeEvidence(t *testing.T) {
	application, store := newWorkflowHTTPTestRuntime(t)
	server := NewServer(":0", application)
	request := evaluationRunRequest(t, server)
	queued := performRequest(server, http.MethodPost, "/tenants/home/conversation-evaluation-runs", "Bearer dev:owner", request)
	if queued.Code != http.StatusCreated {
		t.Fatalf("queue: %s", queued.Body.String())
	}
	var value struct{ Data struct{ ID string } }
	decodeBody(t, queued, &value)
	current, found, err := store.EvaluationRun(context.Background(), "home", model.EvaluationRunID(value.Data.ID))
	if err != nil || !found {
		t.Fatalf("load current run: %v", err)
	}
	// Reproduce an older persisted run without referring to the new field at
	// compile time, so this test fails against the pre-migration runtime.
	raw, err := json.Marshal(current.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	var input map[string]json.RawMessage
	if err := json.Unmarshal(document["Input"], &input); err != nil {
		t.Fatal(err)
	}
	delete(input, "RuntimeContract")
	input["ID"] = json.RawMessage(`"historical-runtime-run"`)
	document["Input"], err = json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot model.EvaluationRunSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	// Domain revisions encapsulate private state and are persisted by adapter mappers.
	snapshot.Input.Workflow = current.Snapshot().Input.Workflow
	snapshot.Input.Cases = current.Snapshot().Input.Cases
	historical, err := model.RestoreEvaluationRun(snapshot)
	if err != nil {
		t.Fatalf("historical evidence must remain readable: %v", err)
	}
	record, ok := audit.NewRecord("historical-runtime-audit", "home", "", "owner", audit.ActionConversationEvaluationRunCreated, audit.SourceSystem, audit.TargetConversationEvaluationRun, "historical-runtime-run", time.Now(), "", nil)
	if !ok {
		t.Fatal("invalid audit fixture")
	}
	if err := store.SaveEvaluationRun(context.Background(), historical, 0, record); err != nil {
		t.Fatal(err)
	}
	completeHTTPWorkflowRun(t, store, "historical-runtime-run")
	path := "/tenants/home/conversation-workflows/" + request["workflowId"].(string) + "/activation"
	body := map[string]any{"revisionId": request["revisionId"], "runId": "historical-runtime-run", "cases": request["cases"]}
	response := performRequest(server, http.MethodPost, path, "Bearer dev:owner", body)
	if response.Code != http.StatusPreconditionFailed {
		t.Fatalf("historical runtime evidence activated: %d %s", response.Code, response.Body.String())
	}
	// A fresh run over exactly the same workflow/provider/case pins remains valid.
	completeHTTPWorkflowRun(t, store, value.Data.ID)
	body["runId"] = value.Data.ID
	response = performRequest(server, http.MethodPost, path, "Bearer dev:owner", body)
	if response.Code != http.StatusOK {
		t.Fatalf("current runtime evidence rejected: %d %s", response.Code, response.Body.String())
	}
}

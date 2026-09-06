package httpserver

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"net/http"
	"strings"
	"testing"
)

// Controlled, non-network snapshot resolver: endpoint tests exercise the real
// authorization, revisions and run repositories without claiming provider quality.
type evaluationHTTPSnapshotResolver struct{}

func (evaluationHTTPSnapshotResolver) SnapshotEvaluationProviders(_ context.Context, scope tenant.ID, revision agentmodel.WorkflowRevision) ([]agentmodel.EvaluationRunProvider, error) {
	definition := revision.Snapshot().Definition.Settings()
	profile := agentmodel.ProviderProfileID(definition.ProviderProfileID)
	if profile == "" {
		profile = agentmodel.ProviderProfileID(scope.String() + "-language")
	}
	return []agentmodel.EvaluationRunProvider{{ProfileID: profile, ConfigurationID: strings.Repeat("a", 64)}}, nil
}
func evaluationRunRequest(t *testing.T, server *http.Server) map[string]any {
	t.Helper()
	workflow := workflowDraftRequest()
	delete(workflow, "expectedRevision")
	created := performRequest(server, http.MethodPost, "/tenants/home/conversation-workflows", "Bearer dev:owner", workflow)
	if created.Code != 201 {
		t.Fatalf("create workflow: %s", created.Body.String())
	}
	var pinnedWorkflow struct {
		Data struct{ ID, WorkflowID string }
	}
	decodeBody(t, created, &pinnedWorkflow)
	saved := performRequest(server, http.MethodPost, "/tenants/home/conversation-evaluation-cases", "Bearer dev:owner", evaluationCaseRequest())
	if saved.Code != 201 {
		t.Fatalf("create case: %s", saved.Body.String())
	}
	var pinnedCase struct{ Data struct{ ID, CaseID string } }
	decodeBody(t, saved, &pinnedCase)
	return map[string]any{"workflowId": pinnedWorkflow.Data.WorkflowID, "revisionId": pinnedWorkflow.Data.ID, "cases": []map[string]string{{"caseId": pinnedCase.Data.CaseID, "revisionId": pinnedCase.Data.ID}}}
}

func coverEvaluationRunScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	const collection = "/tenants/{tenantId}/conversation-evaluation-runs"
	const head = collection + "/{runId}"
	const cancellation = head + "/cancellation"
	const base = "/tenants/home/conversation-evaluation-runs"
	body := evaluationRunRequest(t, server)
	if adversarial {
		coverage.request(t, server, http.MethodPost, collection, base, "Bearer dev:viewer", body, http.StatusForbidden)
		coverage.request(t, server, http.MethodGet, collection, base, "Bearer dev:viewer", nil, http.StatusForbidden)
		coverage.request(t, server, http.MethodGet, head, base+"/unknown", "Bearer dev:viewer", nil, http.StatusForbidden)
		coverage.request(t, server, http.MethodPost, cancellation, base+"/unknown/cancellation", "Bearer dev:viewer", map[string]int{"expectedVersion": 1}, http.StatusForbidden)
		return
	}
	created := coverage.request(t, server, http.MethodPost, collection, base, "Bearer dev:owner", body, http.StatusCreated)
	var value struct{ Data struct{ ID string } }
	decodeBody(t, created, &value)
	coverage.request(t, server, http.MethodGet, collection, base, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, head, base+"/"+value.Data.ID, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPost, cancellation, base+"/"+value.Data.ID+"/cancellation", "Bearer dev:owner", map[string]int{"expectedVersion": 1}, http.StatusOK)
}

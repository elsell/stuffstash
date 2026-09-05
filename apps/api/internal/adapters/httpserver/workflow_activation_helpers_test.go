package httpserver

import (
	"net/http"
	"testing"
)

func coverWorkflowActivationScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	application, store := newWorkflowHTTPTestRuntime(t)
	server := NewServer(":0", application)
	queue := evaluationRunRequest(t, server)
	queued := performRequest(server, http.MethodPost, "/tenants/home/conversation-evaluation-runs", "Bearer dev:owner", queue)
	if queued.Code != http.StatusCreated {
		t.Fatalf("queue fixture: %s", queued.Body.String())
	}
	var value struct{ Data struct{ ID string } }
	decodeBody(t, queued, &value)
	body := map[string]any{"revisionId": queue["revisionId"], "runId": value.Data.ID, "cases": queue["cases"]}
	path := "/tenants/home/conversation-workflows/" + queue["workflowId"].(string) + "/activation"
	const template = "/tenants/{tenantId}/conversation-workflows/{workflowId}/activation"
	if adversarial {
		coverage.request(t, server, http.MethodPost, template, path, "Bearer dev:viewer", body, http.StatusForbidden)
		return
	}
	completeHTTPWorkflowRun(t, store, value.Data.ID)
	coverage.request(t, server, http.MethodPost, template, path, "Bearer dev:owner", body, http.StatusOK)
}

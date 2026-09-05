package httpserver

import (
	"net/http"
	"testing"
)

func coverWorkflowReadScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	queue := evaluationRunRequest(t, server)
	const base = "/tenants/{tenantId}/conversation-workflows"
	path := "/tenants/home/conversation-workflows"
	id := queue["workflowId"].(string)
	revision := queue["revisionId"].(string)
	token, status := "Bearer dev:owner", http.StatusOK
	if adversarial {
		token, status = "Bearer dev:viewer", http.StatusForbidden
	}
	for _, entry := range []struct{ template, path string }{
		{base, path}, {base + "/{workflowId}", path + "/" + id}, {base + "/{workflowId}/revisions", path + "/" + id + "/revisions"}, {base + "/{workflowId}/revisions/{revisionId}", path + "/" + id + "/revisions/" + revision}, {"/tenants/{tenantId}/conversation-workflow-selection", "/tenants/home/conversation-workflow-selection"},
	} {
		coverage.request(t, server, http.MethodGet, entry.template, entry.path, token, nil, status)
	}
}

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
	values := []agentmodel.EvaluationRunProvider{}
	for _, step := range definition.Steps {
		if step.Kind == agentmodel.WorkflowStepRespond && definition.Response == agentmodel.WorkflowResponseGrounded {
			continue
		}
		profile := agentmodel.ProviderProfileID(step.ProviderProfileID)
		if profile == "" {
			profile = agentmodel.ProviderProfileID(scope.String() + "-language")
		}
		values = append(values, agentmodel.EvaluationRunProvider{Step: step.Kind, ProfileID: profile, ConfigurationID: strings.Repeat("a", 64)})
	}
	return values, nil
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

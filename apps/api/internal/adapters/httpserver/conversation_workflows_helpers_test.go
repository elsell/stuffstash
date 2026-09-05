package httpserver

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	modelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/http"
	"path/filepath"
	"testing"
)

func newWorkflowHTTPTestApp(t *testing.T) app.App {
	application, _ := newWorkflowHTTPTestRuntime(t)
	return application
}

func newWorkflowHTTPTestRuntime(t *testing.T) (app.App, gormstore.Store) {
	t.Helper()
	ctx := context.Background()
	db, err := gormstore.OpenSQLite(filepath.Join(t.TempDir(), "workflow.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if err := gormstore.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	repository := gormstore.NewStore(db)
	name, _ := tenant.NewName("Home")
	if err := repository.SaveTenant(ctx, tenant.Tenant{ID: "home", Name: name}); err != nil {
		t.Fatal(err)
	}
	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	seedMemoryStore(t, ctx, store, authorizer, seededState{tenants: []seedTenant{{id: "home", name: "Home", owner: "owner"}, {id: "other-home", name: "Other", owner: "outsider"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "home", name: "Things", owner: "inventory-owner"}}})
	if err := authorizer.GrantInventoryViewer(ctx, principal("viewer"), tenant.ID("home"), inventory.InventoryID("inventory-home")); err != nil {
		t.Fatal(err)
	}
	limits := agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{EvidenceRounds: 5, ModelCalls: 10, ElapsedSeconds: 120, FollowUpTurns: 5}, MaxStepAttempts: 3, MaxNameRunes: 100, MaxInstructionRunes: 2000}
	ids := idgen.NewULIDGenerator()
	commands := modelapp.NewEvaluationRunCommandService(modelapp.EvaluationRunCommandDependencies{Authorizer: authorizer, Runs: repository, Workflows: repository, Cases: repository, Providers: evaluationHTTPSnapshotResolver{}, IDs: ids, Clock: ports.SystemClock{}, Limits: limits, MaxAttempts: 2})
	queries := modelapp.NewEvaluationRunQueryService(modelapp.EvaluationRunQueryDependencies{Authorizer: authorizer, Runs: repository, Audit: repository, IDs: ids, Clock: ports.SystemClock{}})
	activation := modelapp.NewWorkflowActivationService(modelapp.WorkflowActivationDependencies{Authorizer: authorizer, Workflows: repository, Runs: repository, Providers: evaluationHTTPSnapshotResolver{}, IDs: ids, Clock: ports.SystemClock{}, Limits: limits})
	return app.New(app.Dependencies{WorkflowActivation: activation, EvaluationRunCommands: commands, EvaluationRunQueries: queries, Auth: auth.NewLocalDevAuthenticator(), Authorizer: authorizer, Users: store, Tenants: store, ProviderProfiles: store, ConversationWorkflows: repository, EvaluationCases: repository, Audit: repository, ConversationWorkflowLimits: agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{EvidenceRounds: 5, ModelCalls: 10, ElapsedSeconds: 120, FollowUpTurns: 5}, MaxStepAttempts: 3, MaxNameRunes: 100, MaxInstructionRunes: 2000}}), repository
}

func workflowDraftRequest() map[string]any {
	return map[string]any{
		"expectedRevision": 1,
		"definition": map[string]any{
			"name": "Home voice", "retrieval": "expanded", "response": "grounded",
			"budget": map[string]any{"evidenceRounds": 2, "modelCalls": 4, "elapsedSeconds": 30, "followUpTurns": 2},
			"steps": []map[string]any{
				{"kind": "interpret", "attempts": 1, "instructions": "Resolve existing items first."},
				{"kind": "assess", "attempts": 1},
				{"kind": "respond", "attempts": 1},
			},
		},
	}
}

func coverWorkflowDraftScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	createBody := workflowDraftRequest()
	delete(createBody, "expectedRevision")
	const createTemplate = "/tenants/{tenantId}/conversation-workflows"
	const appendTemplate = "/tenants/{tenantId}/conversation-workflows/{workflowId}/revisions"
	if adversarial {
		coverage.request(t, server, http.MethodPost, createTemplate, "/tenants/home/conversation-workflows", "Bearer dev:viewer", createBody, http.StatusForbidden)
		coverage.request(t, server, http.MethodPost, appendTemplate, "/tenants/home/conversation-workflows/unknown/revisions", "Bearer dev:viewer", workflowDraftRequest(), http.StatusForbidden)
		return
	}
	created := coverage.request(t, server, http.MethodPost, createTemplate, "/tenants/home/conversation-workflows", "Bearer dev:owner", createBody, http.StatusCreated)
	var revision struct {
		Data struct {
			WorkflowID string `json:"workflowId"`
		} `json:"data"`
	}
	decodeBody(t, created, &revision)
	coverage.request(t, server, http.MethodPost, appendTemplate, "/tenants/home/conversation-workflows/"+revision.Data.WorkflowID+"/revisions", "Bearer dev:owner", workflowDraftRequest(), http.StatusCreated)
}

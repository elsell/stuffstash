package httpserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestConversationWorkflowDraftRoutesRequireTenantConfiguration(t *testing.T) {
	const home = "home"
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	for _, path := range []string{"/tenants/" + home + "/conversation-workflows", "/tenants/" + home + "/conversation-workflows/workflow-one/revisions"} {
		for _, tc := range []struct {
			name, token string
			status      int
		}{
			{"anonymous", "", http.StatusUnauthorized},
			{"invalid token", "Bearer malformed", http.StatusUnauthorized},
			{"other tenant owner", "Bearer dev:outsider", http.StatusForbidden},
			{"nonmember", "Bearer dev:stranger", http.StatusForbidden},
			{"inventory owner", "Bearer dev:inventory-owner", http.StatusForbidden},
			{"tenant member viewer", "Bearer dev:viewer", http.StatusForbidden},
		} {
			t.Run(path+"/"+tc.name, func(t *testing.T) {
				body := workflowDraftRequest()
				if !strings.HasSuffix(path, "/revisions") {
					delete(body, "expectedRevision")
				}
				response := performRequest(server, http.MethodPost, path, tc.token, body)
				if response.Code != tc.status {
					t.Fatalf("expected %d, got %d: %s", tc.status, response.Code, response.Body.String())
				}
			})
		}
	}
}

func TestConversationWorkflowDraftHTTPPreservesRevisions(t *testing.T) {
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	body := workflowDraftRequest()
	delete(body, "expectedRevision")
	created := performRequest(server, http.MethodPost, "/tenants/home/conversation-workflows", "Bearer dev:owner", body)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	var first struct {
		Data struct {
			ID         string `json:"id"`
			WorkflowID string `json:"workflowId"`
			Number     int    `json:"number"`
		} `json:"data"`
	}
	decodeBody(t, created, &first)
	if first.Data.ID == "" || first.Data.WorkflowID == "" || first.Data.Number != 1 {
		t.Fatalf("invalid first revision: %+v", first)
	}
	path := "/tenants/home/conversation-workflows/" + first.Data.WorkflowID + "/revisions"
	next := performRequest(server, http.MethodPost, path, "Bearer dev:owner", workflowDraftRequest())
	if next.Code != http.StatusCreated {
		t.Fatalf("append: %d %s", next.Code, next.Body.String())
	}
	stale := performRequest(server, http.MethodPost, path, "Bearer dev:owner", workflowDraftRequest())
	if stale.Code != http.StatusConflict {
		t.Fatalf("stale save: %d %s", stale.Code, stale.Body.String())
	}
	invalid := workflowDraftRequest()
	invalid["definition"].(map[string]any)["steps"].([]map[string]any)[0]["providerProfileId"] = "inaccessible-provider"
	invalid["expectedRevision"] = 2
	rejected := performRequest(server, http.MethodPost, path, "Bearer dev:owner", invalid)
	if rejected.Code != http.StatusBadRequest {
		t.Fatalf("inaccessible provider: %d %s", rejected.Code, rejected.Body.String())
	}
	if strings.Contains(rejected.Body.String(), "inaccessible-provider") {
		t.Fatal("provider identity leaked in error")
	}
}

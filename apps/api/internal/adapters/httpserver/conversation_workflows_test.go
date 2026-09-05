package httpserver

import (
	"net/http"
	"testing"
)

func TestConversationWorkflowDraftRoutesRequireTenantConfiguration(t *testing.T) {
	const home = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newProviderProfileTestApp(t, seededState{
		tenants: []seedTenant{{id: home, name: "Home", owner: "owner"}, {id: "other-home", name: "Other", owner: "outsider"}},
	}))
	for _, path := range []string{"/tenants/" + home + "/conversation-workflows", "/tenants/" + home + "/conversation-workflows/workflow-one/revisions"} {
		for _, tc := range []struct {
			name, token string
			status      int
		}{
			{"anonymous", "", http.StatusUnauthorized},
			{"invalid token", "Bearer malformed", http.StatusUnauthorized},
			{"other tenant owner", "Bearer dev:outsider", http.StatusForbidden},
			{"nonmember", "Bearer dev:stranger", http.StatusForbidden},
		} {
			t.Run(path+"/"+tc.name, func(t *testing.T) {
				response := performRequest(server, http.MethodPost, path, tc.token, workflowDraftRequest())
				if response.Code != tc.status {
					t.Fatalf("expected %d, got %d: %s", tc.status, response.Code, response.Body.String())
				}
			})
		}
	}
}

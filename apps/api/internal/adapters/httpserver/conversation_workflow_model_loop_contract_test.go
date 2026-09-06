package httpserver

import (
	"net/http"
	"testing"
)

func modelLoopWorkflowRequest() map[string]any {
	return map[string]any{"definition": map[string]any{
		"name": "Home conversation", "instructions": "Use our household tags when names differ.",
		"budget": map[string]any{"modelCalls": 4, "toolCalls": 4, "elapsedSeconds": 30, "followUpTurns": 2},
	}}
}

func TestWorkflowHTTPUsesModelLoopSettingsWithoutRetiredStages(t *testing.T) {
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	const path = "/tenants/home/conversation-workflows"
	created := performRequest(server, http.MethodPost, path, "Bearer dev:owner", modelLoopWorkflowRequest())
	if created.Code != http.StatusCreated {
		t.Fatalf("model loop settings rejected: %d %s", created.Code, created.Body.String())
	}
	for _, token := range []string{"Bearer dev:viewer", "Bearer dev:stranger", "Bearer dev:outsider"} {
		response := performRequest(server, http.MethodPost, path, token, modelLoopWorkflowRequest())
		if response.Code != http.StatusForbidden {
			t.Fatalf("unauthorized workflow write: %d %s", response.Code, response.Body.String())
		}
	}

	var value struct {
		Data struct {
			ID         string
			WorkflowID string
			Definition map[string]any
		}
	}
	decodeBody(t, created, &value)
	if value.Data.Definition["instructions"] != "Use our household tags when names differ." {
		t.Fatal("general guidance lost")
	}
	budget, ok := value.Data.Definition["budget"].(map[string]any)
	if !ok || budget["toolCalls"] != float64(4) || budget["modelCalls"] != float64(4) {
		t.Fatalf("real loop budgets lost: %+v", budget)
	}
	for _, key := range []string{"steps", "retrieval", "response"} {
		if _, exists := value.Data.Definition[key]; exists {
			t.Fatalf("retired setting still advertised: %s", key)
		}
	}
	if _, exists := budget["evidenceRounds"]; exists {
		t.Fatal("retired evidence rounds still advertised")
	}
}

func TestWorkflowHTTPRejectsRetiredConfigurationFields(t *testing.T) {
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	for _, key := range []string{"steps", "retrieval", "response", "evidenceRounds"} {
		t.Run(key, func(t *testing.T) {
			body := modelLoopWorkflowRequest()
			definition := body["definition"].(map[string]any)
			switch key {
			case "steps":
				definition[key] = []map[string]any{{"kind": "interpret", "attempts": 1}}
			case "retrieval":
				definition[key] = "expanded"
			case "response":
				definition[key] = "grounded"
			case "evidenceRounds":
				definition["budget"].(map[string]any)[key] = 2
			}
			response := performRequest(server, http.MethodPost, "/tenants/home/conversation-workflows", "Bearer dev:owner", body)
			if response.Code != http.StatusUnprocessableEntity && response.Code != http.StatusBadRequest {
				t.Fatalf("retired field accepted: %d %s", response.Code, response.Body.String())
			}
		})
	}
}

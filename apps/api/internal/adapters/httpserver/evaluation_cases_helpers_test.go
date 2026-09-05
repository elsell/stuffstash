package httpserver

import (
	"net/http"
	"testing"
)

type evaluationCaseHistoryWire struct {
	Data []struct {
		ID         string `json:"id"`
		CaseID     string `json:"caseId"`
		Number     int    `json:"number"`
		Definition struct {
			Title string `json:"title"`
		} `json:"definition"`
	} `json:"data"`
	Meta struct {
		Pagination struct {
			NextCursor *string `json:"nextCursor"`
			HasMore    bool    `json:"hasMore"`
		} `json:"pagination"`
	} `json:"meta"`
}

func evaluationCaseRequest() map[string]any {
	return map[string]any{"definition": map[string]any{"title": "Baby clothes", "utterance": "Where are my baby clothes?", "assets": []map[string]any{{"id": "box", "title": "Attic box", "kind": "container"}, {"id": "clothes", "title": "3 to 6 months", "kind": "item", "parentId": "box", "tagNames": []string{"baby", "clothes"}}}, "expectations": map[string]any{"kind": "answer", "referencedAssets": []string{"clothes"}, "locations": []map[string]string{{"assetId": "clothes", "ancestorId": "box"}}}}}
}

func coverEvaluationCaseScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	const collection = "/tenants/{tenantId}/conversation-evaluation-cases"
	const head = collection + "/{caseId}"
	const revisions = head + "/revisions"
	const revision = revisions + "/{revisionId}"
	base := "/tenants/home/conversation-evaluation-cases"
	appendBody := evaluationCaseRequest()
	appendBody["expectedRevision"] = 1
	if adversarial {
		coverage.request(t, server, http.MethodGet, revisions, base+"/unknown/revisions", "Bearer dev:viewer", nil, http.StatusForbidden)
		coverage.request(t, server, http.MethodPost, collection, base, "Bearer dev:viewer", evaluationCaseRequest(), http.StatusForbidden)
		coverage.request(t, server, http.MethodGet, collection, base, "Bearer dev:viewer", nil, http.StatusForbidden)
		coverage.request(t, server, http.MethodGet, head, base+"/unknown", "Bearer dev:viewer", nil, http.StatusForbidden)
		coverage.request(t, server, http.MethodGet, revision, base+"/unknown/revisions/unknown", "Bearer dev:viewer", nil, http.StatusForbidden)
		coverage.request(t, server, http.MethodPost, revisions, base+"/unknown/revisions", "Bearer dev:viewer", appendBody, http.StatusForbidden)
		return
	}
	created := coverage.request(t, server, http.MethodPost, collection, base, "Bearer dev:owner", evaluationCaseRequest(), http.StatusCreated)
	var value struct {
		Data struct {
			ID     string `json:"id"`
			CaseID string `json:"caseId"`
		}
	}
	decodeBody(t, created, &value)
	coverage.request(t, server, http.MethodGet, revisions, base+"/"+value.Data.CaseID+"/revisions", "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, collection, base, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, head, base+"/"+value.Data.CaseID, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodGet, revision, base+"/"+value.Data.CaseID+"/revisions/"+value.Data.ID, "Bearer dev:owner", nil, http.StatusOK)
	coverage.request(t, server, http.MethodPost, revisions, base+"/"+value.Data.CaseID+"/revisions", "Bearer dev:owner", appendBody, http.StatusCreated)
}

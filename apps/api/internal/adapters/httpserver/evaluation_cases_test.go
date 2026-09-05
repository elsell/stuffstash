package httpserver

import (
	"net/http"
	"reflect"
	"testing"
)

func TestEvaluationCaseRoutesRequireTenantConfiguration(t *testing.T) {
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/tenants/home/conversation-evaluation-cases"},
		{http.MethodPost, "/tenants/home/conversation-evaluation-cases/case-one/revisions"},
		{http.MethodGet, "/tenants/home/conversation-evaluation-cases"},
		{http.MethodGet, "/tenants/home/conversation-evaluation-cases/case-one"},
		{http.MethodGet, "/tenants/home/conversation-evaluation-cases/case-one/revisions/revision-one"},
		{http.MethodGet, "/tenants/home/conversation-evaluation-cases/case-one/revisions"},
	} {
		for _, scenario := range []struct {
			name, token string
			status      int
		}{
			{"anonymous", "", http.StatusUnauthorized}, {"malformed", "Bearer malformed", http.StatusUnauthorized},
			{"other tenant owner", "Bearer dev:outsider", http.StatusForbidden}, {"nonmember", "Bearer dev:stranger", http.StatusForbidden},
			{"inventory owner", "Bearer dev:inventory-owner", http.StatusForbidden}, {"viewer", "Bearer dev:viewer", http.StatusForbidden},
		} {
			t.Run(route.method+route.path+scenario.name, func(t *testing.T) {
				body := evaluationCaseRequest()
				if route.method == http.MethodPost && route.path != "/tenants/home/conversation-evaluation-cases" {
					body["expectedRevision"] = 1
				}
				response := performRequest(server, route.method, route.path, scenario.token, body)
				if response.Code != scenario.status {
					t.Fatalf("expected %d got %d: %s", scenario.status, response.Code, response.Body.String())
				}
			})
		}
	}
}

func TestEvaluationCaseHTTPPreservesCompleteDefinition(t *testing.T) {
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	body := evaluationCaseRequest()
	definition := body["definition"].(map[string]any)
	definition["utterance"] = "Lend the baby clothes to Sam"
	expectations := definition["expectations"].(map[string]any)
	expectations["kind"] = "proposal"
	expectations["proposals"] = []map[string]string{{"operation": "checkout", "targetId": "clothes", "details": "For Sam"}}
	expectations["forbiddenOperations"] = []string{"archive"}
	created := performRequest(server, http.MethodPost, "/tenants/home/conversation-evaluation-cases", "Bearer dev:owner", body)
	if created.Code != http.StatusCreated {
		t.Fatalf("create rich case: %d %s", created.Code, created.Body.String())
	}
	var saved struct {
		Data struct {
			ID         string         `json:"id"`
			CaseID     string         `json:"caseId"`
			Definition map[string]any `json:"definition"`
		}
	}
	decodeBody(t, created, &saved)
	for _, suffix := range []string{"", "/revisions/" + saved.Data.ID} {
		read := performRequest(server, http.MethodGet, "/tenants/home/conversation-evaluation-cases/"+saved.Data.CaseID+suffix, "Bearer dev:owner", nil)
		if read.Code != http.StatusOK {
			t.Fatalf("read rich case: %d %s", read.Code, read.Body.String())
		}
		var value struct {
			Data struct {
				Definition map[string]any `json:"definition"`
			}
		}
		decodeBody(t, read, &value)
		if !reflect.DeepEqual(saved.Data.Definition, value.Data.Definition) {
			t.Fatalf("definition changed in storage: %+v", value.Data.Definition)
		}
		assets := value.Data.Definition["assets"].([]any)
		clothes := assets[1].(map[string]any)
		checks := value.Data.Definition["expectations"].(map[string]any)
		location := checks["locations"].([]any)[0].(map[string]any)
		proposal := checks["proposals"].([]any)[0].(map[string]any)
		if clothes["parentId"] != "box" || location["assetId"] != "clothes" || location["ancestorId"] != "box" || proposal["operation"] != "checkout" || proposal["targetId"] != "clothes" || proposal["details"] != "For Sam" {
			t.Fatalf("semantic fields lost: %+v", value.Data.Definition)
		}
	}
}
func TestEvaluationCaseHTTPRevisionsAndPagination(t *testing.T) {
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	created := performRequest(server, http.MethodPost, "/tenants/home/conversation-evaluation-cases", "Bearer dev:owner", evaluationCaseRequest())
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	var first struct {
		Data struct {
			ID     string `json:"id"`
			CaseID string `json:"caseId"`
			Number int    `json:"number"`
		}
	}
	decodeBody(t, created, &first)
	if first.Data.ID == "" || first.Data.CaseID == "" || first.Data.Number != 1 {
		t.Fatal("invalid case revision")
	}
	base := "/tenants/home/conversation-evaluation-cases/" + first.Data.CaseID
	body := evaluationCaseRequest()
	body["expectedRevision"] = 1
	body["definition"].(map[string]any)["title"] = "Edited case"
	appended := performRequest(server, http.MethodPost, base+"/revisions", "Bearer dev:owner", body)
	if appended.Code != http.StatusCreated {
		t.Fatalf("append: %d %s", appended.Code, appended.Body.String())
	}
	stale := performRequest(server, http.MethodPost, base+"/revisions", "Bearer dev:owner", body)
	if stale.Code != http.StatusConflict {
		t.Fatalf("stale save: %d %s", stale.Code, stale.Body.String())
	}
	for _, path := range []string{base, base + "/revisions/" + first.Data.ID} {
		response := performRequest(server, http.MethodGet, path, "Bearer dev:owner", nil)
		if response.Code != http.StatusOK {
			t.Fatalf("read: %d %s", response.Code, response.Body.String())
		}
		var value struct {
			Data struct {
				Number     int `json:"number"`
				Definition struct {
					Title  string `json:"title"`
					Assets []struct {
						TagNames []string `json:"tagNames"`
					} `json:"assets"`
				} `json:"definition"`
			}
		}
		decodeBody(t, response, &value)
		expectedTitle := "Baby clothes"
		expectedNumber := 1
		if path == base {
			expectedTitle = "Edited case"
			expectedNumber = 2
		}
		if value.Data.Number != expectedNumber || value.Data.Definition.Title != expectedTitle || len(value.Data.Definition.Assets) != 2 || len(value.Data.Definition.Assets[1].TagNames) != 2 {
			t.Fatalf("revision fidelity: %+v", value)
		}
	}
	missing := performRequest(server, http.MethodGet, base+"/revisions/missing", "Bearer dev:owner", nil)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing: %d", missing.Code)
	}
	second := performRequest(server, http.MethodPost, "/tenants/home/conversation-evaluation-cases", "Bearer dev:owner", evaluationCaseRequest())
	if second.Code != http.StatusCreated {
		t.Fatal("second case failed")
	}
	listed := performRequest(server, http.MethodGet, "/tenants/home/conversation-evaluation-cases?limit=1", "Bearer dev:owner", nil)
	var page struct {
		Data []struct {
			ID string `json:"id"`
		}
		Meta struct {
			Pagination struct {
				NextCursor string `json:"nextCursor"`
				HasMore    bool   `json:"hasMore"`
			} `json:"pagination"`
		}
	}
	decodeBody(t, listed, &page)
	if listed.Code != http.StatusOK || len(page.Data) != 1 || !page.Meta.Pagination.HasMore || page.Meta.Pagination.NextCursor == "" {
		t.Fatalf("pagination: %s", listed.Body.String())
	}
	next := performRequest(server, http.MethodGet, "/tenants/home/conversation-evaluation-cases?limit=1&cursor="+page.Meta.Pagination.NextCursor, "Bearer dev:owner", nil)
	var nextPage struct {
		Data []struct {
			ID string `json:"id"`
		}
		Meta struct {
			Pagination struct {
				HasMore bool `json:"hasMore"`
			} `json:"pagination"`
		}
	}
	decodeBody(t, next, &nextPage)
	if next.Code != http.StatusOK || len(nextPage.Data) != 1 || nextPage.Data[0].ID == page.Data[0].ID || nextPage.Meta.Pagination.HasMore {
		t.Fatalf("next page: %s", next.Body.String())
	}
	invalid := evaluationCaseRequest()
	invalid["definition"].(map[string]any)["assets"].([]map[string]any)[1]["parentId"] = "outside"
	rejected := performRequest(server, http.MethodPost, "/tenants/home/conversation-evaluation-cases", "Bearer dev:owner", invalid)
	if rejected.Code != http.StatusBadRequest {
		t.Fatalf("invalid fixture: %d", rejected.Code)
	}
}

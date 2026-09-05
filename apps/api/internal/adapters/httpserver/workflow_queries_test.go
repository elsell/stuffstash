package httpserver

import (
	"net/http"
	"testing"
)

func TestWorkflowReadRoutesRequireConfigurePermission(t *testing.T) {
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	for _, path := range []string{"/tenants/home/conversation-workflows", "/tenants/home/conversation-workflows/missing", "/tenants/home/conversation-workflows/missing/revisions", "/tenants/home/conversation-workflows/missing/revisions/missing", "/tenants/home/conversation-workflow-selection"} {
		for _, scenario := range []struct {
			token  string
			status int
		}{{"", 401}, {"Bearer malformed", 401}, {"Bearer dev:outsider", 403}, {"Bearer dev:stranger", 403}, {"Bearer dev:viewer", 403}, {"Bearer dev:inventory-owner", 403}} {
			response := performRequest(server, http.MethodGet, path, scenario.token, nil)
			if response.Code != scenario.status {
				t.Fatalf("%s expected %d got %d: %s", path, scenario.status, response.Code, response.Body.String())
			}
		}
	}
}
func TestWorkflowHTTPReadsRevisionHistoryAndSelection(t *testing.T) {
	application, store := newWorkflowHTTPTestRuntime(t)
	server := NewServer(":0", application)
	queue := evaluationRunRequest(t, server)
	id := queue["workflowId"].(string)
	revisionID := queue["revisionId"].(string)
	base := "/tenants/home/conversation-workflows/" + id
	second := performRequest(server, http.MethodPost, base+"/revisions", "Bearer dev:owner", workflowDraftRequest())
	if second.Code != 201 {
		t.Fatalf("append: %s", second.Body.String())
	}
	var latest struct{ Data struct{ ID string } }
	decodeBody(t, second, &latest)
	for _, entry := range []struct{ path, id string }{{base, latest.Data.ID}, {base + "/revisions/" + revisionID, revisionID}} {
		response := performRequest(server, http.MethodGet, entry.path, "Bearer dev:owner", nil)
		var result struct{ Data struct{ ID string } }
		decodeBody(t, response, &result)
		if response.Code != 200 || result.Data.ID != entry.id {
			t.Fatalf("revision read: %d %s", response.Code, response.Body.String())
		}
	}
	list := performRequest(server, http.MethodGet, "/tenants/home/conversation-workflows?limit=1", "Bearer dev:owner", nil)
	var heads struct {
		Data []struct{ ID, Name, LatestRevisionID string }
	}
	decodeBody(t, list, &heads)
	if list.Code != 200 || len(heads.Data) != 1 || heads.Data[0].ID != id || heads.Data[0].Name != "Home voice" || heads.Data[0].LatestRevisionID != latest.Data.ID {
		t.Fatalf("head projection: %s", list.Body.String())
	}
	page := performRequest(server, http.MethodGet, base+"/revisions?limit=1", "Bearer dev:owner", nil)
	var history struct {
		Data []struct{ ID string }
		Meta struct{ Pagination struct{ NextCursor *string } }
	}
	decodeBody(t, page, &history)
	if page.Code != 200 || len(history.Data) != 1 || history.Data[0].ID != revisionID || history.Meta.Pagination.NextCursor == nil {
		t.Fatalf("history page: %s", page.Body.String())
	}
	next := performRequest(server, http.MethodGet, base+"/revisions?limit=1&cursor="+*history.Meta.Pagination.NextCursor, "Bearer dev:owner", nil)
	var tail struct{ Data []struct{ ID string } }
	decodeBody(t, next, &tail)
	if next.Code != 200 || len(tail.Data) != 1 || tail.Data[0].ID != latest.Data.ID {
		t.Fatalf("history tail: %s", next.Body.String())
	}
	foreign := performRequest(server, http.MethodGet, "/tenants/other-home/conversation-workflows/"+id, "Bearer dev:outsider", nil)
	if foreign.Code != 404 {
		t.Fatalf("foreign revision: %d", foreign.Code)
	}
	before := performRequest(server, http.MethodGet, "/tenants/home/conversation-workflow-selection", "Bearer dev:owner", nil)
	var selection struct {
		Data *struct{ WorkflowID, RevisionID string }
	}
	decodeBody(t, before, &selection)
	if before.Code != 200 || selection.Data != nil {
		t.Fatal("default selection")
	}
	queued := performRequest(server, http.MethodPost, "/tenants/home/conversation-evaluation-runs", "Bearer dev:owner", queue)
	var run struct{ Data struct{ ID string } }
	decodeBody(t, queued, &run)
	if queued.Code != 201 {
		t.Fatal(queued.Body.String())
	}
	completeHTTPWorkflowRun(t, store, run.Data.ID)
	activated := performRequest(server, http.MethodPost, base+"/activation", "Bearer dev:owner", map[string]any{"revisionId": revisionID, "runId": run.Data.ID, "cases": queue["cases"]})
	if activated.Code != 200 {
		t.Fatalf("activate prior revision: %s", activated.Body.String())
	}
	after := performRequest(server, http.MethodGet, "/tenants/home/conversation-workflow-selection", "Bearer dev:owner", nil)
	decodeBody(t, after, &selection)
	if after.Code != 200 || selection.Data == nil || selection.Data.WorkflowID != id || selection.Data.RevisionID != revisionID {
		t.Fatalf("selected revision: %s", after.Body.String())
	}
}

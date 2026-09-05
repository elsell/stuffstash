package httpserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestEvaluationRunRoutesRequireTenantConfiguration(t *testing.T) {
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	const base = "/tenants/home/conversation-evaluation-runs"
	for _, route := range []struct {
		method, path string
		body         any
	}{
		{http.MethodPost, base, map[string]any{"workflowId": "workflow", "revisionId": "revision", "cases": []map[string]string{{"caseId": "case", "revisionId": "revision"}}}},
		{http.MethodGet, base, nil}, {http.MethodGet, base + "/unknown", nil},
		{http.MethodPost, base + "/unknown/cancellation", map[string]int{"expectedVersion": 1}},
	} {
		for _, scenario := range []struct {
			name, token string
			status      int
		}{
			{"anonymous", "", 401}, {"malformed", "Bearer malformed", 401},
			{"other tenant owner", "Bearer dev:outsider", 403}, {"nonmember", "Bearer dev:stranger", 403},
			{"inventory owner", "Bearer dev:inventory-owner", 403}, {"viewer", "Bearer dev:viewer", 403},
		} {
			t.Run(route.method+route.path+scenario.name, func(t *testing.T) {
				response := performRequest(server, route.method, route.path, scenario.token, route.body)
				if response.Code != scenario.status {
					t.Fatalf("expected %d got %d: %s", scenario.status, response.Code, response.Body.String())
				}
			})
		}
	}
}

func TestEvaluationRunHTTPQueueReadCancelAndIsolation(t *testing.T) {
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	body := evaluationRunRequest(t, server)
	const base = "/tenants/home/conversation-evaluation-runs"
	created := performRequest(server, http.MethodPost, base, "Bearer dev:owner", body)
	if created.Code != 201 {
		t.Fatalf("queue: %d %s", created.Code, created.Body.String())
	}
	var queued struct{ Data map[string]any }
	decodeBody(t, created, &queued)
	id, _ := queued.Data["id"].(string)
	if id == "" || queued.Data["state"] != "queued" || queued.Data["version"] != float64(1) || queued.Data["coverage"] != "text_only" || queued.Data["authorId"] != "owner" || queued.Data["totalCases"] != float64(1) {
		t.Fatalf("invalid queued response: %+v", queued.Data)
	}
	if queued.Data["workflowId"] != body["workflowId"] || queued.Data["revisionId"] != body["revisionId"] {
		t.Fatal("workflow pins lost")
	}
	cases := queued.Data["cases"].([]any)
	if len(cases) != 1 || cases[0].(map[string]any)["title"] != "Baby clothes" {
		t.Fatalf("case pins lost: %+v", cases)
	}
	for _, forbidden := range []string{"leaseToken", "LeaseToken", "inputJSON", "progressJSON", "credential", "maxAttempts"} {
		if strings.Contains(created.Body.String(), forbidden) {
			t.Fatalf("private field exposed: %s", forbidden)
		}
	}
	read := performRequest(server, http.MethodGet, base+"/"+id, "Bearer dev:owner", nil)
	if read.Code != 200 {
		t.Fatalf("get: %d %s", read.Code, read.Body.String())
	}
	foreign := performRequest(server, http.MethodGet, "/tenants/other-home/conversation-evaluation-runs/"+id, "Bearer dev:outsider", nil)
	if foreign.Code != 404 {
		t.Fatalf("foreign run exposed: %d %s", foreign.Code, foreign.Body.String())
	}
	for _, version := range []int{2, 1, 1, 2} {
		response := performRequest(server, http.MethodPost, base+"/"+id+"/cancellation", "Bearer dev:owner", map[string]int{"expectedVersion": version})
		expected := 409
		if version == 1 && queued.Data["state"] == "queued" {
			expected = 200
			queued.Data["state"] = "cancelled"
		} else if version == 2 && queued.Data["state"] == "cancelled" {
			expected = 200
		}
		if response.Code != expected {
			t.Fatalf("cancel version %d: expected %d got %d %s", version, expected, response.Code, response.Body.String())
		}
	}
	second := performRequest(server, http.MethodPost, base, "Bearer dev:owner", body)
	if second.Code != 201 {
		t.Fatalf("second queue: %s", second.Body.String())
	}
	listed := performRequest(server, http.MethodGet, base+"?limit=1", "Bearer dev:owner", nil)
	var page struct {
		Data []map[string]any
		Meta struct{ Pagination struct{ NextCursor *string } }
	}
	decodeBody(t, listed, &page)
	if listed.Code != 200 || len(page.Data) != 1 || page.Meta.Pagination.NextCursor == nil {
		t.Fatalf("bounded list: %s", listed.Body.String())
	}
	next := performRequest(server, http.MethodGet, base+"?limit=1&cursor="+*page.Meta.Pagination.NextCursor, "Bearer dev:owner", nil)
	var end struct {
		Data []map[string]any
		Meta struct{ Pagination struct{ NextCursor *string } }
	}
	decodeBody(t, next, &end)
	if next.Code != 200 || len(end.Data) != 1 || end.Meta.Pagination.NextCursor != nil || end.Data[0]["id"] == page.Data[0]["id"] {
		t.Fatalf("next page: %s", next.Body.String())
	}
	badCursor := performRequest(server, http.MethodGet, "/tenants/other-home/conversation-evaluation-runs?cursor="+*page.Meta.Pagination.NextCursor, "Bearer dev:outsider", nil)
	if badCursor.Code != 400 {
		t.Fatalf("foreign cursor: %d", badCursor.Code)
	}
}

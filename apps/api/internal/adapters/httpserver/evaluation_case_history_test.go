package httpserver

import (
	"net/http"
	"net/url"
	"testing"
)

func TestEvaluationCaseHistoryHTTPPaginationAndIsolation(t *testing.T) {
	server := NewServer(":0", newWorkflowHTTPTestApp(t))
	create := func() string {
		t.Helper()
		response := performRequest(server, http.MethodPost, "/tenants/home/conversation-evaluation-cases", "Bearer dev:owner", evaluationCaseRequest())
		if response.Code != http.StatusCreated {
			t.Fatalf("create: %d %s", response.Code, response.Body.String())
		}
		var value struct {
			Data struct {
				CaseID string `json:"caseId"`
			} `json:"data"`
		}
		decodeBody(t, response, &value)
		return value.Data.CaseID
	}
	id := create()
	base := "/tenants/home/conversation-evaluation-cases/" + id + "/revisions"
	for expected := 1; expected < 3; expected++ {
		body := evaluationCaseRequest()
		body["expectedRevision"] = expected
		response := performRequest(server, http.MethodPost, base, "Bearer dev:owner", body)
		if response.Code != http.StatusCreated {
			t.Fatalf("append: %d %s", response.Code, response.Body.String())
		}
	}
	response := performRequest(server, http.MethodGet, base+"?limit=2", "Bearer dev:owner", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("history: %d %s", response.Code, response.Body.String())
	}
	var first evaluationCaseHistoryWire
	decodeBody(t, response, &first)
	if len(first.Data) != 2 || first.Data[0].Number != 1 || first.Data[1].Number != 2 || first.Data[0].Definition.Title != "Baby clothes" || first.Meta.Pagination.NextCursor == nil || !first.Meta.Pagination.HasMore {
		t.Fatalf("history page: %+v", first)
	}
	cursor := url.QueryEscape(*first.Meta.Pagination.NextCursor)
	response = performRequest(server, http.MethodGet, base+"?limit=2&cursor="+cursor, "Bearer dev:owner", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("next history: %d", response.Code)
	}
	var last evaluationCaseHistoryWire
	decodeBody(t, response, &last)
	if len(last.Data) != 1 || last.Data[0].Number != 3 || last.Meta.Pagination.HasMore || last.Meta.Pagination.NextCursor != nil {
		t.Fatalf("last history: %+v", last)
	}
	otherID := create()
	wrongCase := performRequest(server, http.MethodGet, "/tenants/home/conversation-evaluation-cases/"+otherID+"/revisions?cursor="+cursor, "Bearer dev:owner", nil)
	if wrongCase.Code != http.StatusBadRequest {
		t.Fatalf("cross-case cursor: %d %s", wrongCase.Code, wrongCase.Body.String())
	}
	missing := performRequest(server, http.MethodGet, "/tenants/home/conversation-evaluation-cases/missing/revisions", "Bearer dev:owner", nil)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing history: %d", missing.Code)
	}
}

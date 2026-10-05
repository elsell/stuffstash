package httpapi

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCorrelationAndMetadataPreserveWireMeaning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-ID") != "trace-123" || r.Header.Get("Authorization") != "Bearer owner" {
			t.Errorf("headers: %v", r.Header)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("cursor") == "null" {
			io.WriteString(w, `{"data":null,"meta":{"pagination":{"limit":50,"hasMore":false,"nextCursor":null}}}`)
			return
		}
		io.WriteString(w, `{"$schema":"https://example.test/schema","data":[],"meta":{"requestId":"trace-123","tenantId":"home","pagination":{"limit":50,"hasMore":false,"nextCursor":""}}}`)
	}))
	defer server.Close()
	client, err := New(server.URL, "owner", server.Client(), Options{RequestID: "trace-123"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Tenants(context.Background(), ports.Page{Limit: 50})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(result)
	for _, field := range []string{`"data":[]`, `"$schema":"https://example.test/schema"`, `"meta":`, `"requestId":"trace-123"`, `"tenantId":"home"`, `"nextCursor":""`} {
		if !strings.Contains(string(data), field) {
			t.Fatalf("lost %s in %s", field, data)
		}
	}

	nullResult, err := client.Tenants(context.Background(), ports.Page{Limit: 50, Cursor: "null"})
	if err != nil {
		t.Fatal(err)
	}
	nullJSON, _ := json.Marshal(nullResult)
	if !strings.Contains(string(nullJSON), `"data":null`) || !strings.Contains(string(nullJSON), `"nextCursor":null`) {
		t.Fatalf("null values changed: %s", nullJSON)
	}
	if _, err := New(server.URL, "owner", server.Client(), Options{RequestID: "bad\r\nAuthorization: other"}); err == nil {
		t.Fatal("header injection accepted")
	}
}

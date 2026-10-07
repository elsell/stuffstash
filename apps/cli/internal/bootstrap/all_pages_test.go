package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestAllPagesCombinesInventoryListAndPreservesRequest(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		wantCursor := "start"
		if calls == 2 {
			wantCursor = "next"
		}
		if calls > 2 || r.Method != "GET" || r.URL.Path != "/tenants/home/inventories" || r.URL.Query().Get("limit") != "2" || r.URL.Query().Get("cursor") != wantCursor || r.Header.Get("Authorization") != "Bearer owner" {
			t.Errorf("incorrect page request: %s %s", r.Method, r.URL)
		}
		next, more := `"next"`, true
		if calls == 2 {
			next, more = "null", false
		}
		fmt.Fprintf(w, `{"$schema":"inventory-schema","data":[{"id":"inventory-%d","tenantId":"home","name":"Home"}],"meta":{"requestId":"page-%d","tenantId":"home","pagination":{"limit":2,"hasMore":%t,"nextCursor":%s}}}`, calls, calls, more, next)
	}))
	defer server.Close()
	var out, diagnostic bytes.Buffer
	code := Run(context.Background(), []string{"inventories", "list", "--all", "--cursor", "start", "--limit", "2", "--tenant", "home", "--json", "--no-input"}, binaryEnvironment(t, server.URL), &out, &diagnostic)
	var got ports.Result[[]ports.Inventory]
	if code != 0 || json.Unmarshal(out.Bytes(), &got) != nil || len(got.Data) != 2 || got.Data[0].ID != "inventory-1" || got.Data[1].ID != "inventory-2" || calls != 2 || diagnostic.Len() != 0 {
		t.Fatalf("not a complete clean list: code=%d calls=%d out=%s diagnostic=%s", code, calls, &out, &diagnostic)
	}
	if got.Meta == nil || got.Meta.RequestID == nil || *got.Meta.RequestID != "page-2" || got.Pagination == nil || got.Pagination.HasMore || got.Meta.Pagination.HasMore || got.Schema == nil || *got.Schema != "inventory-schema" {
		t.Fatalf("lost final page metadata: %+v", got)
	}
}

func TestAllPagesStopsWithoutPartialSuccess(t *testing.T) {
	for _, mode := range []string{"repeated", "missing", "no-pagination", "canceled", "forbidden"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if calls > 2 {
					t.Error("continued after failure")
					w.WriteHeader(500)
					return
				}
				if r.URL.Query().Get("lifecycleState") != "archived" || r.URL.Query().Get("sort") != "updated_desc" || r.URL.Path != "/tenants/home/inventories/garage/assets" {
					t.Errorf("lost scope or filters: %s", r.URL)
				}
				if mode == "canceled" {
					cancel()
				}
				if mode == "forbidden" && calls == 2 {
					w.WriteHeader(403)
					return
				}
				pagination := `"pagination":{"limit":2,"hasMore":true,"nextCursor":"next"}`
				if calls == 2 && mode == "missing" {
					pagination = `"pagination":{"limit":2,"hasMore":true,"nextCursor":null}`
				}
				if calls == 2 && mode == "no-pagination" {
					pagination = `"requestId":"last"`
				}
				fmt.Fprintf(w, `{"data":[{"id":"asset"}],"meta":{%s}}`, pagination)
			}))
			defer server.Close()
			var out, diagnostic bytes.Buffer
			code := Run(ctx, []string{"assets", "list", "--all", "--limit", "2", "--lifecycle", "archived", "--sort", "updated_desc", "--tenant", "home", "--inventory", "garage", "--json", "--no-input"}, binaryEnvironment(t, server.URL), &out, &diagnostic)
			expected := 1
			category := "protocol"
			if mode == "canceled" {
				expected = 130
				category = "canceled"
			} else if mode == "forbidden" {
				category = "forbidden"
			}
			if code != expected || out.Len() != 0 || !strings.Contains(diagnostic.String(), `"category":"`+category+`"`) {
				t.Fatalf("partial/incorrect failure: %d %s %s", code, &out, &diagnostic)
			}
			if mode == "canceled" && calls != 1 {
				t.Fatal("continued after cancellation")
			}
		})
	}
}

func TestAllPagesRejectsUnsupportedCommandsBeforeAccess(t *testing.T) {
	for _, command := range [][]string{{"assets", "create"}, {"assets", "audit", "asset"}, {"notifications", "read-all"}, {"import-jobs", "list"}, {"provider-profiles", "list"}} {
		for _, option := range []string{"--all", "--all=false"} {
			var out, diagnostic bytes.Buffer
			getenv := func(string) string { return "" }
			code := Run(context.Background(), append(append([]string{}, command...), option, "--no-input", "--json"), getenv, &out, &diagnostic)
			if code != 2 || !strings.Contains(diagnostic.String(), "--all") {
				t.Fatalf("unsupported traversal: %v %d %s", command, code, &diagnostic)
			}
		}
	}
}

func TestAllPagesExpirationKeepsFinalSummaryAndAllItems(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls > 2 || r.URL.Path != "/tenants/home/inventories/garage/expiration-assets" || r.URL.Query().Get("mode") != "soon" || r.URL.Query().Get("q") != "milk" {
			t.Errorf("changed expiration request %s", r.URL)
		}
		next, more := `"next"`, true
		if calls == 2 {
			next, more = "null", false
		}
		fmt.Fprintf(w, `{"data":{"items":[{"id":"asset-%d"}],"counts":{"all":%d,"expired":0,"soon":2},"timezone":"UTC"},"meta":{"pagination":{"limit":2,"hasMore":%t,"nextCursor":%s}}}`, calls, calls, more, next)
	}))
	defer server.Close()
	var out, diagnostic bytes.Buffer
	code := Run(context.Background(), []string{"assets", "expiration", "--all", "--mode", "soon", "--query", "milk", "--limit", "2", "--tenant", "home", "--inventory", "garage", "--json", "--no-input"}, binaryEnvironment(t, server.URL), &out, &diagnostic)
	var got ports.Result[ports.ExpirationWorkspace]
	if code != 0 || json.Unmarshal(out.Bytes(), &got) != nil || len(got.Data.Items) != 2 || got.Data.Items[0].ID != "asset-1" || got.Data.Items[1].ID != "asset-2" || got.Data.Counts.All != 2 || got.Data.Timezone != "UTC" {
		t.Fatalf("lost workspace %d %s %s", code, &out, &diagnostic)
	}
}

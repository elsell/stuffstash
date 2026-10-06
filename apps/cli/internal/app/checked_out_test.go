package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckedOutCommandDispatchesInventoryList(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "GET" || r.URL.Path != "/tenants/home/inventories/tools/checked-out-assets" {
			t.Errorf("wrong route: %s %s", r.Method, r.URL.Path)
		}
		io.WriteString(w, `{"data":[],"meta":{}}`)
	}))
	defer server.Close()
	o, err := Parse([]string{"assets", "checked-out", "--tenant", "home", "--inventory", "tools"}, func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if err := validateCommand(o); err != nil {
		t.Fatal(err)
	}
	api, err := httpapi.New(server.URL, "owner", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	result, err := execute(context.Background(), api, o)
	if err != nil || calls != 1 {
		t.Fatalf("dispatch: %v calls=%d", err, calls)
	}
	list, ok := result.(ports.Result[[]ports.CheckedOutAsset])
	if !ok || list.Data == nil || len(list.Data) != 0 {
		t.Fatalf("empty response changed: %+v", result)
	}
}

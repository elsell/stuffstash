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

func TestPrintJobListPreservesNullAndEmpty(t *testing.T) {
	for _, data := range []string{"null", "[]"} {
		t.Run(data, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, `{"data":`+data+`,"meta":{}}`) }))
			defer server.Close()
			client, err := New(server.URL, "owner", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			result, err := client.PrintJobs(context.Background(), ports.Scope{Tenant: "home", Inventory: "garage"}, ports.Page{Limit: 1}, "")
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(result)
			if err != nil || !strings.Contains(string(raw), `"data":`+data) {
				t.Fatalf("changed list: %s %v", raw, err)
			}
		})
	}
}

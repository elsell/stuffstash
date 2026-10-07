package bootstrap

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type failedResultWriter struct{}

func (failedResultWriter) Write([]byte) (int, error) {
	return 0, errors.New("private-writer-path-and-token")
}

func TestResultFailureDoesNotRepeatCompletedMutation(t *testing.T) {
	for _, jsonOutput := range []bool{false, true} {
		t.Run(map[bool]string{false: "human", true: "json"}[jsonOutput], func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "PATCH" || r.URL.Path != "/tenants/home/inventories/garage/tags/tag" || r.Header.Get("Authorization") != "Bearer owner" {
					t.Error("unexpected mutation")
				}
				io.WriteString(w, `{"data":{"id":"tag","tenantId":"home","inventoryId":"garage","key":"tools","displayName":"Tools","color":"","lifecycleState":"active"},"meta":{}}`)
			}))
			defer server.Close()
			env := binaryEnvironment(t, server.URL)
			args := []string{"tags", "update", "tag", "--tag-color", "", "--tenant", "home", "--inventory", "garage", "--no-input"}
			if jsonOutput {
				args = append(args, "--json")
			}
			var diagnostic bytes.Buffer
			code := Run(context.Background(), args, env, failedResultWriter{}, &diagnostic)
			message := diagnostic.String()
			if code != 1 || calls != 1 {
				t.Fatalf("output failure must not replay mutation: exit=%d calls=%d %s", code, calls, message)
			}
			if !strings.Contains(message, "before you repeat a change") || !strings.Contains(message, "already be complete") || strings.Contains(message, "private-writer") {
				t.Fatalf("unsafe output recovery: %s", message)
			}
			if jsonOutput && !strings.Contains(message, `"category":"output"`) {
				t.Fatalf("missing stable output category: %s", message)
			}
		})
	}
}

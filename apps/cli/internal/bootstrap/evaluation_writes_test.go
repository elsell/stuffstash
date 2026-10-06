package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEvaluationWritesContractAndDenials(t *testing.T) {
	definition := `{"title":"Find things","utterance":"Where is the drill?","assets":[{"id":"tool","title":"Drill","kind":"item","description":null,"parentId":null,"tagNames":["tools"]}],"expectations":{"kind":"answer","referencedAssets":["tool"],"locations":null,"proposals":[{"operation":"rename","targetId":"tool","destinationId":null,"newTitle":"Drill","newKind":null,"details":null}],"forbiddenOperations":["delete"]}}`
	revision := `{"authorId":"owner","createdAt":"2026-01-01T00:00:00Z","definition":` + definition + `,"id":"rev","number":9007199254740993,"caseId":"case"}`
	run := `{"id":"run","state":"queued","version":9007199254740993,"workflowId":"wf","revisionId":"wf-rev","totalCases":1,"completedCases":0,"passedCases":0,"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z","authorId":"owner","coverage":"text_only","cases":[{"caseId":"case","revisionId":"case-rev","title":"Find things"}],"providers":[{"profileId":"profile","configurationId":"config"}],"results":null,"startedAt":null,"finishedAt":null,"failureCode":null}`
	for _, tc := range []struct {
		command          []string
		path, body, data string
	}{
		{[]string{"cases", "create"}, "/conversation-evaluation-cases", `{"$schema":"schema","definition":` + definition + `}`, revision},
		{[]string{"revisions", "create", "case"}, "/conversation-evaluation-cases/case/revisions", `{"$schema":"schema","definition":` + definition + `,"expectedRevision":9007199254740993}`, revision},
		{[]string{"runs", "create"}, "/conversation-evaluation-runs", `{"$schema":"schema","workflowId":"wf","revisionId":"wf-rev","cases":[{"caseId":"case","revisionId":"case-rev"}]}`, run},
	} {
		t.Run(strings.Join(tc.command, "_"), func(t *testing.T) {
			calls, status := 0, 200
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "POST" || r.Header.Get("X-Request-ID") != "trace" {
					t.Error("write fetched or retried")
				}
				if r.URL.Path != "/tenants/home"+tc.path || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				body, _ := io.ReadAll(r.Body)
				if string(body) != tc.body {
					t.Errorf("changed exact input: %s", body)
				}
				w.WriteHeader(status)
				if status != 200 {
					io.WriteString(w, "private-denial")
					return
				}
				io.WriteString(w, `{"$schema":"result-schema","data":`+tc.data+`,"meta":{"requestId":"trace","tenantId":"home"}}`)
			}))
			defer server.Close()
			dir := t.TempDir()
			os.Chmod(dir, 0700)
			store := credentials.File{Path: filepath.Join(dir, "session.json")}
			if err := store.Save(context.Background(), ports.Session{Server: server.URL, Issuer: "https://id.example", Subject: "owner", IDToken: "owner", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			getenv := func(k string) string {
				switch k {
				case "STUFF_STASH_CLI_SERVER":
					return server.URL
				case "STUFF_STASH_CLI_CREDENTIAL_FILE":
					return store.Path
				case "STUFF_STASH_CLI_CONFIG_FILE":
					return filepath.Join(dir, "config", "contexts.json")
				case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
					return "true"
				}
				return ""
			}

			input := filepath.Join(dir, "input.json")
			if err := os.WriteFile(input, []byte(tc.body), 0600); err != nil {
				t.Fatal(err)
			}
			args := append([]string{"evaluation"}, tc.command...)
			args = append(args, "--tenant", "home", "--input", input, "--json", "--no-input", "--request-id", "trace")
			var out, diag bytes.Buffer
			if code := Run(context.Background(), args, getenv, &out, &diag); code != 2 || calls != 0 {
				t.Fatalf("unconfirmed write: %d %s", code, &diag)
			}
			args = append(args, "--yes")
			out.Reset()
			diag.Reset()
			if code := Run(context.Background(), args, getenv, &out, &diag); code != 0 || calls != 1 {
				t.Fatalf("write: %d %s", code, &diag)
			}
			decode := func(body string) map[string]any {
				t.Helper()
				var v map[string]any
				d := json.NewDecoder(strings.NewReader(body))
				d.UseNumber()
				if err := d.Decode(&v); err != nil {
					t.Fatal(err)
				}
				return v
			}
			actual := decode(out.String())
			expected := decode(`{"data":` + tc.data + `,"$schema":"result-schema","meta":{"requestId":"trace","tenantId":"home"}}`)
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("response changed: %#v want %#v", actual, expected)
			}

			for _, denial := range []int{401, 403, 409, 500} {
				status = denial
				before := calls
				out.Reset()
				diag.Reset()
				if code := Run(context.Background(), args, getenv, &out, &diag); code == 0 || calls != before+1 || strings.Contains(out.String()+diag.String(), "private-denial") {
					t.Fatalf("unsafe denial/retry: %d %s", code, &diag)
				}
			}
			status = 200
			before := calls
			if code := Run(context.Background(), append(args, "--tenant", "other"), getenv, &out, &diag); code == 0 || calls != before+1 {
				t.Fatal("wrong scope succeeded")
			}
		})
	}
}

package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/contextfile"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
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

func TestEvaluationInspectionBoundary(t *testing.T) {
	revision := `{"authorId":"owner","caseId":"case","createdAt":"2026-01-01T00:00:00Z","id":"revision","number":9007199254740993,"definition":{"title":"Find drill","utterance":"Where is drill?","assets":[{"id":"drill","kind":"item","title":"Drill","description":null,"parentId":"garage","tagNames":[]},{"id":"garage","kind":"location","title":"Garage"}],"expectations":{"kind":"answer","forbiddenOperations":null,"locations":[{"ancestorId":"garage","assetId":"drill"}],"proposals":[{"operation":"move","targetId":"drill","destinationId":null,"details":"Notes","newKind":"item","newTitle":"New drill"}],"referencedAssets":["drill"]}}}`
	run := `{"authorId":"owner","cases":[{"caseId":"case","revisionId":"revision","title":"Find drill"}],"completedCases":1,"coverage":"inference-only","createdAt":"2026-01-01T00:00:00Z","failureCode":null,"finishedAt":null,"id":"run","passedCases":1,"providers":[{"configurationId":"config","profileId":"profile"}],"results":[{"caseRevisionId":"revision","completedAt":"2026-01-01T00:00:00Z","durationMilliseconds":9007199254740993.25,"modelCalls":9007199254740993,"observation":{"executedOperations":[],"kind":"answer","locations":null,"proposals":[],"referencedAssets":["drill"]},"verdict":{"failures":[{"code":"missing","fixtureId":null,"operation":"move"}],"passed":false}}],"revisionId":"workflow-revision","startedAt":"2026-01-01T00:00:00Z","state":"running","totalCases":1,"updatedAt":"2026-01-01T00:00:00Z","version":9007199254740993,"workflowId":"workflow"}`
	caseHead := `{"createdAt":"2026-01-01T00:00:00Z","id":"case","latestRevision":9007199254740993,"latestRevisionId":"revision","title":"Find drill","updatedAt":"2026-01-01T00:00:00Z"}`
	runHead := `{"completedCases":1,"createdAt":"2026-01-01T00:00:00Z","id":"run","passedCases":1,"revisionId":"workflow-revision","state":"running","totalCases":1,"updatedAt":"2026-01-01T00:00:00Z","version":9007199254740993,"workflowId":"workflow"}`
	cases := []struct {
		command, path, data string
		list                bool
	}{
		{"cases list", "conversation-evaluation-cases", "[" + caseHead + "]", true},
		{"cases show case", "conversation-evaluation-cases/case", revision, false},
		{"revisions list case", "conversation-evaluation-cases/case/revisions", "[" + revision + "]", true},
		{"revisions show case revision", "conversation-evaluation-cases/case/revisions/revision", revision, false},
		{"runs list", "conversation-evaluation-runs", "[" + runHead + "]", true},
		{"runs show run", "conversation-evaluation-runs/run", run, false},
	}
	var expectedPath, data string
	var list, denied bool
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if denied || r.Header.Get("Authorization") != "Bearer owner" || r.URL.Path != "/tenants/home/"+expectedPath {
			w.WriteHeader(403)
			io.WriteString(w, "private-secret")
			return
		}
		if r.Method != "GET" || r.Header.Get("X-Request-ID") != "trace-client" {
			t.Error("wrong method or correlation")
		}
		if list && (r.URL.Query().Get("limit") != "2" || r.URL.Query().Get("cursor") != "next") {
			t.Error("missing pagination")
		}
		if !list && r.URL.RawQuery != "" {
			t.Error("detail query")
		}
		meta := `{"requestId":"trace","tenantId":"home"}`
		if list {
			meta = `{"requestId":"trace","tenantId":"home","pagination":{"limit":2,"nextCursor":"more","hasMore":true}}`
		}
		io.WriteString(w, `{"$schema":"evaluation-schema","data":`+data+`,"meta":`+meta+`}`)
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
	for _, tc := range cases {
		t.Run(tc.command, func(t *testing.T) {
			expectedPath, data, list = tc.path, tc.data, tc.list
			args := append([]string{"evaluation"}, strings.Fields(tc.command)...)
			args = append(args, "--tenant", "home", "--no-input", "--request-id", "trace-client")
			if list {
				args = append(args, "--limit", "2", "--cursor", "next")
			}
			var out, diag bytes.Buffer
			if code := Run(context.Background(), append(args, "--json"), getenv, &out, &diag); code != 0 {
				t.Fatalf("read: %d %s", code, &diag)
			}
			decode := func(raw string) any {
				var v any
				d := json.NewDecoder(strings.NewReader(raw))
				d.UseNumber()
				if err := d.Decode(&v); err != nil {
					t.Fatal(err)
				}
				return v
			}
			got := decode(out.String()).(map[string]any)
			if !reflect.DeepEqual(got["data"], decode(data)) {
				t.Fatalf("data changed: %s", &out)
			}
			if got["$schema"] != "evaluation-schema" || got["meta"].(map[string]any)["requestId"] != "trace" {
				t.Fatal("envelope changed")
			}
			out.Reset()
			diag.Reset()
			if code := Run(context.Background(), args, getenv, &out, &diag); code != 0 || out.Len() == 0 {
				t.Fatalf("human output: %d %s", code, &diag)
			}
			if tc.command == "runs show run" && !strings.Contains(out.String(), "9007199254740993.25") {
				t.Fatal("human details lost duration precision")
			}
			if list && !strings.Contains(out.String(), "more") {
				t.Fatal("missing human pagination")
			}
			denied = true
			out.Reset()
			diag.Reset()
			if code := Run(context.Background(), args, getenv, &out, &diag); code == 0 || strings.Contains(out.String()+diag.String(), "private-secret") {
				t.Fatal("unsafe denial")
			}
			denied = false
			before := calls
			if code := Run(context.Background(), append(args, "--tenant", "other"), getenv, &out, &diag); code == 0 || calls != before+1 {
				t.Fatal("wrong household allowed")
			}
			before = calls
			if code := Run(context.Background(), append(args, "--title", "ignored"), getenv, &out, &diag); code != 2 || calls != before {
				t.Fatal("unrelated input accepted")
			}
			if !list {
				if code := Run(context.Background(), append(args, "--limit", "50"), getenv, &out, &diag); code != 2 || calls != before {
					t.Fatal("detail pagination accepted")
				}
			}
		})
	}
	config := contextfile.Store{Path: filepath.Join(dir, "config", "contexts.json")}
	principal := contexts.Principal(ports.Session{Issuer: "https://id.example", Subject: "owner"})
	if err := (contexts.Manager{Store: config}).Remember(context.Background(), contexts.Entry{Name: "home", Server: server.URL, Principal: principal, Tenant: "home"}); err != nil {
		t.Fatal(err)
	}
	expectedPath, data, list = "conversation-evaluation-runs/run", run, false
	var out, diag bytes.Buffer
	if code := Run(context.Background(), []string{"evaluation", "runs", "show", "run", "--json", "--no-input", "--request-id", "trace-client"}, getenv, &out, &diag); code != 0 {
		t.Fatalf("saved household: %d %s", code, &diag)
	}

}

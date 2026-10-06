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

func TestWorkflowReadContract(t *testing.T) {
	revision := `{"authorId":"owner","createdAt":"2026-01-01T00:00:00Z","definition":{"name":"Find things","instructions":"Look carefully","providerProfileId":"provider","budget":{"elapsedSeconds":9007199254740993,"followUpTurns":2,"modelCalls":3,"toolCalls":4}},"id":"rev","number":9007199254740993,"settingsMigration":"migration","workflowId":"workflow"}`
	head := `{"id":"workflow","name":"Find things","activeRevisionId":"rev","latestRevisionId":"rev","latestRevision":9007199254740993,"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-02T00:00:00Z"}`
	cases := []struct {
		args              []string
		path, data, field string
		list              bool
	}{
		{[]string{"list"}, "", "[" + head + "]", `"latestRevision":9007199254740993`, true},
		{[]string{"show", "workflow"}, "/workflow", revision, `"settingsMigration":"migration"`, false},
		{[]string{"revisions", "list", "workflow"}, "/workflow/revisions", "[" + revision + "]", `"elapsedSeconds":9007199254740993`, true},
		{[]string{"revisions", "show", "workflow", "rev"}, "/workflow/revisions/rev", revision, `"providerProfileId":"provider"`, false},
		{[]string{"selection", "show"}, "selection", `{"workflowId":"workflow","revisionId":"rev"}`, `"revisionId":"rev"`, false},
	}
	calls, status := 0, 200
	var currentPath, currentData string
	var list bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if status != 200 || !strings.HasPrefix(r.URL.Path, "/tenants/home/") {
			code := status
			if code == 200 {
				code = 403
			}
			w.WriteHeader(code)
			io.WriteString(w, "private-denial-body")
			return
		}
		if r.Method != "GET" || r.URL.Path != currentPath || r.Header.Get("Authorization") != "Bearer owner" || r.Header.Get("X-Request-ID") != "workflow-read" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
		}
		if list {
			if r.URL.Query().Get("limit") != "7" || r.URL.Query().Get("cursor") != "a+b/=" {
				t.Error("lost pagination")
			}
		} else if r.URL.RawQuery != "" {
			t.Error("unexpected filters")
		}
		io.WriteString(w, `{"$schema":"workflow-schema","data":`+currentData+`,"meta":{"requestId":"trace","tenantId":"home","pagination":{"limit":7,"hasMore":true,"nextCursor":"next"}}}`)
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
			return filepath.Join(dir, "contexts.json")
		case "STUFF_STASH_CLI_ALLOW_LOOPBACK_HTTP":
			return "true"
		}
		return ""
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, "_"), func(t *testing.T) {
			currentPath = "/tenants/home/conversation-workflows" + tc.path
			if tc.path == "selection" {
				currentPath = "/tenants/home/conversation-workflow-selection"
			}
			currentData, list = tc.data, tc.list
			args := append([]string{"workflows"}, tc.args...)
			args = append(args, "--tenant", "home", "--no-input", "--request-id", "workflow-read")
			if list {
				args = append(args, "--limit", "7", "--cursor", "a+b/=")
			}
			var out, diag bytes.Buffer
			if code := Run(context.Background(), append(args, "--json"), getenv, &out, &diag); code != 0 {
				t.Fatalf("read %d %s", code, &diag)
			}
			for _, field := range []string{tc.field, `"$schema":"workflow-schema"`, `"requestId":"trace"`, `"nextCursor":"next"`} {
				if !strings.Contains(out.String(), field) {
					t.Fatalf("missing %s in %s", field, &out)
				}
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			decode := func(raw []byte) any {
				var v any
				d := json.NewDecoder(bytes.NewReader(raw))
				d.UseNumber()
				if err := d.Decode(&v); err != nil {
					t.Fatal(err)
				}
				return v
			}
			if !reflect.DeepEqual(decode(envelope.Data), decode([]byte(tc.data))) {
				t.Fatalf("response fields changed: %s", envelope.Data)
			}

			out.Reset()
			diag.Reset()
			if code := Run(context.Background(), args, getenv, &out, &diag); code != 0 || out.Len() == 0 {
				t.Fatalf("human %d %s %s", code, &out, &diag)
			}
			for _, denial := range []int{401, 403, 404} {
				status = denial
				out.Reset()
				diag.Reset()
				if code := Run(context.Background(), append(args, "--json"), getenv, &out, &diag); code == 0 || strings.Contains(out.String()+diag.String(), "private-denial-body") {
					t.Fatal("unsafe denial")
				}
			}
			status = 200

			before := calls
			if code := Run(context.Background(), append(args, "--tenant", "other"), getenv, &out, &diag); code == 0 || calls != before+1 {
				t.Fatal("wrong household succeeded")
			}
			before = calls
			if code := Run(context.Background(), append(args, "--title", "ignored"), getenv, &out, &diag); code != 2 || calls != before {
				t.Fatal("unsupported flag reached API")
			}
			if !list {
				if code := Run(context.Background(), append(args, "--limit", "7"), getenv, &out, &diag); code != 2 || calls != before {
					t.Fatal("detail pagination reached API")
				}
			}
			if list {
				currentData = "[]"
				out.Reset()
				diag.Reset()
				if code := Run(context.Background(), append(args, "--json"), getenv, &out, &diag); code != 0 || !strings.Contains(out.String(), `"data":[]`) {
					t.Fatalf("empty list lost: %d %s %s", code, &out, &diag)
				}
			}

			if list || tc.path == "selection" {
				currentData = "null"
				out.Reset()
				diag.Reset()
				if code := Run(context.Background(), append(args, "--json"), getenv, &out, &diag); code != 0 || !strings.Contains(out.String(), `"data":null`) {
					t.Fatalf("null lost %d %s %s", code, &out, &diag)
				}
			}
		})
	}
	currentPath = "/tenants/home/conversation-workflow-selection"
	currentData = "null"
	list = false
	args := []string{"workflows", "selection", "show", "--no-input", "--request-id", "workflow-read"}
	var out, diag bytes.Buffer
	before := calls
	if code := Run(context.Background(), args, getenv, &out, &diag); code != 2 || calls != before {
		t.Fatalf("missing scope %d %s", code, &diag)
	}
	manager := contexts.Manager{Store: contextfile.Store{Path: filepath.Join(dir, "contexts.json")}}
	session, err := store.Load(context.Background(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Remember(context.Background(), contexts.Entry{Name: "home", Server: server.URL, Principal: contexts.Principal(session), Tenant: "home"}); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diag.Reset()
	if code := Run(context.Background(), args, getenv, &out, &diag); code != 0 || !strings.Contains(out.String(), "No workflow selected.") {
		t.Fatalf("saved household %d %s %s", code, &out, &diag)
	}

}

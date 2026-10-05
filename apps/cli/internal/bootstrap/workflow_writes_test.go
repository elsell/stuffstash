package bootstrap

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/credentials"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWorkflowWritesContractAndDenials(t *testing.T) {
	definition := `{"name":"Find things","instructions":null,"providerProfileId":null,"budget":{"elapsedSeconds":9007199254740993,"followUpTurns":0,"modelCalls":3,"toolCalls":4}}`
	revision := `{"authorId":"owner","createdAt":"2026-01-01T00:00:00Z","definition":` + definition + `,"id":"rev","number":9007199254740993,"workflowId":"wf","settingsMigration":"migration"}`
	for _, tc := range []struct {
		command          []string
		path, body, data string
	}{
		{[]string{"create"}, "", `{"$schema":"schema","definition":` + definition + `}`, revision},
		{[]string{"revisions", "create", "wf"}, "/wf/revisions", `{"$schema":"schema","definition":` + definition + `,"expectedRevision":9007199254740993}`, revision},
		{[]string{"activate", "wf"}, "/wf/activation", `{"$schema":"schema","revisionId":"rev","runId":"run","cases":[{"caseId":"case","revisionId":"case-rev"}],"expected":{"workflowId":"old","revisionId":"old-rev"}}`, revision},
	} {
		t.Run(strings.Join(tc.command, "_"), func(t *testing.T) {
			calls, status := 0, 200
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "POST" {
					t.Error("write fetched or retried")
				}
				if r.URL.Path != "/tenants/home/conversation-workflows"+tc.path || r.Header.Get("Authorization") != "Bearer owner" {
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
			args := append([]string{"workflows"}, tc.command...)
			args = append(args, "--tenant", "home", "--input", input, "--json", "--no-input")
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
			for _, field := range []string{`"$schema":"result-schema"`, `"requestId":"trace"`, `"tenantId":"home"`, `"revisionId":"rev"`} {
				if field == `"revisionId":"rev"` {
					field = `"number":9007199254740993`
				}
				if !strings.Contains(out.String(), field) {
					t.Fatalf("missing %s: %s", field, &out)
				}
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

func TestWorkflowWriteHelpExamples(t *testing.T) {
	for _, command := range [][]string{{"workflows", "create"}, {"workflows", "revisions", "create"}, {"workflows", "activate"}} {
		var out, diag bytes.Buffer
		args := append(append([]string{}, command...), "--help")
		code := Run(context.Background(), args, func(string) string { t.Fatal("help read configuration"); return "" }, &out, &diag)
		example := "stuffstash " + strings.Join(command, " ")
		if code != 0 || !strings.Contains(out.String(), example) || strings.Contains(out.String(), "stuffstash stuffstash") || !strings.Contains(out.String(), "--input") {
			t.Fatalf("unusable example: %d %s %s", code, &out, &diag)
		}
	}
}

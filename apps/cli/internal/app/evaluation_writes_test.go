package app

import (
	"bytes"
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEvaluationWriteInputBeforeCredentials(t *testing.T) {
	for _, command := range [][]string{{"evaluation", "cases", "create"}, {"evaluation", "revisions", "create", "case"}, {"evaluation", "runs", "create"}} {
		for _, body := range []string{`{}`, `{"definition":null}`, `{"definition":false}`, `{"definition":{"title":"x","utterance":"y","expectations":{"kind":"answer","privateField":1}}}`, `{"expectedRevision":1.5}`, `{"expectedRevision":9223372036854775808}`, `{"workflowId":"wf","revisionId":"rev","cases":null}`, `{"workflowId":"wf","revisionId":"rev","cases":[]}`, `{"workflowId":"wf","revisionId":"rev","cases":[{"caseId":"case"}]}`, `{"unknown":"private"}`} {
			o, err := Parse(append(append([]string{}, command...), "--input", "-", "--yes", "--json"), func(string) string { return "" })
			if err != nil {
				t.Fatal(err)
			}
			err = (Runner{InputFiles: requestInput(body)}).Run(context.Background(), o)
			var failure *ports.Error
			if !errors.As(err, &failure) || failure.Category != "usage" {
				t.Fatalf("invalid body reached auth: %v", err)
			}
		}
		o, err := Parse(command, func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		if err = (Runner{}).Run(context.Background(), o); err == nil {
			t.Fatal("missing input accepted")
		}
		for _, flag := range []string{"--limit", "--cursor", "--title", "--idempotency-key"} {
			if _, err := Parse(append(append([]string{}, command...), flag, "1"), func(string) string { return "" }); err == nil {
				t.Fatalf("irrelevant flag %s accepted", flag)
			}
		}
	}
}

type evaluationWriteConfirmation struct {
	notice    *bytes.Buffer
	t         *testing.T
	fragments []string
}

func (p evaluationWriteConfirmation) Pick(_ context.Context, _ string, choices []ports.Choice) (string, error) {
	for _, fragment := range append([]string{`Household: "home"`}, p.fragments...) {
		if !strings.Contains(p.notice.String(), fragment) {
			p.t.Fatalf("missing target/warning %q: %s", fragment, p.notice)
		}
	}
	if choices[0].ID != "cancel" {
		p.t.Fatal("confirmation must default to cancel")
	}
	return "cancel", nil
}
func TestEvaluationWritesDeclined(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++ }))
	defer server.Close()
	api, err := httpapi.New(server.URL, "owner", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		command   []string
		body      string
		fragments []string
	}{
		{[]string{"evaluation", "cases", "create"}, `{"definition":{"title":"Find","utterance":"where","expectations":{"kind":"answer"}}}`, []string{"Create an evaluation case"}},
		{[]string{"evaluation", "revisions", "create", "case"}, `{"expectedRevision":9007199254740993,"definition":{"title":"Find","utterance":"where","expectations":{"kind":"answer"}}}`, []string{`Case: "case"`, "9007199254740993"}},
		{[]string{"evaluation", "runs", "create"}, `{"workflowId":"wf","revisionId":"wf-rev","cases":[{"caseId":"case","revisionId":"rev"}]}`, []string{`Workflow: "wf"`, `Revision: "wf-rev"`, "1 case", "model providers"}},
	} {
		var out, notice bytes.Buffer
		runner := Runner{Output: presentation.Output{Stdout: &out, Stderr: &notice}, Observer: lifecycleObserver{}, Picker: evaluationWriteConfirmation{&notice, t, tc.fragments}}
		o := Options{Server: server.URL, Scope: ports.Scope{Tenant: "home"}, Command: tc.command, InputPath: "-", RequestBody: []byte(tc.body)}
		if _, err := prepareEvaluationWrite(o); err != nil {
			t.Fatal(err)
		}
		if err := runner.writeEvaluation(context.Background(), o, api); !errors.Is(err, context.Canceled) || calls != 0 {
			t.Fatalf("declined write executed: %v %d", err, calls)
		}
	}
}

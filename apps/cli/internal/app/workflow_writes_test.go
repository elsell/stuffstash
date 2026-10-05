package app

import (
	"bytes"
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWorkflowWriteInputBeforeCredentials(t *testing.T) {
	for _, command := range [][]string{{"workflows", "create"}, {"workflows", "revisions", "create", "wf"}, {"workflows", "activate", "wf"}} {
		for _, body := range []string{`{}`, `{"definition":null}`, `{"definition":false}`, `{"definition":{"name":"x","budget":{"modelCalls":1.5}}}`, `{"revisionId":"rev","runId":"run","cases":false}`, `{"revisionId":"rev","runId":"run","cases":[],"expected":false}`, `{"unknown":"private"}`} {
			o, err := Parse(append(append([]string{}, command...), "--input", "-", "--yes", "--json"), func(string) string { return "" })
			if err != nil {
				t.Fatal(err)
			}
			err = (Runner{InputFiles: requestInput(body)}).Run(context.Background(), o)
			var failure *ports.Error
			if !errors.As(err, &failure) || failure.Category != "usage" {
				t.Fatalf("invalid input reached authentication: %v", err)
			}
		}
		o, err := Parse(command, func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		if err = (Runner{}).Run(context.Background(), o); err == nil {
			t.Fatal("missing structured input accepted")
		}
	}
}

type workflowConfirmation struct {
	notice *bytes.Buffer
	t      *testing.T
}

func (p workflowConfirmation) Pick(_ context.Context, _ string, choices []ports.Choice) (string, error) {
	for _, fragment := range []string{`household: "home"`, `workflow: "wf"`, "Change the selected workflow"} {
		if !strings.Contains(p.notice.String(), fragment) {
			p.t.Fatalf("missing target/warning: %s", p.notice)
		}
	}
	if choices[0].ID != "cancel" {
		p.t.Fatal("confirmation must default to cancel")
	}
	return "cancel", nil
}
func TestWorkflowActivationDeclinedAndNullEvidence(t *testing.T) {
	calls := 0
	body := `{"revisionId":"rev","runId":"run","cases":null,"expected":null}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		got, _ := io.ReadAll(r.Body)
		if string(got) != body {
			t.Error("null evidence changed")
		}
		io.WriteString(w, `{"data":{"id":"rev"},"meta":{}}`)
	}))
	defer server.Close()
	api, err := httpapi.New(server.URL, "owner", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	var out, notice bytes.Buffer
	runner := Runner{Output: presentation.Output{Stdout: &out, Stderr: &notice}, Observer: lifecycleObserver{}, Picker: workflowConfirmation{&notice, t}}
	o := Options{Server: server.URL, Scope: ports.Scope{Tenant: "home"}, Command: []string{"workflows", "activate", "wf"}, InputPath: "-", RequestBody: []byte(body)}
	if _, err := prepareWorkflowInput(o); err != nil {
		t.Fatal(err)
	}
	if err := runner.writeWorkflow(context.Background(), o, api); !errors.Is(err, context.Canceled) || calls != 0 {
		t.Fatalf("declined write executed: %v %d", err, calls)
	}
	o.Yes = true
	if err := runner.writeWorkflow(context.Background(), o, api); err != nil || calls != 1 {
		t.Fatalf("null activation: %v %d", err, calls)
	}
}

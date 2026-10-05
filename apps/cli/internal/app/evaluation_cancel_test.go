package app

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestInvalidEvaluationCancellationBeforeAuthentication(t *testing.T) {
	for _, body := range []string{`{}`, `{"expectedVersion":null}`, `{"expectedVersion":0}`, `{"expectedVersion":-1}`, `{"expectedVersion":1.5}`, `{"expectedVersion":"2"}`, `{"expectedVersion":9223372036854775808}`, `{"expectedVersion":2,"extra":true}`} {
		o, err := Parse([]string{"evaluation", "runs", "cancel", "run", "--input", "-", "--yes", "--json"}, func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		// Missing credentials port ensures invalid input cannot reach authentication.
		err = (Runner{InputFiles: requestInput(body)}).Run(context.Background(), o)
		var failure *ports.Error
		if !errors.As(err, &failure) || failure.Category != "usage" {
			t.Fatalf("invalid version reached auth: %s %v", body, err)
		}
	}
	for _, args := range [][]string{{"evaluation", "runs", "cancel", "run", "--yes"}, {"evaluation", "runs", "cancel", "run", "--json"}} {
		o, err := Parse(args, func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		if err = (Runner{}).Run(context.Background(), o); err == nil {
			t.Fatal("missing explicit version accepted")
		}
	}
	for _, flag := range []string{"--limit", "--cursor", "--title", "--idempotency-key", "--revision"} {
		if _, err := Parse([]string{"evaluation", "runs", "cancel", "run", flag, "1"}, func(string) string { return "" }); err == nil {
			t.Fatalf("unsupported flag: %s", flag)
		}
	}
}

type evaluationCancellationPicker struct {
	choice string
	notice *bytes.Buffer
	calls  *int
	t      *testing.T
}

func (p evaluationCancellationPicker) Pick(context.Context, string, []ports.Choice) (string, error) {
	if *p.calls != 1 || !strings.Contains(p.notice.String(), "version: 7") || !strings.Contains(p.notice.String(), `household: "home"`) || !strings.Contains(p.notice.String(), `run: "run"`) {
		p.t.Fatal("target/version not shown before confirmation")
	}
	return p.choice, nil
}
func TestEvaluationCancellationInteractiveConfirmation(t *testing.T) {
	for _, choice := range []string{"cancel", "confirm"} {
		t.Run(choice, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Header.Get("Authorization") != "Bearer owner" {
					t.Error("missing authentication")
				}
				switch r.Method {
				case "GET":
					if r.URL.Path != "/tenants/home/conversation-evaluation-runs/run" {
						t.Error("wrong detail scope")
					}
					io.WriteString(w, `{"data":{"id":"run","version":7},"meta":{}}`)
				case "POST":
					if r.URL.Path != "/tenants/home/conversation-evaluation-runs/run/cancellation" {
						t.Error("wrong cancellation scope")
					}
					var body struct {
						Version int64 `json:"expectedVersion"`
					}
					if json.NewDecoder(r.Body).Decode(&body) != nil || body.Version != 7 {
						t.Error("current version lost")
					}
					io.WriteString(w, `{"data":{"id":"run","version":8,"state":"cancellation_requested"},"meta":{}}`)
				default:
					t.Error("wrong method")
				}
			}))
			defer server.Close()
			api, err := httpapi.New(server.URL, "owner", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			var out, notice bytes.Buffer
			runner := Runner{Output: presentation.Output{Stdout: &out, Stderr: &notice}, Observer: lifecycleObserver{}, Picker: evaluationCancellationPicker{choice: choice, notice: &notice, calls: &calls, t: t}}
			o := Options{Server: server.URL, Scope: ports.Scope{Tenant: "home"}, Command: []string{"evaluation", "runs", "cancel", "run"}}
			err = runner.cancelEvaluationRun(context.Background(), o, api)
			if choice == "cancel" {
				if !errors.Is(err, context.Canceled) || calls != 1 {
					t.Fatalf("canceled prompt wrote: %d %v", calls, err)
				}
			} else if err != nil || calls != 2 || !strings.Contains(out.String(), "cancellation_requested") {
				t.Fatalf("confirmation failed: %d %v %s", calls, err, &out)
			}
		})
	}
}

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

func TestEvaluationCancellationExplicitCASAndResult(t *testing.T) {
	response := `{"$schema":"run-schema","meta":{"requestId":"trace","tenantId":"home"},"data":{"id":"run","authorId":"owner","cases":[{"caseId":"case","revisionId":"case-rev","title":"Find drill"}],"completedCases":1,"coverage":"inference-only","createdAt":"2026-01-01T00:00:00Z","failureCode":null,"finishedAt":null,"passedCases":1,"providers":[{"configurationId":"config","profileId":"profile"}],"results":[{"caseRevisionId":"case-rev","completedAt":"2026-01-01T00:00:00Z","durationMilliseconds":9007199254740993.25,"modelCalls":1,"observation":{"executedOperations":[],"kind":"answer","locations":null,"proposals":[],"referencedAssets":["drill"]},"verdict":{"failures":[],"passed":true}}],"revisionId":"workflow-rev","startedAt":"2026-01-01T00:00:00Z","state":"cancellation_requested","totalCases":2,"updatedAt":"2026-01-01T00:00:00Z","version":9007199254740994,"workflowId":"workflow"}}`
	calls, status := 0, 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" {
			t.Error("explicit cancellation fetched or retried")
		}
		if r.Header.Get("Authorization") != "Bearer owner" || r.URL.Path != "/tenants/home/conversation-evaluation-runs/run/cancellation" {
			w.WriteHeader(403)
			io.WriteString(w, "private-secret")
			return
		}
		if r.Header.Get("X-Request-ID") != "client-trace" {
			t.Error("correlation missing")
		}
		var body struct {
			ExpectedVersion int64  `json:"expectedVersion"`
			Schema          string `json:"$schema"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || body.ExpectedVersion != 9007199254740993 || body.Schema != "request-schema" {
			t.Error("CAS or schema changed")
		}
		w.WriteHeader(status)
		if status != 200 {
			io.WriteString(w, "private-secret")
			return
		}
		io.WriteString(w, response)
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
	input := filepath.Join(dir, "request.json")
	if err := os.WriteFile(input, []byte(`{"expectedVersion":9007199254740993,"$schema":"request-schema"}`), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"evaluation", "runs", "cancel", "run", "--tenant", "home", "--input", input, "--json", "--no-input", "--request-id", "client-trace"}
	var out, diag bytes.Buffer
	if code := Run(context.Background(), args, getenv, &out, &diag); code != 2 || calls != 0 {
		t.Fatalf("unconfirmed cancellation: %d %s", code, &diag)
	}
	args = append(args, "--yes")
	out.Reset()
	diag.Reset()
	if code := Run(context.Background(), args, getenv, &out, &diag); code != 0 || calls != 1 {
		t.Fatalf("cancel: %d %s", code, &diag)
	}
	decode := func(raw string) any {
		d := json.NewDecoder(strings.NewReader(raw))
		d.UseNumber()
		var v any
		if err := d.Decode(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	if !reflect.DeepEqual(decode(response), decode(out.String())) {
		t.Fatalf("result changed: %s", &out)
	}
	for _, denial := range []int{401, 403, 409, 500} {
		status = denial
		before := calls
		out.Reset()
		diag.Reset()
		if code := Run(context.Background(), args, getenv, &out, &diag); code == 0 || calls != before+1 || strings.Contains(out.String()+diag.String(), "private-secret") {
			t.Fatalf("unsafe failure %d: %s", denial, &diag)
		}
		if denial == 500 && !strings.Contains(diag.String(), "unknown") {
			t.Fatal("uncertain result was not explained")
		}
		if denial == 409 && !strings.Contains(diag.String(), "evaluation runs show") {
			t.Fatal("missing conflict guidance")
		}
	}
	before := calls
	if code := Run(context.Background(), append(args, "--tenant", "other"), getenv, &out, &diag); code == 0 || calls != before+1 {
		t.Fatal("cross-household cancellation")
	}
}

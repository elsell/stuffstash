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

type portabilityPicker struct{ values []string }

func (p *portabilityPicker) Pick(_ context.Context, _ string, choices []ports.Choice) (string, error) {
	if len(p.values) == 0 {
		return "", errors.New("unexpected prompt")
	}
	value := p.values[0]
	p.values = p.values[1:]
	for _, c := range choices {
		if c.ID == value {
			return value, nil
		}
	}
	return "", errors.New("unknown choice")
}

type portabilityText struct{ value string }

func (p portabilityText) ReadText(context.Context, string, int) (string, error) { return p.value, nil }

type portabilitySecrets struct{ calls int }

func (p *portabilitySecrets) ReadSecret(context.Context, string, int) (string, error) {
	p.calls++
	return "protected-value", nil
}

type portabilityStream struct {
	data   []byte
	closed *bool
}

func (s portabilityStream) OpenStream(context.Context, string) (io.ReadCloser, error) {
	return &portabilityReader{Reader: bytes.NewReader(s.data), closed: s.closed}, nil
}

type portabilityReader struct {
	io.Reader
	closed *bool
}

func (r *portabilityReader) Close() error {
	if r.closed != nil {
		*r.closed = true
	}
	return nil
}
func TestGuidedImportUsesMaskedCredentialsAndExplicitSecurityChoices(t *testing.T) {
	secrets := &portabilitySecrets{}
	r := Runner{Picker: &portabilityPicker{values: []string{"legacy_homebox", "no", "no", "no"}}, TextInput: portabilityText{"https://source.example"}, SecretInput: secrets}
	o, err := r.prepareImportSource(context.Background(), Options{Command: []string{"import-jobs", "preview"}})
	if err != nil {
		t.Fatal(err)
	}
	var input importSourceInput
	json.Unmarshal(o.RequestBody, &input)
	if secrets.calls != 2 || input.Username == nil || *input.Username != "protected-value" || input.Password == nil || *input.Password != "protected-value" || input.AllowPrivateNetwork == nil || *input.AllowPrivateNetwork || input.AllowInsecureTLS == nil || *input.AllowInsecureTLS || input.IncludeImages == nil || *input.IncludeImages {
		t.Fatalf("unsafe source options: %#v", input.SourceType)
	}
	closed := false
	r.StreamFiles = portabilityStream{[]byte("name\nDrill\n"), &closed}
	r.TextInput = portabilityText{"backup.csv"}
	r.Picker = &portabilityPicker{values: []string{"legacy_homebox_csv", "no"}}
	o, err = r.prepareImportSource(context.Background(), Options{Command: []string{"import-jobs", "preview"}})
	if err != nil || !closed {
		t.Fatalf("CSV stream: %v", err)
	}
	json.Unmarshal(o.RequestBody, &input)
	if input.ContentBase64 == nil || *input.ContentBase64 != "bmFtZQpEcmlsbAo=" {
		t.Fatal("CSV bytes not preserved")
	}
}
func TestPortabilityDeclinedConfirmationNeverMutates(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(500) }))
	defer server.Close()
	var out, notice bytes.Buffer
	r := Runner{Output: presentation.Output{Stdout: &out, Stderr: &notice}, Observer: lifecycleObserver{}, ArchiveAPI: func(server, token string) (ports.ArchiveAPI, error) {
		return httpapi.New(server, token, http.DefaultClient)
	}, ImportSources: func(server, token string) (ports.ImportSources, error) {
		return httpapi.New(server, token, http.DefaultClient)
	}}
	for _, command := range [][]string{{"archive-jobs", "approve", "job"}, {"archive-jobs", "retry", "job"}, {"archive-jobs", "delete", "job"}, {"import-jobs", "preview"}, {"import-jobs", "start", "job"}} {
		r.Picker = &portabilityPicker{values: []string{"cancel"}}
		o := Options{Command: command, Server: server.URL, Scope: ports.Scope{Tenant: "home", Inventory: "inv"}, RequestBody: []byte(`{"name":"Restored"}`)}
		var err error
		if command[0] == "archive-jobs" {
			err = r.archiveCommand(context.Background(), o, "owner")
		} else {
			err = r.importSourceCommand(context.Background(), o, "owner")
		}
		if !errors.Is(err, context.Canceled) || calls != 0 {
			t.Fatalf("declined mutation %v %d", err, calls)
		}
	}
	if !strings.Contains(notice.String(), `Household: "home"`) {
		t.Fatal("missing target notice")
	}
}
func TestPortabilityInputAndFlagGuards(t *testing.T) {
	for _, args := range [][]string{{"archive-jobs", "show", "job", "--input", "-"}, {"archive-jobs", "download", "job", "--yes"}, {"archive-jobs", "approve", "job", "--name", "secret"}, {"import-jobs", "preview", "--idempotency-key", "key"}, {"import-jobs", "start", "job", "--password", "secret"}} {
		if _, err := Parse(args, func(string) string { return "" }); err == nil {
			t.Errorf("accepted unrelated flags: %s", args[0])
		}
	}
	for _, command := range [][]string{{"archive-jobs", "create"}, {"archive-jobs", "approve", "job"}, {"import-jobs", "preview"}, {"import-jobs", "start", "job"}} {
		r := Runner{InputFiles: requestInput(`{"unknown":"secret-marker"}`), StreamFiles: portabilityStream{data: []byte(`{"sourceType":"legacy_homebox","includeImages":"secret-marker"}`)}}
		o := Options{Command: command, InputPath: "-"}
		err := r.Run(context.Background(), o)
		var failure *ports.Error
		if !errors.As(err, &failure) || failure.Category != "usage" || strings.Contains(err.Error(), "secret-marker") {
			t.Fatalf("invalid request reached auth: %v", err)
		}
	}
}

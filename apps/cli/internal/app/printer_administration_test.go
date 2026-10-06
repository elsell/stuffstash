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

func TestPrinterAdministrationInvalidBeforeCredentials(t *testing.T) {
	for _, command := range [][]string{{"printers", "create"}, {"printers", "update", "p"}, {"connectors", "print", "update", "c"}, {"print-settings", "update"}} {
		for _, body := range []string{`{}`, `{"revision":null}`, `{"revision":1.5}`, `{"generation":-1}`, `{"name":false}`, `{"unknown":"private"}`} {
			args := append(append([]string{}, command...), "--input", "-", "--yes", "--json")
			if command[0] == "printers" && command[1] == "create" {
				args = append(args, "--idempotency-key", "request")
			}
			o, err := Parse(args, func(string) string { return "" })
			if err != nil {
				t.Fatal(err)
			}
			err = (Runner{InputFiles: requestInput(body)}).Run(context.Background(), o)
			var failure *ports.Error
			if !errors.As(err, &failure) || failure.Category != "usage" {
				t.Fatalf("invalid input reached auth: %v", err)
			}
		}
	}
}

type printerCreationPrompt struct {
	values []string
	calls  int
}

func (p *printerCreationPrompt) ReadText(context.Context, string, int) (string, error) {
	v := p.values[p.calls]
	p.calls++
	return v, nil
}
func TestPrinterAdministrationGuidedCreationAndFlags(t *testing.T) {
	prompt := &printerCreationPrompt{values: []string{"Kitchen", "brother", "roll", "4294967295"}}
	o, err := Parse([]string{"printers", "create", "--idempotency-key", "request"}, func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	o, err = (Runner{TextInput: prompt}).preparePrinterAdministration(context.Background(), o)
	if err != nil || prompt.calls != 4 || !strings.Contains(string(o.RequestBody), `"presetVersion":4294967295`) {
		t.Fatalf("guided create: %s %v", o.RequestBody, err)
	}
	for _, args := range [][]string{
		{"printers", "create", "--input", "-", "--adapter", ""},
		{"printers", "update", "p", "--retired=false"},
		{"printers", "update", "p", "--name", "Name"},
		{"connectors", "print", "update", "c", "--idempotency-key", "key"},
		{"print-settings", "update", "--show-reference=false"},
		{"printers", "show", "p", "--preset-version", "1"},
	} {
		if _, err := Parse(args, func(string) string { return "" }); err == nil {
			t.Fatalf("ignored flag: %v", args)
		}
	}
	for _, version := range []string{"0", "4294967296", "private-secret"} {
		o, err := Parse([]string{"printers", "create", "--name", "Kitchen", "--adapter", "brother", "--label-size", "roll", "--preset-version", version, "--idempotency-key", "key", "--no-input"}, func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		if _, err = (Runner{}).preparePrinterAdministration(context.Background(), o); err == nil || strings.Contains(err.Error(), "private-secret") {
			t.Fatalf("invalid scalar version: %v", err)
		}
	}
}
func TestPrinterAdministrationDeclinedConfirmation(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(500) }))
	defer server.Close()
	api, err := httpapi.New(server.URL, "owner", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range [][]string{{"printers", "create"}, {"printers", "update", "p"}, {"connectors", "print", "update", "c"}, {"print-settings", "update"}} {
		var out, notice bytes.Buffer
		r := Runner{PrinterAdministrationAPI: func(string, string) (ports.PrinterAdministrationAPI, error) { return api, nil }, Picker: firstScopeChoice{}, Output: presentation.Output{Stdout: &out, Stderr: &notice}}
		err := r.printerAdministration(context.Background(), Options{Server: server.URL, Scope: ports.Scope{Tenant: "home", Inventory: "garage"}, Command: command}, "owner")
		if !errors.Is(err, context.Canceled) || calls != 0 || !strings.Contains(notice.String(), `inventory: "garage"`) || !strings.Contains(notice.String(), `household: "home"`) {
			t.Fatalf("declined write: %v %s", err, &notice)
		}
	}
}

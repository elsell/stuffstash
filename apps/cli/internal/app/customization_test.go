package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"testing"
)

func TestCustomizationRejectsBeforeCredentials(t *testing.T) {
	for _, args := range [][]string{
		{"asset-types", "list", "--json"},
		{"field-definitions", "list", "--no-input"},
		{"asset-types", "create", "--scope", "household", "--input", "-"},
		{"field-definitions", "create", "--scope", "inventory", "--input", "-"},
	} {
		o, err := Parse(args, func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		for _, body := range []string{`{}`, `{"key":"key","displayName":false}`, `{"unknown":"secret"}`} {
			err = (Runner{InputFiles: requestInput(body)}).Run(context.Background(), o)
			var failure *ports.Error
			if !errors.As(err, &failure) || failure.Category != "usage" {
				t.Fatalf("invalid input reached credentials: %v", err)
			}
		}
	}
	for _, args := range [][]string{
		{"asset-types", "list", "--scope", "automatic"},
		{"asset-types", "show", "id", "--limit", "2"},
		{"field-definitions", "create", "--input", "-", "--name="},
		{"asset-types", "update", "id", "--input", "-", "--expiration-enabled=false"},
		{"field-definitions", "create", "--description", "ignored"},
		{"asset-types", "archive", "id", "--input", "-"},
		{"asset-types", "list", "--scope="},
	} {
		o, err := Parse(args, func(string) string { return "" })
		if err == nil {
			err = (Runner{}).Run(context.Background(), o)
		}
		if err == nil {
			t.Fatalf("accepted invalid flags: %v", args)
		}
	}
}

type definitionPrompts struct {
	answers    []string
	selections []string
}

func (p *definitionPrompts) ReadText(context.Context, string, int) (string, error) {
	v := p.answers[0]
	p.answers = p.answers[1:]
	return v, nil
}
func (p *definitionPrompts) Pick(_ context.Context, _ string, choices []ports.Choice) (string, error) {
	v := p.selections[0]
	p.selections = p.selections[1:]
	for _, c := range choices {
		if c.ID == v {
			return v, nil
		}
	}
	return "", context.Canceled
}
func TestCustomizationGuidanceAndScope(t *testing.T) {
	ctx := context.Background()
	for _, level := range []string{"household", "inventory"} {
		p := &definitionPrompts{selections: []string{level, "enum"}, answers: []string{"grade", "Grade", "A", "B", ""}}
		r := Runner{Picker: p, TextInput: p}
		o, err := r.chooseDefinitionLevel(ctx, Options{Command: []string{"field-definitions", "create"}, Scope: ports.Scope{Tenant: "home", Inventory: "saved-inventory"}})
		if err != nil || o.DefinitionLevel != level || requiresInventory(o) != (level == "inventory") {
			t.Fatalf("scope selection: %+v %v", o, err)
		}
		o, err = r.prepareInput(ctx, o)
		if err != nil {
			t.Fatal(err)
		}
		var v fieldDefinitionInput
		if err = json.Unmarshal(o.RequestBody, &v); err != nil {
			t.Fatal(err)
		}
		if *v.Key != "grade" || *v.DisplayName != "Grade" || *v.Type != "enum" || len(*v.EnumOptions) != 2 || len(p.answers) != 0 || len(p.selections) != 0 {
			t.Fatalf("guided input: %s", o.RequestBody)
		}
	}
	o, err := Parse([]string{"asset-types", "create", "--scope", "household", "--key", "food", "--name", "Food", "--expiration-enabled=false"}, func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	o, err = (Runner{}).prepareInput(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	var v assetTypeInput
	json.Unmarshal(o.RequestBody, &v)
	if v.ExpirationEnabled == nil || *v.ExpirationEnabled {
		t.Fatal("explicit false lost")
	}
	var out bytes.Buffer
	r := Runner{Picker: &definitionPrompts{selections: []string{"cancel"}}, Output: presentation.Output{Stdout: &out, Stderr: &out}}
	// No API is configured: cancellation must return before constructing or invoking it.
	o = Options{Command: []string{"asset-types", "delete", "id"}, DefinitionLevel: "household", Scope: ports.Scope{Tenant: "home"}}
	if err = r.customizationCommand(ctx, o, "owner"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
}

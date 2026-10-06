package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"testing"
)

type providerSetupPrompt struct {
	choices, text []string
	secretCalls   int
}

func (p *providerSetupPrompt) Pick(_ context.Context, _ string, choices []ports.Choice) (string, error) {
	if len(p.choices) == 0 {
		return "", errors.New("unexpected prompt")
	}
	v := p.choices[0]
	p.choices = p.choices[1:]
	for _, c := range choices {
		if c.ID == v {
			return v, nil
		}
	}
	return "", errors.New("choice not offered")
}
func (p *providerSetupPrompt) ReadText(context.Context, string, int) (string, error) {
	if len(p.text) == 0 {
		return "", errors.New("unexpected text input")
	}
	v := p.text[0]
	p.text = p.text[1:]
	return v, nil
}
func (p *providerSetupPrompt) ReadSecret(context.Context, string, int) (string, error) {
	p.secretCalls++
	return "hidden-value", nil
}
func TestGuidedProviderSetup(t *testing.T) {
	p := &providerSetupPrompt{choices: []string{"local_http", "language_inference", "endpointUrl", "runtimeOptions", "enable", "false", "save"}, text: []string{"Local model", "https://provider.example", `{"limit":9007199254740993,"tools":false}`}}
	r := Runner{Picker: p, TextInput: p, SecretInput: p}
	o, err := r.prepareProviderWrite(context.Background(), Options{Command: []string{"provider-profiles", "create"}})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	json.Unmarshal(o.RequestBody, &fields)
	if string(fields["enable"]) != "false" || string(fields["runtimeOptions"]) != `{"limit":9007199254740993,"tools":false}` {
		t.Fatal(string(o.RequestBody))
	}
	p.choices = []string{"server_adc"}
	o, err = r.prepareProviderWrite(context.Background(), Options{Command: []string{"provider-profiles", "credential", "profile"}})
	if err != nil || p.secretCalls != 0 || string(o.RequestBody) != `{"purpose":"server_adc"}` {
		t.Fatalf("%s %v", o.RequestBody, err)
	}
	p.choices = []string{"api_key"}
	o, err = r.prepareProviderWrite(context.Background(), Options{Command: []string{"provider-profiles", "credential", "profile"}})
	if err != nil || p.secretCalls != 1 {
		t.Fatal(err)
	}
}
func TestProviderSetupNoPromptWithoutTerminalInput(t *testing.T) {
	p := &providerSetupPrompt{}
	for _, o := range []Options{{Command: []string{"provider-profiles", "create"}, NoInput: true}, {Command: []string{"provider-profiles", "credential", "profile"}, JSON: true}} {
		if _, err := (Runner{Picker: p, TextInput: p, SecretInput: p}).prepareProviderWrite(context.Background(), o); err == nil {
			t.Fatal("script prompted")
		}
	}
	if p.secretCalls != 0 {
		t.Fatal("read secret")
	}
	for _, body := range []string{`{"purpose":"api_key","credential":"secret","extra":"value"}`, `{"purpose":"server_adc","credential":"secret"}`, `{"purpose":"api_key","credential":false}`} {
		if err := validateProviderInput("credential", []byte(body)); err == nil {
			t.Fatal("bad credential input accepted")
		}
	}
}

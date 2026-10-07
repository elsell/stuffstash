package app

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
)

func isProviderWrite(o Options) bool {
	return isProviderProfileCommand(o) && len(o.Command) > 1 && (o.Command[1] == "create" || o.Command[1] == "update" || o.Command[1] == "credential")
}
func validateProviderWrite(o Options, scope bool) error {
	n := 3
	if o.Command[1] == "create" {
		n = 2
	}
	if len(o.Command) != n {
		return ports.Failure("usage", "Use provider-profiles create, update PROFILE_ID, or credential PROFILE_ID.")
	}
	if scope && o.Scope.Tenant == "" {
		return ports.Failure("usage", "Supply --tenant or select a saved household context.")
	}
	return nil
}
func providerWriteFlags(o Options, flags *flag.FlagSet) error {
	if !isProviderWrite(o) {
		return nil
	}
	invalid := ""
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "server", "tenant", "context", "credential-file", "allow-loopback-http", "json", "no-input", "request-id", "color", "help", "input", "yes":
		default:
			invalid = f.Name
		}
	})
	if invalid != "" {
		return ports.Failure("usage", "Provider setup does not accept --"+invalid+". Use --input FILE|- or interactive setup. Do not put credentials in command arguments.")
	}
	return nil
}
func validateProviderInput(action string, body []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(body, &fields) != nil || fields == nil {
		return ports.Failure("usage", "Supply a JSON object for provider setup.")
	}
	allowed := map[string]string{"$schema": "string"}
	required := []string{}
	if action == "credential" {
		allowed["purpose"] = "string"
		allowed["credential"] = "string"
		required = []string{"purpose"}
	} else {
		for _, k := range []string{"displayName", "endpointUrl", "modelName", "promptTemplate"} {
			allowed[k] = "string"
		}
		allowed["runtimeOptions"] = "object"
		allowed["capabilityMetadata"] = "object"
		if action == "create" {
			allowed["capability"] = "string"
			allowed["providerKind"] = "string"
			allowed["enable"] = "bool"
			required = []string{"displayName", "capability", "providerKind"}
		}
	}
	for key, value := range fields {
		kind, ok := allowed[key]
		if !ok {
			return ports.Failure("usage", "The provider input has an unsupported field. Examine this command's JSON fields in --help.")
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			continue
		}
		valid := false
		switch kind {
		case "string":
			var v string
			valid = json.Unmarshal(value, &v) == nil
		case "object":
			var v map[string]json.RawMessage
			valid = json.Unmarshal(value, &v) == nil && v != nil
		case "bool":
			var v bool
			valid = json.Unmarshal(value, &v) == nil
		}
		if !valid {
			// Only schema-approved field names can reach this message. Never include input values.
			expected := "a JSON " + kind
			if kind == "bool" {
				expected = "true or false"
			}
			return ports.Failure("usage", "Supply "+expected+" for "+key+".")
		}
	}
	for _, key := range required {
		var v string
		if json.Unmarshal(fields[key], &v) != nil || strings.TrimSpace(v) == "" {
			return ports.Failure("usage", "Supply "+key+" as a JSON string that is not empty.")
		}
	}
	if action == "update" && len(fields) == 0 {
		return ports.Failure("usage", "Supply at least one provider field to change.")
	}
	if action == "credential" {
		var purpose, secret string
		json.Unmarshal(fields["purpose"], &purpose)
		json.Unmarshal(fields["credential"], &secret)
		if purpose != "api_key" && purpose != "oauth_bearer" && purpose != "server_adc" {
			return ports.Failure("usage", "Use credential purpose api_key, oauth_bearer, or server_adc.")
		}
		if purpose == "server_adc" && secret != "" || purpose != "server_adc" && strings.TrimSpace(secret) == "" {
			return ports.Failure("usage", "Supply credential for api_key or oauth_bearer. Omit credential for server_adc.")
		}
	}
	return nil
}
func (r Runner) prepareProviderWrite(ctx context.Context, o Options) (Options, error) {
	if o.InputPath != "" {
		return o, validateProviderInput(o.Command[1], o.RequestBody)
	}
	if r.Picker == nil || r.TextInput == nil || o.JSON || o.NoInput {
		return o, ports.Failure("usage", "Supply --input FILE|- with provider settings and --yes for scripts. Run without --no-input in a terminal for guided setup.")
	}
	body, err := r.guideProviderSetup(ctx, o.Command[1])
	if err != nil {
		return o, err
	}
	o.RequestBody = body
	return o, validateProviderInput(o.Command[1], body)
}
func (r Runner) writeProvider(ctx context.Context, o Options, token string) error {
	if r.ProviderWrites == nil {
		return ports.Failure("configuration", "Provider setup is not available. Update the CLI and try again.")
	}
	profile := "new"
	if len(o.Command) > 2 {
		profile = o.Command[2]
	}
	if err := r.Output.Notice("Provider profile: " + strconv.Quote(profile) + "; Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Action: " + o.Command[1]); err != nil {
		return err
	}
	if err := r.confirmAction(ctx, o, "Save provider settings", "Save", "Change this household's provider settings. Existing clients can use the updated settings. No provider test will run."); err != nil {
		return err
	}
	api, err := r.ProviderWrites(o.Server, token)
	if err != nil {
		return err
	}
	var result ports.Result[ports.ProviderProfile]
	switch o.Command[1] {
	case "create":
		result, err = api.CreateProvider(ctx, o.Scope.Tenant, o.RequestBody)
	case "update":
		result, err = api.UpdateProvider(ctx, o.Scope.Tenant, o.Command[2], o.RequestBody)
	case "credential":
		result, err = api.ReplaceProviderCredential(ctx, o.Scope.Tenant, o.Command[2], o.RequestBody)
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.provider_profile.write.completed")
	return r.Output.Result(result)
}

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
)

type assetTypeInput struct {
	Schema            *string `json:"$schema,omitempty"`
	Key               *string `json:"key,omitempty"`
	DisplayName       *string `json:"displayName,omitempty"`
	Description       *string `json:"description,omitempty"`
	ExpirationEnabled *bool   `json:"expirationEnabled,omitempty"`
}
type fieldDefinitionInput struct {
	Schema             *string   `json:"$schema,omitempty"`
	Key                *string   `json:"key,omitempty"`
	DisplayName        *string   `json:"displayName,omitempty"`
	Type               *string   `json:"type,omitempty"`
	EnumOptions        *[]string `json:"enumOptions,omitempty"`
	Applicability      *string   `json:"applicability,omitempty"`
	CustomAssetTypeIDs *[]string `json:"customAssetTypeIds,omitempty"`
}

func decodeDefinitionInput(body []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return ports.Failure("usage", "Definition JSON has unknown fields or incorrect types. Use the request fields shown by this command's --help.")
	}
	return nil
}
func nonemptyDefinition(v *string) bool { return v != nil && strings.TrimSpace(*v) != "" }
func (r Runner) prepareCustomization(ctx context.Context, o Options) (Options, error) {
	create := o.Command[1] == "create"
	if o.InputPath == "" {
		if err := r.guideDefinition(ctx, &o); err != nil {
			return o, err
		}
		var name *string
		if o.ConnectorName != "" {
			name = &o.ConnectorName
		}
		if o.Command[0] == "asset-types" {
			o.RequestBody, _ = json.Marshal(assetTypeInput{Key: o.TagKey, DisplayName: name, Description: o.TypeDescription, ExpirationEnabled: o.TypeExpiration})
		} else {
			v := fieldDefinitionInput{Key: o.TagKey, DisplayName: name}
			if o.FieldType != "" {
				v.Type = &o.FieldType
			}
			if create && o.FieldType == "enum" {
				options, err := r.guideEnumOptions(ctx, o)
				if err != nil {
					return o, err
				}
				v.EnumOptions = &options
			}
			o.RequestBody, _ = json.Marshal(v)
		}
	}
	if o.Command[0] == "asset-types" {
		var v assetTypeInput
		if err := decodeDefinitionInput(o.RequestBody, &v); err != nil {
			return o, err
		}
		if create && (!nonemptyDefinition(v.Key) || !nonemptyDefinition(v.DisplayName)) {
			return o, ports.Failure("usage", "Creating an asset type requires key and displayName. Supply --key and --name, or --input FILE|-.")
		}
		if !create {
			var keys map[string]json.RawMessage
			json.Unmarshal(o.RequestBody, &keys)
			if _, found := keys["key"]; found {
				return o, ports.Failure("usage", "Asset type updates do not accept key. Use displayName, description, and expirationEnabled.")
			}
		}
	} else {
		var v fieldDefinitionInput
		if err := decodeDefinitionInput(o.RequestBody, &v); err != nil {
			return o, err
		}
		if create && (!nonemptyDefinition(v.Key) || !nonemptyDefinition(v.DisplayName) || !nonemptyDefinition(v.Type)) {
			return o, ports.Failure("usage", "Creating a field requires key, displayName, and type. Supply --key, --name and --field-type, or --input FILE|-.")
		}
		if create && !validFieldType(*v.Type) {
			return o, ports.Failure("usage", "Select field type text, number, boolean, date, url, or enum.")
		}
	}
	return o, nil
}
func validFieldType(v string) bool {
	switch v {
	case "text", "number", "boolean", "date", "url", "enum":
		return true
	}
	return false
}
func (r Runner) guideDefinition(ctx context.Context, o *Options) error {
	interactive := r.TextInput != nil && !o.JSON && !o.NoInput
	create := o.Command[1] == "create"
	if create && o.TagKey == nil && interactive {
		v, err := r.TextInput.ReadText(ctx, "Definition key", 80)
		if err != nil {
			return err
		}
		o.TagKey = &v
	}
	needName := create || o.TypeDescription == nil && o.TypeExpiration == nil
	if needName && o.ConnectorName == "" && interactive {
		v, err := r.TextInput.ReadText(ctx, "Display name", 120)
		if err != nil {
			return err
		}
		o.ConnectorName = v
	}
	if needName && strings.TrimSpace(o.ConnectorName) == "" {
		return ports.Failure("usage", "Supply --name or --input FILE|- with the definition fields.")
	}
	if create && o.Command[0] == "field-definitions" && o.FieldType == "" && r.Picker != nil && !o.JSON && !o.NoInput {
		choices := []ports.Choice{}
		for _, v := range []string{"text", "number", "boolean", "date", "url", "enum"} {
			choices = append(choices, ports.Choice{ID: v, Label: v})
		}
		v, err := r.Picker.Pick(ctx, "Field type", choices)
		if err != nil {
			return err
		}
		if !validFieldType(v) {
			return context.Canceled
		}
		o.FieldType = v
	}
	return nil
}
func (r Runner) guideEnumOptions(ctx context.Context, o Options) ([]string, error) {
	if r.TextInput == nil || o.JSON || o.NoInput {
		return nil, ports.Failure("usage", "An enum field needs options. Supply --input FILE|- with enumOptions as an array of strings.")
	}
	var options []string
	for {
		value, err := r.TextInput.ReadText(ctx, "Enum option (empty finishes after the first option)", 120)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(value) == "" {
			if len(options) > 0 {
				return options, nil
			}
			return nil, ports.Failure("usage", "Supply at least one enum option.")
		}
		options = append(options, value)
	}
}

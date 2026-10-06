package app

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
)

type archiveCreateInput struct {
	Schema      *string `json:"$schema,omitempty"`
	InventoryID string  `json:"inventoryId"`
	Photos      *bool   `json:"photos"`
	OtherFiles  *bool   `json:"otherFiles"`
}
type archiveApproveInput struct {
	Schema *string `json:"$schema,omitempty"`
	Name   string  `json:"name"`
}

func decodePortability(body []byte, v any) error {
	if err := validateJSONObject(body); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return ports.Failure("usage", "Input has unknown fields or incorrect types. Use the JSON shape shown by this command's --help.")
	}
	return nil
}
func (r Runner) preparePortability(ctx context.Context, o Options) (Options, error) {
	if o.InputPath != "" {
		if r.InputFiles == nil {
			return o, ports.Failure("configuration", "JSON input is unavailable. Update the CLI.")
		}
		var body []byte
		var err error
		if isImportSource(o) {
			body, err = r.readImportFile(ctx, o.InputPath, maximumImportJSONBytes)
		} else {
			body, err = r.InputFiles.Read(ctx, o.InputPath)
		}
		if err != nil {
			return o, err
		}
		o.RequestBody = body
	}
	if isImportSource(o) {
		return r.prepareImportSource(ctx, o)
	}
	if o.InputPath == "" {
		if o.JSON || o.NoInput || r.Picker == nil {
			return o, ports.Failure("usage", "Supply --input FILE|- with the complete archive request JSON.")
		}
		if o.Command[1] == "create" {
			return o, nil
		}
		if r.TextInput == nil {
			return o, ports.Failure("usage", "Supply --input FILE|- with the destination inventory name.")
		}
		name, err := r.TextInput.ReadText(ctx, "New inventory name", 120)
		if err != nil {
			return o, err
		}
		o.RequestBody, _ = json.Marshal(map[string]string{"name": name})
	}
	if o.Command[1] == "create" {
		var v archiveCreateInput
		if err := decodePortability(o.RequestBody, &v); err != nil {
			return o, err
		}
		if strings.TrimSpace(v.InventoryID) == "" || v.Photos == nil || v.OtherFiles == nil {
			return o, ports.Failure("usage", "Archive creation requires inventoryId, photos and otherFiles. Supply both booleans explicitly.")
		}
		if o.Scope.Inventory != "" && o.Scope.Inventory != v.InventoryID {
			return o, ports.Failure("usage", "The JSON inventoryId differs from the explicit inventory selection. Choose one target consistently.")
		}
		o.Scope.Inventory = v.InventoryID
		o.Selection.Inventory = v.InventoryID
	} else {
		var v archiveApproveInput
		if err := decodePortability(o.RequestBody, &v); err != nil {
			return o, err
		}
		if strings.TrimSpace(v.Name) == "" {
			return o, ports.Failure("usage", "Restore approval requires a nonempty name for the new inventory.")
		}
	}
	return o, nil
}
func (r Runner) portabilityChoice(ctx context.Context, title, yes, no string) (bool, error) {
	if r.Picker == nil {
		return false, ports.Failure("usage", "Use --input FILE|- outside an interactive terminal.")
	}
	choice, err := r.Picker.Pick(ctx, title, []ports.Choice{{ID: "no", Label: no}, {ID: "yes", Label: yes}})
	if err != nil {
		return false, err
	}
	if choice != "no" && choice != "yes" {
		return false, context.Canceled
	}
	return choice == "yes", nil
}

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
		return ports.Failure("usage", "The input has unknown fields or incorrect types. Use the JSON fields shown by this command's --help.")
	}
	return nil
}
func (r Runner) preparePortability(ctx context.Context, o Options) (Options, error) {
	if o.InputPath != "" {
		if r.InputFiles == nil {
			return o, ports.Failure("configuration", "JSON input is not available. Update the CLI.")
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
			return o, ports.Failure("usage", "Supply inventoryId, photos and otherFiles. Use true or false for each of the two file options.")
		}
		if o.Scope.Inventory != "" && o.Scope.Inventory != v.InventoryID {
			return o, ports.Failure("usage", "The JSON inventoryId does not match the selected inventory. Use the same inventory ID in both places.")
		}
		o.Scope.Inventory = v.InventoryID
		o.Selection.Inventory = v.InventoryID
	} else {
		var v archiveApproveInput
		if err := decodePortability(o.RequestBody, &v); err != nil {
			return o, err
		}
		if strings.TrimSpace(v.Name) == "" {
			return o, ports.Failure("usage", "Supply a name for the new inventory in the restore approval input.")
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

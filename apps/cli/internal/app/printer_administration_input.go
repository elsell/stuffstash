package app

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
)

type printerCreateInput struct {
	Schema        *string `json:"$schema,omitempty"`
	Name          string  `json:"name"`
	AdapterID     string  `json:"adapterId"`
	PresetID      string  `json:"presetId"`
	PresetVersion uint32  `json:"presetVersion"`
}
type printerUpdateInput struct {
	Schema        *string `json:"$schema,omitempty"`
	Revision      uint64  `json:"revision"`
	Name          *string `json:"name,omitempty"`
	PresetID      *string `json:"presetId,omitempty"`
	PresetVersion *uint32 `json:"presetVersion,omitempty"`
	Retired       *bool   `json:"retired,omitempty"`
}
type connectorUpdateInput struct {
	Schema     *string   `json:"$schema,omitempty"`
	Generation uint64    `json:"generation"`
	Name       *string   `json:"name,omitempty"`
	PrinterIDs *[]string `json:"printerIds,omitempty"`
	Revoked    *bool     `json:"revoked,omitempty"`
}
type printSettingsInput struct {
	Schema               *string         `json:"$schema,omitempty"`
	Revision             *uint64         `json:"revision"`
	DefaultPrinterID     json.RawMessage `json:"defaultPrinterId"`
	PrintOnCreateDefault *bool           `json:"printOnCreateDefault"`
	Template             *struct {
		ID      string `json:"id"`
		Version uint32 `json:"version"`
		Options *struct {
			ShowReference *bool `json:"showReference"`
		} `json:"options"`
	} `json:"template"`
}

func decodePrinterAdministration(body []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if !json.Valid(body) || d.Decode(v) != nil {
		return ports.Failure("usage", "The printer administration input has unknown fields or incorrect field types. Use this command's --help for its request shape.")
	}
	return nil
}
func (r Runner) preparePrinterAdministration(ctx context.Context, o Options) (Options, error) {
	if isPrinterCreation(o) && (strings.TrimSpace(o.IdempotencyKey) == "" || len(o.IdempotencyKey) > 200 || strings.ContainsAny(o.IdempotencyKey, "\r\n")) {
		return o, ports.Failure("usage", "Supply --idempotency-key with 1 to 200 characters to create a printer. Use the same key only for the same request.")
	}
	if o.InputPath == "" {
		if !isPrinterCreation(o) {
			return o, ports.Failure("usage", "Supply complete JSON with --input FILE or --input - for stdin, including the current revision or generation. See --help.")
		}
		v := printerCreateInput{Name: o.ConnectorName, AdapterID: o.PrinterAdapterID, PresetID: o.LabelSize}
		version := o.PrinterPresetVersion
		for _, field := range []struct {
			value *string
			label string
		}{{&v.Name, "Printer name"}, {&v.AdapterID, "Printer adapter ID"}, {&v.PresetID, "Printer media preset ID"}, {&version, "Printer media preset version"}} {
			if *field.value == "" && r.TextInput != nil && !o.NoInput && !o.JSON {
				value, err := r.TextInput.ReadText(ctx, field.label, 200)
				if err != nil {
					return o, err
				}
				*field.value = value
			}
		}
		number, err := strconv.ParseUint(version, 10, 32)
		if err != nil || number == 0 {
			return o, ports.Failure("usage", "Supply --preset-version with an integer from 1 through 4294967295, from the printer catalog.")
		}
		v.PresetVersion = uint32(number)
		o.RequestBody, _ = json.Marshal(v)
	}
	switch {
	case isPrinterCreation(o):
		var v printerCreateInput
		if err := decodePrinterAdministration(o.RequestBody, &v); err != nil {
			return o, err
		}
		if strings.TrimSpace(v.Name) == "" || strings.TrimSpace(v.AdapterID) == "" || strings.TrimSpace(v.PresetID) == "" || v.PresetVersion == 0 {
			return o, ports.Failure("usage", "Supply name, adapterId, presetId and presetVersion greater than zero to create a printer. Use JSON or --name, --adapter, --label-size and --preset-version.")
		}
	case o.Command[0] == "printers":
		var v printerUpdateInput
		if err := decodePrinterAdministration(o.RequestBody, &v); err != nil {
			return o, err
		}
		if v.Revision == 0 {
			return o, ports.Failure("usage", "Supply an integer greater than zero for the printer revision.")
		}
	case o.Command[0] == "connectors":
		var v connectorUpdateInput
		if err := decodePrinterAdministration(o.RequestBody, &v); err != nil {
			return o, err
		}
		if v.Generation == 0 {
			return o, ports.Failure("usage", "Supply an integer greater than zero for the connector generation.")
		}
	default:
		var v printSettingsInput
		if err := decodePrinterAdministration(o.RequestBody, &v); err != nil {
			return o, err
		}
		if v.Revision == nil || len(v.DefaultPrinterID) == 0 || v.PrintOnCreateDefault == nil || v.Template == nil || v.Template.ID == "" || v.Template.Version == 0 || v.Template.Options == nil || v.Template.Options.ShowReference == nil {
			return o, ports.Failure("usage", "Supply revision, defaultPrinterId (string or null), printOnCreateDefault, and template. In template, supply id, version greater than zero, and options.showReference.")
		}
		var printer *string
		if err := decodePrinterAdministration(v.DefaultPrinterID, &printer); err != nil {
			return o, err
		}
	}
	return o, nil
}

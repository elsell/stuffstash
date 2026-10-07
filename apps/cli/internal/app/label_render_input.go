package app

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isLabelRender(o Options) bool {
	return len(o.Command) == 3 && o.Command[0] == "labels" && o.Command[1] == "render"
}
func labelRenderFlags(o Options, flags *flag.FlagSet) error {
	if !isLabelRender(o) || o.InputPath == "" {
		return nil
	}
	invalid := ""
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "server", "tenant", "inventory", "context", "json", "no-input", "request-id", "color", "help", "allow-loopback-http", "credential-file", "input", "output":
		default:
			invalid = f.Name
		}
	})
	if invalid != "" {
		return ports.Failure("usage", "This render does not accept --"+invalid+" with the selected input. Use either structured --input or render selection flags.")
	}
	return nil
}
func renderFields(body []byte, required ...string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(body, &fields) != nil || fields == nil {
		return nil, ports.Failure("usage", "Supply JSON objects for media, margins, template and options in the render input.")
	}
	for _, key := range required {
		if raw, ok := fields[key]; !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return nil, ports.Failure("usage", "The render input is missing this field: "+key+". See labels render --help.")
		}
	}
	return fields, nil
}
func decodeRenderObject(body []byte, target any) error {
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if !json.Valid(body) || d.Decode(target) != nil {
		return ports.Failure("usage", "The render input has unknown fields, incorrect types, or versions outside 0 through 4294967295. Use snake_case media and template option names.")
	}
	return nil
}
func prepareLabelRender(o Options) (Options, error) {
	if o.InputPath == "" {
		return o, nil
	}
	var input struct {
		Schema   *string         `json:"$schema,omitempty"`
		Media    json.RawMessage `json:"media"`
		Template json.RawMessage `json:"template"`
		Format   string          `json:"format"`
	}
	if err := decodeRenderObject(o.RequestBody, &input); err != nil {
		return o, err
	}
	if input.Format != "png" && input.Format != "pdf" {
		return o, ports.Failure("usage", "Use png or pdf for the render input format.")
	}
	media, err := renderFields(input.Media, "width_micrometers", "height_micrometers", "margins_micrometers", "resolution_dpi", "raster_width", "raster_height", "orientation", "color_mode", "cut_policy", "display_rotation")
	if err != nil {
		return o, err
	}
	if _, err = renderFields(media["margins_micrometers"], "left", "right", "top", "bottom"); err != nil {
		return o, err
	}
	var dimensions printing.Media
	if err = decodeRenderObject(input.Media, &dimensions); err != nil {
		return o, err
	}
	for _, key := range []string{"preset_id", "version"} {
		if raw, ok := media[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return o, ports.Failure("usage", "Supply the declared type for optional media preset_id and version fields. Omit them instead of null.")
		}
	}
	template, err := renderFields(input.Template, "id", "version", "options")
	if err != nil {
		return o, err
	}
	if _, err = renderFields(template["options"], "show_reference"); err != nil {
		return o, err
	}
	var selection struct {
		ID      string `json:"id"`
		Version uint32 `json:"version"`
		Options struct {
			ShowReference bool `json:"show_reference"`
		} `json:"options"`
	}
	if err = decodeRenderObject(input.Template, &selection); err != nil {
		return o, err
	}
	o.Format = input.Format
	return o, nil
}

package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isBinaryCommand(o Options) bool {
	if len(o.Command) < 2 {
		return false
	}
	switch o.Command[0] + " " + o.Command[1] {
	case "attachments download", "attachments thumbnail", "attachments upload", "labels download", "inventories export":
		return true
	}
	return false
}
func validateBinary(o Options, scope bool) error {
	n := 3
	if o.Command[0] == "attachments" && o.Command[1] != "upload" {
		n = 4
	}
	if o.Command[0] == "inventories" {
		n = 2
	}
	if len(o.Command) != n {
		return ports.Failure("usage", "Supply the required resource IDs. Run this command with --help.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or choose a saved context.")
	}
	if o.Command[1] == "upload" {
		if o.FilePath == "" || o.FilePath == "-" {
			return ports.Failure("usage", "Supply --file PATH to a JPEG, PNG, WebP, or PDF file.")
		}
		if o.Transfer != "direct" && o.Transfer != "api" {
			return ports.Failure("usage", "Use --transfer direct or --transfer api.")
		}
		return nil
	}
	if o.OutputPath == "" {
		return ports.Failure("usage", "Supply --output PATH or --output -.")
	}
	if o.OutputPath == "-" && o.JSON {
		return ports.Failure("usage", "Binary stdout cannot include JSON results. Remove --json, or choose an output file.")
	}
	if o.Command[1] == "thumbnail" && o.Variant != "" && o.Variant != "small" && o.Variant != "medium" && o.Variant != "large" {
		return ports.Failure("usage", "Use --variant small, medium, or large.")
	}
	if o.Command[0] == "inventories" && o.Format != "json" && o.Format != "csv" {
		return ports.Failure("usage", "Choose --format json for complete inventory data, or --format csv for a report.")
	}
	return nil
}
func (r Runner) binaryCommand(ctx context.Context, o Options, token string) error {
	if o.Command[1] == "upload" {
		return r.uploadAttachment(ctx, o, token)
	}
	if r.BinaryAPI == nil || r.BinaryFiles == nil {
		return ports.Failure("configuration", "File downloads are unavailable. Update the CLI and try again.")
	}
	api, err := r.BinaryAPI(o.Server, token)
	if err != nil {
		return err
	}
	var body ports.BinaryContent
	switch o.Command[0] + " " + o.Command[1] {
	case "attachments download":
		body, err = api.AttachmentContent(ctx, o.Scope, o.Command[2], o.Command[3])
	case "attachments thumbnail":
		body, err = api.AttachmentThumbnail(ctx, o.Scope, o.Command[2], o.Command[3], o.Variant)
	case "labels download":
		body, err = api.LabelContent(ctx, o.Scope, o.Command[2])
	case "inventories export":
		body, err = api.ExportInventory(ctx, o.Scope, o.Format)
	}
	if err != nil {
		return err
	}
	defer body.Body.Close()
	if err = r.BinaryFiles.PublishContent(ctx, o.OutputPath, body); err != nil {
		return err
	}
	if o.OutputPath == "-" {
		return nil
	}
	return r.Output.Result(map[string]string{"path": o.OutputPath, "status": "saved"})
}

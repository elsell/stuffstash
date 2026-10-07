package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isAttachmentCommand(o Options) bool { return len(o.Command) > 0 && o.Command[0] == "attachments" }
func validateAttachments(o Options, requireScope bool) error {
	valid := false
	if len(o.Command) == 3 {
		valid = o.Command[1] == "list" && o.Command[2] != ""
	}
	if len(o.Command) == 4 && o.Command[2] != "" && o.Command[3] != "" {
		switch o.Command[1] {
		case "show", "archive", "restore", "delete", "complete-upload":
			valid = true
		}
	}
	if !valid {
		return ports.Failure("usage", "Run attachments --help for the command and its asset, attachment, or upload IDs.")
	}
	if o.IdempotencyKey != "" {
		return ports.Failure("usage", "Attachment commands do not support retry keys. Remove --idempotency-key.")
	}
	if o.Title != "" || o.Kind != "" || o.Parent != "" || o.ConnectorName != "" {
		return ports.Failure("usage", "Attachment commands do not accept asset fields. Remove --title, --kind, --parent, and --name.")
	}
	if requireScope && missingResourceScope(o) {
		return ports.Failure("usage", "Supply --tenant and --inventory, or select a saved context.")
	}
	return nil
}
func (r Runner) attachmentsCommand(ctx context.Context, o Options, token string) error {
	if o.Command[1] == "complete-upload" {
		return r.completeAttachmentUpload(ctx, o, token)
	}
	if r.AttachmentsAPI == nil {
		return ports.Failure("configuration", "Attachment commands are not available. Update the CLI and try again.")
	}
	api, err := r.AttachmentsAPI(o.Server, token)
	if err != nil {
		return err
	}
	action, asset := o.Command[1], o.Command[2]
	if action == "list" {
		result, err := listPages(ctx, o, func(page ports.Page) (ports.Result[[]ports.Attachment], error) {
			return api.Attachments(ctx, o.Scope, asset, page)
		})
		if err != nil {
			return err
		}
		return r.Output.Result(result)
	}
	id := o.Command[3]
	if action == "show" {
		result, err := api.Attachment(ctx, o.Scope, asset, id)
		if err != nil {
			return err
		}
		return r.Output.Result(result)
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory) + ". Asset: " + strconv.Quote(asset) + ". Attachment: " + strconv.Quote(id)); err != nil {
		return err
	}
	switch action {
	case "archive":
		if err := r.confirmAction(ctx, o, "Archive attachment", "Archive", "Hide this attachment from active lists."); err != nil {
			return err
		}
	case "delete":
		if err := r.confirmAction(ctx, o, "Delete attachment", "Delete", "Permanently delete this attachment."); err != nil {
			return err
		}
	}
	var result any
	if action == "delete" {
		err = api.DeleteAttachment(ctx, o.Scope, asset, id)
		result = map[string]string{"status": "deleted", "tenantId": o.Scope.Tenant, "inventoryId": o.Scope.Inventory, "assetId": asset, "attachmentId": id}
	} else {
		result, err = api.ChangeAttachment(ctx, o.Scope, asset, id, ports.AttachmentAction(action))
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.attachment.command.completed")
	return r.Output.Result(result)
}

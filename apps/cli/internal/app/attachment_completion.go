package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func (r Runner) completeAttachmentUpload(ctx context.Context, o Options, token string) error {
	if r.AttachmentUploads == nil {
		return ports.Failure("configuration", "Upload commands are not available. Update the CLI and try again.")
	}
	api, err := r.AttachmentUploads(o.Server, token)
	if err != nil {
		return err
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory) + ". Asset: " + strconv.Quote(o.Command[2])); err != nil {
		return err
	}
	result, err := api.CompleteAttachmentUpload(ctx, o.Scope, o.Command[2], o.Command[3])
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "The upload completion result is unknown. Run attachments list ASSET_ID for this asset before you try again.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.attachment.upload.completed")
	return r.Output.Result(result)
}

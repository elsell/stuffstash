package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"strconv"
	"time"
)

func (r Runner) uploadAttachment(ctx context.Context, o Options, token string) error {
	if r.UploadFiles == nil || r.AttachmentUploads == nil || r.UploadTransfer == nil {
		return ports.Failure("configuration", "File uploads are unavailable. Update the CLI and try again.")
	}
	file, err := r.UploadFiles.OpenUpload(ctx, o.FilePath)
	if err != nil {
		return err
	}
	defer file.Body.Close()
	api, err := r.AttachmentUploads(o.Server, token)
	if err != nil {
		return err
	}
	if err = r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory) + "; asset: " + strconv.Quote(o.Command[2])); err != nil {
		return err
	}
	var result ports.Result[ports.Attachment]
	if o.Transfer == "api" {
		reader, writer := io.Pipe()
		defer reader.Close()
		go func() {
			header, _ := json.Marshal(map[string]string{"fileName": file.FileName, "contentType": file.ContentType})
			_, e := writer.Write(append(header[:len(header)-1], []byte(`,"contentBase64":"`)...))
			if e == nil {
				enc := base64.NewEncoder(base64.StdEncoding, writer)
				_, e = io.Copy(enc, file.Body)
				if e == nil {
					e = enc.Close()
				}
			}
			if e == nil {
				_, e = io.WriteString(writer, `"}`)
			}
			writer.CloseWithError(e)
		}()
		result, err = api.CreateAttachment(ctx, o.Scope, o.Command[2], reader)
	} else {
		b, _ := json.Marshal(map[string]any{"fileName": file.FileName, "contentType": file.ContentType, "sizeBytes": file.SizeBytes})
		started, e := api.StartAttachmentUpload(ctx, o.Scope, o.Command[2], bytes.NewReader(b))
		if e != nil {
			return e
		}
		if started.Data.UploadID == "" {
			return ports.Failure("protocol", "The server did not return an upload ID. Inspect attachments list before another upload.")
		}
		if started.Data.ExpiresAt != "" {
			expires, e := time.Parse(time.RFC3339, started.Data.ExpiresAt)
			if e != nil || !expires.After(r.Clock.Now()) {
				return ports.Failure("protocol", "The upload instructions have expired or are invalid. Start a new upload.")
			}
		}
		if err = r.UploadTransfer.Send(ctx, started.Data, file.FileName, file.ContentType, file.SizeBytes, file.Body); err != nil {
			return err
		}
		result, err = api.CompleteAttachmentUpload(ctx, o.Scope, o.Command[2], started.Data.UploadID)
	}
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if !errors.As(err, &failure) || (failure.Category != "network" && failure.Category != "protocol") {
			return err
		}
		return ports.Failure(failure.Category, "Could not confirm the upload. Run attachments list for this asset before another upload. Do not retry automatically.")
	}
	r.Observer.Event(ctx, "cli.attachment.upload.completed")
	return r.Output.Result(result)
}

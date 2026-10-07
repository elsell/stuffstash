package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func isArchiveCommand(o Options) bool { return len(o.Command) > 0 && o.Command[0] == "archive-jobs" }
func isArchiveBody(o Options) bool {
	return isArchiveCommand(o) && len(o.Command) > 1 && (o.Command[1] == "create" || o.Command[1] == "approve")
}
func archiveMutation(o Options) bool {
	if !isArchiveCommand(o) || len(o.Command) < 2 {
		return false
	}
	switch o.Command[1] {
	case "create", "upload", "approve", "retry", "delete":
		return true
	}
	return false
}
func validateArchive(o Options, scoped bool) error {
	c := o.Command
	valid := len(c) == 2 && (c[1] == "list" || c[1] == "create" || c[1] == "upload") || len(c) == 3 && c[2] != "" && (c[1] == "show" || c[1] == "preview" || c[1] == "approve" || c[1] == "retry" || c[1] == "delete" || c[1] == "download")
	if !valid {
		return ports.Failure("usage", "Use archive-jobs list, create, upload, or show|preview|approve|retry|delete|download JOB_ID.")
	}
	if scoped && missingResourceScope(o) {
		return ports.Failure("usage", "Select a household. To create an archive, also select an inventory.")
	}
	if c[1] == "upload" && o.FilePath == "" {
		return ports.Failure("usage", "Supply --file PATH or --file - with a ZIP stream.")
	}
	if c[1] == "download" && o.OutputPath == "-" && o.JSON {
		return ports.Failure("usage", "Do not combine --output - with --json. Use a file path for JSON status or omit --json for binary stdout.")
	}
	if c[1] == "download" && o.OutputPath == "" {
		return ports.Failure("usage", "Supply --output PATH or --output - for archive bytes.")
	}
	return nil
}
func (r Runner) archiveCommand(ctx context.Context, o Options, token string) error {
	if r.ArchiveAPI == nil {
		return ports.Failure("configuration", "Archive commands are not available. Update the CLI.")
	}
	api, err := r.ArchiveAPI(o.Server, token)
	if err != nil {
		return err
	}
	action := o.Command[1]
	id := ""
	if len(o.Command) == 3 {
		id = o.Command[2]
	}
	if action == "create" && len(o.RequestBody) == 0 {
		photos, e := r.portabilityChoice(ctx, "Include photos in this archive?", "Include photos", "Omit photos")
		if e != nil {
			return e
		}
		other, e := r.portabilityChoice(ctx, "Include other files in this archive?", "Include other files", "Omit other files")
		if e != nil {
			return e
		}
		o.RequestBody, _ = json.Marshal(map[string]any{"inventoryId": o.Scope.Inventory, "photos": photos, "otherFiles": other})
	}
	if archiveMutation(o) {
		detail := map[string]string{"create": "Create an archive of this inventory.", "upload": "Upload a ZIP for restore review; no local extraction or immediate restore.", "approve": "Restore this reviewed archive into a new inventory. Existing inventories are not replaced.", "retry": "Ask the server to retry this archive job.", "delete": "Delete this archive job and its retained archive content."}[action]
		if action == "approve" {
			var v archiveApproveInput
			_ = json.Unmarshal(o.RequestBody, &v)
			detail += " New inventory: " + strconv.Quote(v.Name) + "."
		}
		if err = r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory filter: " + strconv.Quote(o.Scope.Inventory) + ". Job: " + strconv.Quote(id) + ". " + detail); err != nil {
			return err
		}
		if err = r.confirmAction(ctx, o, "Archive: "+action, "Continue", detail); err != nil {
			return err
		}
	}
	if action == "create" || action == "upload" {
		if o.IdempotencyKey == "" {
			var key [16]byte
			if _, err = rand.Read(key[:]); err != nil {
				return err
			}
			o.IdempotencyKey = hex.EncodeToString(key[:])
		}
		if err = r.Output.Notice("Archive request key: " + strconv.Quote(o.IdempotencyKey) + ". After an uncertain result, examine archive-jobs list before you try again with this same key and input."); err != nil {
			return err
		}
	}
	var result any
	switch action {
	case "list":
		result, err = api.ArchiveJobs(ctx, o.Scope, o.Page)
	case "show":
		result, err = api.ArchiveJob(ctx, o.Scope, id)
	case "preview":
		result, err = api.ArchivePreview(ctx, o.Scope, id)
	case "create":
		result, err = api.CreateArchiveJob(ctx, o.Scope.Tenant, o.IdempotencyKey, o.RequestBody)
	case "approve":
		result, err = api.ApproveArchiveJob(ctx, o.Scope, id, o.RequestBody)
	case "retry":
		result, err = api.RetryArchiveJob(ctx, o.Scope, id)
	case "delete":
		err = api.DeleteArchiveJob(ctx, o.Scope, id)
		result = map[string]string{"status": "deleted", "jobId": id, "tenantId": o.Scope.Tenant}
	case "upload":
		if r.StreamFiles == nil {
			return ports.Failure("configuration", "Streaming input is not available. Update the CLI.")
		}
		body, e := r.StreamFiles.OpenStream(ctx, o.FilePath)
		if e != nil {
			return e
		}
		defer body.Close()
		result, err = api.UploadArchive(ctx, o.Scope.Tenant, o.IdempotencyKey, body)
	case "download":
		if r.BinaryFiles == nil {
			return ports.Failure("configuration", "Binary output is not available. Update the CLI.")
		}
		content, e := api.DownloadArchive(ctx, o.Scope, id)
		if e != nil {
			return e
		}
		defer content.Body.Close()
		if err = r.BinaryFiles.PublishContent(ctx, o.OutputPath, content); err != nil {
			return err
		}
		r.Observer.Event(ctx, "cli.archive_job.download.completed")
		if o.OutputPath == "-" {
			return nil
		}
		return r.Output.Result(map[string]string{"status": "downloaded", "path": o.OutputPath, "jobId": id})
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.archive_job."+action+".completed")
	return r.Output.Result(result)
}

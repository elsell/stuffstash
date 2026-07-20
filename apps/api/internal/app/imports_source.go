package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) readImportSource(ctx context.Context, input ImportSourceInput) (importplan.Plan, error) {
	request, err := a.importSourceRequest(input)
	if err != nil {
		return importplan.Plan{}, err
	}
	return a.readImportSourceRequest(ctx, request)
}

func (a App) readImportSourceRequest(ctx context.Context, request ports.ImportSourceRequest) (importplan.Plan, error) {
	if a.importSources == nil {
		return importplan.Plan{}, ErrInvalidInput
	}
	return a.importSources.ReadImportPlan(ctx, request)
}

func (a App) importSourceRequest(input ImportSourceInput) (ports.ImportSourceRequest, error) {
	sourceType := importplan.SourceType(input.SourceType)
	var content []byte
	if strings.TrimSpace(input.ContentBase64) != "" {
		if sourceType != importplan.SourceLegacyHomeboxCSV {
			return ports.ImportSourceRequest{}, NewImportSourceInvalidInputError("Uploaded CSV content is only valid for CSV imports. Choose CSV upload or remove the file content.")
		}
		encoded := strings.TrimSpace(input.ContentBase64)
		if base64.StdEncoding.DecodedLen(len(encoded)) > MaxImportCSVBytes+2 {
			return ports.ImportSourceRequest{}, NewImportSourceInvalidInputError(importCSVTooLargeDetail())
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return ports.ImportSourceRequest{}, NewImportSourceInvalidInputError("CSV import file could not be decoded. Choose a valid exported CSV file and try again.")
		}
		if len(decoded) > MaxImportCSVBytes {
			return ports.ImportSourceRequest{}, NewImportSourceInvalidInputError(importCSVTooLargeDetail())
		}
		content = decoded
	}
	return ports.ImportSourceRequest{
		SourceType:          sourceType,
		BaseURL:             input.BaseURL,
		Username:            input.Username,
		Password:            input.Password,
		IncludeImages:       input.IncludeImages,
		AllowInsecureTLS:    input.AllowInsecureTLS,
		AllowPrivateNetwork: input.AllowPrivateNetwork,
		MaxAttachmentBytes:  int64(a.maxAttachmentBytes),
		FileName:            input.FileName,
		Content:             content,
	}, nil
}

func importCSVTooLargeDetail() string {
	return fmt.Sprintf("CSV import file is too large. Choose a CSV up to %d MB.", MaxImportCSVBytes/(1024*1024))
}

func (a App) importJobCommand(input StartImportJobInput) (ports.ImportJobCommand, error) {
	return ports.ImportJobCommand{
		Principal:   input.Principal,
		RequestID:   input.RequestID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		JobID:       input.JobID,
	}, nil
}

func (a App) importJobSourceScope(tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) ports.ImportJobSourceScope {
	return ports.ImportJobSourceScope{TenantID: tenantID, InventoryID: inventoryID, JobID: jobID}
}

func (a App) storeImportJobSource(ctx context.Context, job importjob.Record, request ports.ImportSourceRequest) error {
	if a.importSourceVault == nil {
		return ErrInvalidInput
	}
	now := a.clock.Now().UTC()
	return a.importSourceVault.StoreImportJobSource(ctx, a.importJobSourceScope(importJobTenantID(job.TenantID), importJobInventoryID(job.InventoryID), job.ID), request, now.Add(a.importJobTimeout), now)
}

func (a App) importJobSourceRequest(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (ports.ImportSourceRequest, error) {
	if a.importSourceVault == nil {
		return ports.ImportSourceRequest{}, ErrInvalidInput
	}
	request, found, err := a.importSourceVault.ImportJobSourceRequest(ctx, a.importJobSourceScope(tenantID, inventoryID, jobID))
	if err != nil {
		return ports.ImportSourceRequest{}, err
	}
	if !found {
		return ports.ImportSourceRequest{}, ErrPrecondition
	}
	return request, nil
}

func importSourceInputError(err error) error {
	var userError ports.ImportSourceUserError
	if errors.As(err, &userError) {
		return NewImportSourceInvalidInputError(strings.TrimSpace(userError.Detail))
	}
	return ErrInvalidInput
}

func (a App) normalizedImportPlanForJob(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, plan importplan.Plan) (importplan.Plan, error) {
	plan = cloneImportPlan(plan)
	fingerprint, err := sourceFingerprint(plan)
	if err != nil {
		return importplan.Plan{}, err
	}
	sourceLinkWarnings, linkedAssetSourceIDs, err := a.sourceLinkDuplicateWarnings(ctx, tenantID, inventoryID, importJobSourceRefFromPlan(plan, fingerprint), plan)
	if err != nil {
		return importplan.Plan{}, err
	}
	plan.Messages = append(plan.Messages, sourceLinkWarnings...)
	plan.Messages = append(plan.Messages, a.duplicateWarnings(ctx, tenantID, inventoryID, plan, linkedAssetSourceIDs)...)
	plan.Messages = append(plan.Messages, archivedWarnings(plan)...)
	plan.Messages = safeImportMessages(plan.Messages)
	stripAttachmentContent(&plan)
	return plan, nil
}

func cloneImportPlan(plan importplan.Plan) importplan.Plan {
	clone := plan
	clone.Fields = append([]importplan.FieldDefinition(nil), plan.Fields...)
	clone.Assets = make([]importplan.Asset, len(plan.Assets))
	for index, planned := range plan.Assets {
		clone.Assets[index] = planned
		if planned.CustomFields != nil {
			clone.Assets[index].CustomFields = make(map[string]any, len(planned.CustomFields))
			for key, value := range planned.CustomFields {
				clone.Assets[index].CustomFields[key] = value
			}
		}
	}
	clone.Attachments = make([]importplan.Attachment, len(plan.Attachments))
	for index, attachment := range plan.Attachments {
		clone.Attachments[index] = attachment
		clone.Attachments[index].Content = nil
		clone.Attachments[index].SizeBytes = 0
	}
	clone.Messages = append([]importplan.Message(nil), plan.Messages...)
	return clone
}

func safeImportMessages(messages []importplan.Message) []importplan.Message {
	safe := make([]importplan.Message, 0, len(messages))
	for _, message := range messages {
		message.Code = safeImportMessageText(message.Code)
		message.Summary = safeImportMessageText(message.Summary)
		message.Detail = safeImportMessageText(message.Detail)
		message.SourceID = safeImportMessageText(message.SourceID)
		message.SourceName = safeImportMessageText(message.SourceName)
		safe = append(safe, message)
	}
	return safe
}

func safeImportMessageText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	unsafeFragments := []string{
		"password",
		"passwd",
		"bearer ",
		"authorization:",
		"token=",
		"access_token",
		"refresh_token",
		"secret",
		"ciphertext",
		"nonce",
		"storage key",
		"s3://",
		"file://",
	}
	for _, fragment := range unsafeFragments {
		if strings.Contains(lower, fragment) {
			return ""
		}
	}
	const maxImportMessageTextLength = 240
	if len(value) > maxImportMessageTextLength {
		return value[:maxImportMessageTextLength]
	}
	return value
}

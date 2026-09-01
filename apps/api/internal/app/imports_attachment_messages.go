package app

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type importAttachmentSessionStartError struct {
	detail string
}

type importAttachmentStorageError struct{}

func (importAttachmentStorageError) Error() string {
	return "Stuff Stash could not save the image to configured media storage. Check Garage or S3 connectivity and configuration, then preview and start a new import. Already imported records were kept."
}

func (e importAttachmentSessionStartError) Error() string {
	return e.detail
}

func importAttachmentSessionFailureMessage(err error) importplan.Message {
	detail := "Stuff Stash could not establish a source session for image downloads. Check that the source is reachable, then preview and start a new import. Already imported records were kept."
	var userError ports.ImportSourceUserError
	if errors.As(err, &userError) && strings.TrimSpace(userError.Detail) != "" {
		detail = strings.TrimSpace(userError.Detail) + ". Check the source connection, then preview and start a new import. Already imported records were kept."
	} else if errors.Is(err, context.DeadlineExceeded) {
		detail = "The source did not respond before image import timed out. Check the source connection, then preview and start a new import. Already imported records were kept."
	}
	return importplan.Message{
		Code:     "attachment-session-unavailable",
		Severity: importplan.SeverityError,
		Summary:  "Image import could not start",
		Detail:   detail,
	}
}

func importAttachmentReadFailureMessage(err error, attachment importplan.Attachment) importplan.Message {
	message := importplan.Message{
		Code:       "attachment-unavailable",
		Severity:   importplan.SeverityWarning,
		Summary:    "Attachment could not be downloaded",
		Detail:     "attachment could not be downloaded",
		SourceID:   attachment.SourceID,
		SourceName: attachment.FileName,
	}
	var readErr ports.ImportAttachmentReadError
	if !errors.As(err, &readErr) {
		return message
	}
	switch readErr.Reason {
	case ports.ImportAttachmentTooLarge:
		message.Code = "attachment-too-large"
		message.Summary = "Attachment is too large"
		message.Detail = "attachment exceeds the import size limit"
	case ports.ImportAttachmentUnsupportedType:
		message.Code = "attachment-unsupported-type"
		message.Summary = "Attachment type is not supported"
		message.Detail = "attachment content type is not supported"
	}
	return message
}

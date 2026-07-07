package app

import (
	"fmt"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
)

var (
	ErrUnauthenticated = apperrors.ErrUnauthenticated
	ErrUnauthorized    = apperrors.ErrUnauthorized
	ErrValidation      = apperrors.ErrValidation
	ErrConflict        = apperrors.ErrConflict
	ErrPrecondition    = apperrors.ErrPrecondition
	ErrNotFound        = apperrors.ErrNotFound

	// ErrInvalidInput is retained for existing application call sites that have
	// not yet moved to the more precise validation/conflict/precondition vocabulary.
	ErrInvalidInput                     = apperrors.ErrInvalidInput
	ErrAttachmentFileNameInvalid        = fmt.Errorf("%w: invalid attachment file name", ErrInvalidInput)
	ErrAttachmentContentTypeUnsupported = fmt.Errorf("%w: unsupported attachment file type", ErrInvalidInput)
	ErrAttachmentContentMismatch        = fmt.Errorf("%w: attachment content type mismatch", ErrInvalidInput)
	ErrAttachmentContentEmpty           = fmt.Errorf("%w: empty attachment content", ErrInvalidInput)
	ErrAttachmentTooLarge               = fmt.Errorf("%w: attachment too large", ErrInvalidInput)
)

type ImportSourceInvalidInputError struct {
	Detail string
}

func (e ImportSourceInvalidInputError) Error() string {
	return e.Detail
}

func (e ImportSourceInvalidInputError) Unwrap() error {
	return ErrInvalidInput
}

func NewImportSourceInvalidInputError(detail string) error {
	if detail == "" {
		detail = "Invalid request."
	}
	return ImportSourceInvalidInputError{Detail: detail}
}

type ImportSourceChangedAfterPreviewError struct{}

func (ImportSourceChangedAfterPreviewError) Error() string {
	return "Import source changed after preview. Preview the source again before starting the import."
}

func (ImportSourceChangedAfterPreviewError) Unwrap() error {
	return ErrPrecondition
}

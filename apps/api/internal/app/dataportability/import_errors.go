package dataportability

import "github.com/stuffstash/stuff-stash/internal/app/apperrors"

type ImportSourceInvalidInputError struct {
	Detail string
}

func (e ImportSourceInvalidInputError) Error() string {
	return e.Detail
}

func (e ImportSourceInvalidInputError) Unwrap() error {
	return apperrors.ErrInvalidInput
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
	return apperrors.ErrPrecondition
}

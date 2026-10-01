package app

import (
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	inventoryapp "github.com/stuffstash/stuff-stash/internal/app/inventories"
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
	ErrAttachmentFileNameInvalid        = apperrors.ErrAttachmentFileNameInvalid
	ErrAttachmentContentTypeUnsupported = apperrors.ErrAttachmentContentTypeUnsupported
	ErrAttachmentContentMismatch        = apperrors.ErrAttachmentContentMismatch
	ErrAttachmentContentEmpty           = apperrors.ErrAttachmentContentEmpty
	ErrAttachmentTooLarge               = apperrors.ErrAttachmentTooLarge
	ErrInvitationInvalid                = inventoryapp.ErrInvitationInvalid
	ErrInvitationEmailMismatch          = inventoryapp.ErrInvitationEmailMismatch
)

type ImportSourceInvalidInputError = dataportability.ImportSourceInvalidInputError
type ImportSourceChangedAfterPreviewError = dataportability.ImportSourceChangedAfterPreviewError

func NewImportSourceInvalidInputError(detail string) error {
	return dataportability.NewImportSourceInvalidInputError(detail)
}

package apperrors

import "fmt"

var (
	ErrAttachmentFileNameInvalid        = fmt.Errorf("%w: invalid attachment file name", ErrInvalidInput)
	ErrAttachmentContentTypeUnsupported = fmt.Errorf("%w: unsupported attachment file type", ErrInvalidInput)
	ErrAttachmentContentMismatch        = fmt.Errorf("%w: attachment content type mismatch", ErrInvalidInput)
	ErrAttachmentContentEmpty           = fmt.Errorf("%w: empty attachment content", ErrInvalidInput)
	ErrAttachmentTooLarge               = fmt.Errorf("%w: attachment too large", ErrInvalidInput)
)

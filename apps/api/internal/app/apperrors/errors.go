package apperrors

import (
	"errors"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

var (
	ErrUnauthenticated = ports.ErrUnauthenticated
	ErrUnauthorized    = ports.ErrForbidden
	ErrValidation      = errors.New("validation")
	ErrConflict        = ports.ErrConflict
	ErrPrecondition    = errors.New("precondition")
	ErrNotFound        = errors.New("not_found")

	ErrInvalidInput = ErrValidation
)

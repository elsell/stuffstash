package app

import (
	"errors"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

var (
	ErrUnauthenticated = ports.ErrUnauthenticated
	ErrUnauthorized    = ports.ErrForbidden
	ErrInvalidInput    = errors.New("invalid_input")
	ErrNotFound        = errors.New("not_found")
)

package ports

import "errors"

var ErrOutboxClaimLost = errors.New("outbox claim lost")
var ErrAuthorizationOutboxClaimLost = ErrOutboxClaimLost
var ErrConflict = errors.New("conflict")
var ErrBlobNotFound = errors.New("blob not found")
var ErrDirectUploadInvalid = errors.New("direct upload invalid")
var ErrDirectUploadExpired = errors.New("direct upload expired")
var ErrDirectUploadIncomplete = errors.New("direct upload incomplete")
var ErrDirectUploadMismatch = errors.New("direct upload mismatch")

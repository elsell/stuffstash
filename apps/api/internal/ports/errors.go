package ports

import "errors"

var ErrOutboxClaimLost = errors.New("outbox claim lost")
var ErrAuthorizationOutboxClaimLost = ErrOutboxClaimLost
var ErrConflict = errors.New("conflict")
var ErrBlobNotFound = errors.New("blob not found")

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

// AssetKindConflict retains conflict classification while identifying safe user guidance.
type AssetKindConflict string

func (e AssetKindConflict) Error() string { return string(e) }
func (e AssetKindConflict) Unwrap() error { return ErrConflict }

const (
	ErrAssetKindHasContents AssetKindConflict = "Move all contents, including archived assets, before changing this asset to an item."
	ErrAssetKindHasCheckout AssetKindConflict = "Return this asset before changing it to a location."
)

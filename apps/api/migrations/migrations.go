package migrations

import "embed"

// Files contains the reviewed SQL migrations used by the stuff-stash binary.
//
//go:embed *.sql
var Files embed.FS

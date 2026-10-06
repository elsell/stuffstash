package ports

import "context"

// ImportSources submits source descriptions to the inventory server. The CLI
// does not fetch source URLs or interpret credentials itself.
type ImportSources interface {
	PreviewImportJob(context.Context, Scope, []byte) (Result[ImportJob], error)
	StartImportJob(context.Context, Scope, string, []byte) (Result[ImportJob], error)
}

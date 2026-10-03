package ports

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
)

var ErrAttemptNotFound = errors.New("print attempt not found")
var ErrRecoveryRequired = errors.New("printer has unresolved recovery evidence; reconcile before printing")

// PrintJobs is scoped to the configured authenticated API and connector. Neither
// artifact URLs nor arbitrary transport paths cross this application port.
type PrintJobs interface {
	Unsettled(context.Context, string) ([]printing.AttemptStatus, error)
	Claim(context.Context, string, printing.AttemptControl) (*printing.Claim, error)
	Artifact(context.Context, printing.AttemptControl, int64) ([]byte, string, error)
	Start(context.Context, printing.AttemptControl) (printing.AttemptStatus, error)
	Renew(context.Context, printing.AttemptControl) (printing.AttemptStatus, error)
	Outcome(context.Context, printing.AttemptControl, printing.Evidence) error
	Attempt(context.Context, string) (printing.AttemptStatus, error)
	Reconcile(context.Context, string, uint64, printing.Evidence) error
}
type PrintIdentity interface {
	Attempt() (string, error)
	ClaimToken() (string, error)
}

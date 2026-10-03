package ports

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

var ErrPrintJobNotFound = errors.New("print job not found")

// PrintJobAudit constructs the application-owned safe history record for an
// atomic state change. A failed callback or audit write rolls back the mutation.
type PrintJobAudit func(before, after printing.Job) (audit.Record, error)
type PrintJobMutation func(job *printing.Job, printer printing.Printer) error

type PrintJobCreate struct {
	Job                printing.Job
	PrinterRevision    uint64
	RequestFingerprint string
	Audit              audit.Record
}
type PrintClaim struct {
	Authority    printing.ConsumerAuthority
	Owner        printing.AttemptAuthority
	Now          time.Time
	Lease        time.Duration
	ReportMaxAge time.Duration
	Audit        PrintJobAudit
}
type PrintJobUpdate struct {
	Scope     printing.Scope
	PrinterID printing.PrinterID
	JobID     printing.JobID
	// Nil authority denotes an already-authorized human command. Consumer
	// commands recheck the registration deny fence in the same transaction.
	Authority *printing.ConsumerAuthority
	Now       time.Time
	Change    PrintJobMutation
	Audit     PrintJobAudit
}

type PrintJobRepository interface {
	CreatePrintJob(context.Context, PrintJobCreate) (printing.Job, bool, error)
	GetPrintJob(context.Context, printing.Scope, printing.JobID) (printing.Job, error)
	ListPrintJobs(context.Context, printing.Scope, printing.PrinterID, int, string) ([]printing.Job, error)
	ClaimPrintJob(context.Context, PrintClaim) (printing.Job, bool, error)
	UpdatePrintJob(context.Context, PrintJobUpdate) (printing.Job, error)
	FindPrintAttempt(context.Context, printing.Scope, printing.ConnectorID, printing.AttemptID) (printing.Job, error)
}

package ports

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
)

var (
	ErrDeviceInUse    = errors.New("printer device is already in use")
	ErrJournalMissing = errors.New("printer recovery journal is missing")
	ErrJournalCorrupt = errors.New("printer recovery journal is invalid; reconcile before printing")
)

// PrintState grants journal access only while owning the physical device lock.
type PrintState interface {
	Acquire(context.Context, string) (LockedPrintState, error)
}
type LockedPrintState interface {
	Load(context.Context) (*printing.JournalRecord, error)
	Save(context.Context, *printing.JournalRecord) error
	Close() error
}

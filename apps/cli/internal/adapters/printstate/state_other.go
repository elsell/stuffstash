//go:build !linux

package printstate

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (Store) Acquire(context.Context, string) (ports.LockedPrintState, error) {
	return nil, errors.New("USB print workers are supported only on Linux")
}

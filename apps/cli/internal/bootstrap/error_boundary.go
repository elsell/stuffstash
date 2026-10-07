package bootstrap

import (
	"errors"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// Unknown infrastructure errors can contain private paths, tokens or provider
// details. Only reviewed messages and known sentinel guidance reach the terminal.
func fallbackErrorMessage(err error) string {
	switch {
	case errors.Is(err, ports.ErrConnectorNotRegistered):
		return "The connector is not registered. Run connectors print register to pair the connector."
	case errors.Is(err, ports.ErrRecoveryRequired):
		return "The printer has an unresolved print attempt. Examine the print job and printer before you print again."
	case errors.Is(err, ports.ErrDeviceInUse):
		return "The printer is in use. Wait until the other process releases the printer."
	case errors.Is(err, ports.ErrJournalMissing):
		return "The printer recovery journal is missing. Examine unresolved print attempts before you print again."
	case errors.Is(err, ports.ErrJournalCorrupt):
		return "The printer recovery journal is not correct. Examine unresolved print attempts before you print again."
	case errors.Is(err, ports.ErrAttemptNotFound):
		return "The print attempt was not found. Examine the print job before further action."
	default:
		return "The CLI cannot complete the command. Run the command with --help to examine its options. Examine the server state before you make the change again."
	}
}

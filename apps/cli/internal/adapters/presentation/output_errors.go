package presentation

import (
	"fmt"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// Result can follow a completed mutation. Output failure must not tell the
// caller that the server rejected that mutation or expose writer internals.
func (o Output) Result(value any) error {
	if err := o.result(value); err != nil {
		return ports.Failure("output", "The CLI cannot write the result. A server change can already be complete. Examine the server state before you make the change again.")
	}
	return nil
}

func (o Output) Notice(message string) error {
	if _, err := fmt.Fprintln(o.Stderr, message); err != nil {
		return ports.Failure("output", "The CLI cannot write a diagnostic message. Examine the output destination. Examine the server state before you make the change again.")
	}
	return nil
}

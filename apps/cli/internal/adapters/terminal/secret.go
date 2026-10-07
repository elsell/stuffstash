package terminal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"golang.org/x/term"
	"io"
	"unicode/utf8"
)

func (p Picker) ReadSecret(ctx context.Context, title string, maximum int) (string, error) {
	if p.Input == nil || p.Output == nil || !term.IsTerminal(int(p.Input.Fd())) || !term.IsTerminal(int(p.Output.Fd())) {
		return "", ports.Failure("usage", "Secret input is not available. Supply --input FILE or pipe JSON through --input -.")
	}
	restore, err := prepareOutput(p.Output)
	if err != nil {
		return "", err
	}
	defer restore()
	state, err := term.MakeRaw(int(p.Input.Fd()))
	if err != nil {
		return "", ports.Failure("input", "The CLI cannot start secret input. Supply --input FILE instead.")
	}
	defer term.Restore(int(p.Input.Fd()), state)
	return runSecret(ctx, p.Input, p.Output, title, maximum)
}
func runSecret(ctx context.Context, input io.Reader, output io.Writer, title string, maximum int) (string, error) {
	limit := maximum
	if limit > 4095 {
		limit = 4095
	}
	stream := &textIO{ctx: ctx, reader: bufio.NewReader(input), writer: output, secretLimit: limit}
	terminal := term.NewTerminal(stream, "")
	for {
		value, err := terminal.ReadPassword(clean(title) + ": ")
		if errors.Is(err, io.EOF) {
			return "", ports.Failure("input", "Secret input ended before completion. Supply --input FILE instead.")
		}
		if err != nil && !errors.Is(err, term.ErrPasteIndicator) {
			return "", err
		}
		if value != "" && utf8.RuneCountInString(value) <= maximum {
			return value, nil
		}
		if _, err := fmt.Fprintf(output, "Enter 1 to %d characters.\r\n", maximum); err != nil {
			return "", err
		}
	}
}

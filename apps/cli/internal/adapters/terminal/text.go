package terminal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"golang.org/x/term"
)

func (p Picker) ReadText(ctx context.Context, title string, maximum int) (string, error) {
	if p.Input == nil || p.Output == nil || !term.IsTerminal(int(p.Input.Fd())) || !term.IsTerminal(int(p.Output.Fd())) {
		return "", ports.Failure("usage", "Interactive input is not available. Supply --name or --input.")
	}
	restoreOutput, err := prepareOutput(p.Output)
	if err != nil {
		return "", err
	}
	defer restoreOutput()
	state, err := term.MakeRaw(int(p.Input.Fd()))
	if err != nil {
		return "", ports.Failure("input", "Cannot start terminal input. Supply --name or --input.")
	}
	defer term.Restore(int(p.Input.Fd()), state)
	width, height, err := term.GetSize(int(p.Output.Fd()))
	if err != nil || width < 1 || height < 1 {
		width, height = 80, 24
	}
	return runText(ctx, p.Input, p.Output, title, maximum, width, height)
}

type textIO struct {
	secretLimit int
	secretRead  int
	ctx         context.Context
	reader      *bufio.Reader
	writer      io.Writer
	pending     []byte
}

func (t *textIO) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if len(t.pending) == 0 {
		r, err := readRune(t.ctx, t.reader)
		if err != nil {
			return 0, err
		}
		if r == 3 {
			return 0, context.Canceled
		}
		if t.secretLimit > 0 && r != '\r' && r != '\n' {
			t.secretRead++
			if t.secretRead > t.secretLimit {
				return 0, ports.Failure("input", "Secret input is too long. Use --input FILE instead.")
			}
		}
		t.pending = utf8.AppendRune(nil, r)
	}
	n := copy(p, t.pending)
	t.pending = t.pending[n:]
	return n, nil
}
func (t *textIO) Write(p []byte) (int, error) { return t.writer.Write(p) }
func runText(ctx context.Context, input io.Reader, output io.Writer, title string, maximum int, size ...int) (string, error) {
	stream := &textIO{ctx: ctx, reader: bufio.NewReader(input), writer: output}
	terminal := term.NewTerminal(stream, clean(title)+": ")
	if len(size) == 2 {
		if err := terminal.SetSize(size[0], size[1]); err != nil {
			return "", err
		}
	}
	for {
		line, err := terminal.ReadLine()
		if errors.Is(err, io.EOF) {
			return "", ports.Failure("input", "Input ended before a name was entered. Supply --name or --input.")
		}
		if err != nil && !errors.Is(err, term.ErrPasteIndicator) {
			return "", err
		}
		line = strings.TrimSpace(line)
		if line != "" && utf8.RuneCountInString(line) <= maximum {
			return line, nil
		}
		if _, err := fmt.Fprintf(output, "Enter a name with 1 to %d characters.\r\n", maximum); err != nil {
			return "", err
		}
	}
}

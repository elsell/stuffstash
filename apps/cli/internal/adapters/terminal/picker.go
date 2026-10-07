// Package terminal provides compact, keyboard-driven CLI prompts.
package terminal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"golang.org/x/term"
)

type Picker struct {
	Input, Output *os.File
	Color         bool
}

func Available(input, output, diagnostics *os.File) bool {
	return input != nil && output != nil && diagnostics != nil && term.IsTerminal(int(input.Fd())) && term.IsTerminal(int(output.Fd())) && term.IsTerminal(int(diagnostics.Fd()))
}
func (p Picker) Pick(ctx context.Context, title string, choices []ports.Choice) (string, error) {
	if p.Input == nil || p.Output == nil || !term.IsTerminal(int(p.Input.Fd())) || !term.IsTerminal(int(p.Output.Fd())) {
		return "", ports.Failure("usage", "Interactive input is not available. Use this command's --help to find the required input options.")
	}
	restoreOutput, err := prepareOutput(p.Output)
	if err != nil {
		return "", err
	}
	defer restoreOutput()
	state, err := term.MakeRaw(int(p.Input.Fd()))
	if err != nil {
		return "", ports.Failure("input", "The CLI cannot start terminal input. Use this command's --help to find the required input options.")
	}
	defer term.Restore(int(p.Input.Fd()), state)
	width, _, err := term.GetSize(int(p.Output.Fd()))
	if err != nil || width < 20 {
		width = 70
	}
	return runPicker(ctx, p.Input, p.Output, title, choices, p.Color, width-1)
}
func runPicker(ctx context.Context, input io.Reader, output io.Writer, title string, choices []ports.Choice, color bool, width int) (string, error) {
	reader := bufio.NewReader(input)
	query := []rune{}
	selected, lines := 0, 0
	chosen := ""
	defer func() {
		clear(output, lines)
		if chosen != "" {
			_, _ = fmt.Fprintf(output, "%s: %s\r\n", clean(title), clean(chosen))
		}
	}()
	for {
		matches := make([]ports.Choice, 0, len(choices))
		for _, choice := range choices {
			if strings.Contains(strings.ToLower(clean(choice.Label+" "+choice.Detail)), strings.ToLower(string(query))) {
				matches = append(matches, choice)
			}
		}
		if selected >= len(matches) {
			selected = 0
		}
		if err := clear(output, lines); err != nil {
			return "", err
		}
		frame := []string{clean(title), "Search: " + string(query)}
		start := 0
		if selected >= 6 {
			start = selected - 5
		}
		for i := start; i < len(matches) && i < start+6; i++ {
			label := "  " + clip(clean(matches[i].Label), width/2) + "  " + clip(clean(matches[i].Detail), width-width/2-4)
			if i == selected {
				label = "> " + clip(clean(matches[i].Label), width/2) + "  " + clip(clean(matches[i].Detail), width-width/2-4)
			}
			label = clip(label, width)
			if color && i == selected {
				label = "\x1b[36m" + label + "\x1b[0m"
			}
			frame = append(frame, label)
		}
		if len(matches) == 0 {
			frame = append(frame, "  No matches")
		}
		frame = append(frame, "↑/↓ move · Enter choose · Esc cancel")
		lines = 0
		for _, line := range frame {
			if _, err := fmt.Fprintf(output, "%s\r\n", clipStyled(line, width)); err != nil {
				return "", err
			}
			lines++
		}
		key, err := readRune(ctx, reader)
		if err != nil {
			return "", err
		}
		switch key {
		case 3:
			return "", context.Canceled
		case 27:
			escapeCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
			next, nextErr := readRune(escapeCtx, reader)
			if nextErr != nil || next != '[' {
				cancel()
				return "", context.Canceled
			}
			direction, directionErr := readRune(escapeCtx, reader)
			cancel()
			if directionErr != nil {
				return "", context.Canceled
			}
			if direction == 'B' && selected+1 < len(matches) {
				selected++
			}
			if direction == 'A' && selected > 0 {
				selected--
			}
		case '\r', '\n':
			if len(matches) > 0 {
				chosen = matches[selected].Label
				return matches[selected].ID, nil
			}
		case 8, 127:
			if len(query) > 0 {
				query = query[:len(query)-1]
				selected = 0
			}
		default:
			if !unicode.IsControl(key) && !unicode.Is(unicode.Cf, key) && len(query) < 128 {
				query = append(query, key)
				selected = 0
			}
		}
	}
}

// A canceled CLI invocation restores raw mode immediately and exits the command.
// At most one console read remains pending until the process closes its input.
func readRune(ctx context.Context, reader *bufio.Reader) (rune, error) {
	type result struct {
		value rune
		err   error
	}
	ready := make(chan result, 1)
	go func() { value, _, err := reader.ReadRune(); ready <- result{value, err} }()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case got := <-ready:
		if got.err != nil {
			return 0, ports.Failure("input", "Input ended before you completed the prompt. Use this command's --help to find the required input options.")
		}
		return got.value, nil
	}
}
func clear(output io.Writer, lines int) error {
	for i := 0; i < lines; i++ {
		if _, err := io.WriteString(output, "\x1b[1A\r\x1b[2K"); err != nil {
			return err
		}
	}
	return nil
}
func clean(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return ' '
		}
		return r
	}, value)
}
func clip(value string, width int) string {
	var out strings.Builder
	columns := 0
	for _, r := range value {
		size := 1
		if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) {
			size = 0
		} else if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hangul, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) || r >= 0x1f000 {
			size = 2
		}
		if columns+size > width-1 {
			out.WriteRune('…')
			break
		}
		out.WriteRune(r)
		columns += size
	}
	return out.String()
}
func clipStyled(value string, width int) string {
	if strings.HasPrefix(value, "\x1b[36m") {
		return value
	}
	return clip(value, width)
}

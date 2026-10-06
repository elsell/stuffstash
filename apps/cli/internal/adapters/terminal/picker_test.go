package terminal

import (
	"bytes"
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
	"testing"
)

func TestPickerSupportsArrowsSearchAndCancellation(t *testing.T) {
	choices := []ports.Choice{{ID: "attic", Label: "Attic"}, {ID: "garage", Label: "Garage"}}
	for _, test := range []struct {
		input, want string
		cancel      bool
	}{{"\x1b[B\r", "garage", false}, {"garx\x7f\r", "garage", false}, {"\x03", "", true}, {"\x1b", "", true}} {
		var output bytes.Buffer
		got, err := runPicker(context.Background(), strings.NewReader(test.input), &output, "Inventory", choices, false, 70)
		if test.cancel {
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("expected cancellation: %v", err)
			}
		} else if err != nil || got != test.want {
			t.Fatalf("input=%q got=%q err=%v", test.input, got, err)
		}
	}
}
func TestPickerDoesNotRenderServerControlSequences(t *testing.T) {
	var output bytes.Buffer
	_, err := runPicker(context.Background(), strings.NewReader("\r"), &output, "Inventory", []ports.Choice{{ID: "safe", Label: "Home\x1b]52;c;secret\a", Detail: "line\nnext"}}, false, 70)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "\x1b]52") || strings.Contains(output.String(), "\a") || strings.Contains(output.String(), "line\nnext") {
		t.Fatal("server content can control terminal")
	}
}

func TestDuplicateLongNamesKeepIdentifiersVisible(t *testing.T) {
	var output bytes.Buffer
	choices := []ports.Choice{{ID: "alpha", Label: strings.Repeat("Same name ", 8), Detail: "alpha"}, {ID: "beta", Label: strings.Repeat("Same name ", 8), Detail: "beta"}}
	got, err := runPicker(context.Background(), strings.NewReader("\x1b[B\r"), &output, "Inventory", choices, false, 30)
	if err != nil || got != "beta" || !strings.Contains(output.String(), "alpha") || !strings.Contains(output.String(), "beta") {
		t.Fatalf("choices not distinguishable: got=%q error=%v output=%q", got, err, output.String())
	}
}

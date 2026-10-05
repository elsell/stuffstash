package terminal

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestTextPromptUsesEditingAndCancels(t *testing.T) {
	var output bytes.Buffer
	got, err := runText(context.Background(), strings.NewReader("Garge\x1b[D\x1b[Da\r"), &output, "Inventory name", 120)
	if err != nil || got != "Garage" {
		t.Fatalf("editing: %q %v", got, err)
	}
	if _, err := runText(context.Background(), strings.NewReader("\x03"), &output, "Name", 120); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
	got, err = runText(context.Background(), strings.NewReader(" \rToo long\r Home \r"), &output, "Name", 4)
	if err != nil || got != "Home" {
		t.Fatalf("validation: %q %v", got, err)
	}
}

func TestTextPromptAcceptsPastedName(t *testing.T) {
	var output bytes.Buffer
	got, err := runText(context.Background(), strings.NewReader("\x1b[200~Home\x1b[201~\r"), &output, "Name", 120)
	if err != nil || got != "Home" {
		t.Fatalf("paste failed: %q %v", got, err)
	}
}

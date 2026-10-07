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

func TestSharedTextPromptRecoveryDoesNotInventFieldOptions(t *testing.T) {
	for _, field := range []string{"Invitee email address", "Device revision", "Asset title"} {
		var output bytes.Buffer
		_, err := runText(context.Background(), strings.NewReader(""), &output, field, 120)
		if err == nil || strings.Contains(err.Error(), "--name") || strings.Contains(err.Error(), "a name") || !strings.Contains(err.Error(), "--help") {
			t.Fatalf("wrong recovery for %s: %v", field, err)
		}
		_, err = (Picker{}).ReadText(context.Background(), field, 120)
		if err == nil || strings.Contains(err.Error(), "--name") || !strings.Contains(err.Error(), "--help") {
			t.Fatalf("wrong nonterminal recovery for %s: %v", field, err)
		}
	}
}

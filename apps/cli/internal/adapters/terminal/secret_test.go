package terminal

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSecretPromptDoesNotEchoAndCancels(t *testing.T) {
	var output bytes.Buffer
	secret := "private-token-123"
	got, err := runSecret(context.Background(), strings.NewReader(secret+"\r"), &output, "Device token", 8192)
	if err != nil || got != secret {
		t.Fatalf("secret input failed: %v", err)
	}
	if strings.Contains(output.String(), secret) {
		t.Fatal("secret echoed")
	}
	if _, err := runSecret(context.Background(), strings.NewReader("\x03"), &output, "Device token", 8192); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
}

func TestSecretPromptRejectsTruncation(t *testing.T) {
	var output bytes.Buffer
	got, err := runSecret(context.Background(), strings.NewReader(strings.Repeat("x", 4097)+"\r"), &output, "Token", 8192)
	if err == nil || got != "" {
		t.Fatal("overlong secret returned as a partial token")
	}
	if strings.Contains(output.String(), "xxxx") {
		t.Fatal("secret echoed")
	}
}

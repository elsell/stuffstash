package identity

import (
	"strings"
	"testing"
)

func TestNewDisplayNameNormalizesSafeNames(t *testing.T) {
	name, ok := NewDisplayName("  Alex Rivera  ")
	if !ok || name.String() != "Alex Rivera" {
		t.Fatalf("expected trimmed display name, got %q, %t", name.String(), ok)
	}
}

func TestNewDisplayNameRejectsUnsafeBoundaries(t *testing.T) {
	for _, value := range []string{"", "Alex\nRivera", strings.Repeat("a", 121)} {
		if name, ok := NewDisplayName(value); ok {
			t.Fatalf("expected %q to be rejected, got %q", value, name.String())
		}
	}
	if _, ok := NewDisplayName(strings.Repeat("a", 120)); !ok {
		t.Fatal("expected a 120-rune display name to be accepted")
	}
}

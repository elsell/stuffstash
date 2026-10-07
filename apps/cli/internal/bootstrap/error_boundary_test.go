package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestErrorBoundaryHidesUnexpectedDetailsAndPreservesContracts(t *testing.T) {
	for _, tc := range []struct {
		name           string
		err            error
		exit           int
		category, hint string
	}{
		{"unexpected", errors.New("private-key-path-and-token"), 2, "configuration", "before you make the change again"},
		{"sign in", fmt.Errorf("private-key-path-and-token: %w", ports.ErrNotLoggedIn), 1, "authentication", "stuffstash login"},
		{"registration", fmt.Errorf("private-key-path-and-token: %w", ports.ErrConnectorNotRegistered), 2, "configuration", "connectors print register"},
		{"journal", fmt.Errorf("private-key-path-and-token: %w", ports.ErrJournalCorrupt), 2, "configuration", "before you print again"},
		{"canceled", fmt.Errorf("private-key-path-and-token: %w", context.Canceled), 130, "canceled", "canceled"},
		{"typed", ports.Failure("network", "Reviewed network guidance."), 1, "network", "Reviewed network guidance."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var diagnostic bytes.Buffer
			code := exit(presentation.Output{Stdout: io.Discard, Stderr: &diagnostic, JSON: true}, tc.err)
			var value struct {
				Error struct{ Category, Message string }
			}
			if err := json.Unmarshal(diagnostic.Bytes(), &value); err != nil {
				t.Fatal(err)
			}
			if code != tc.exit || value.Error.Category != tc.category || !strings.Contains(value.Error.Message, tc.hint) {
				t.Fatalf("changed error contract: exit=%d %s", code, &diagnostic)
			}
			if strings.Contains(diagnostic.String(), "private-key-path-and-token") {
				t.Fatal("unexpected error details escaped the boundary")
			}
		})
	}
}

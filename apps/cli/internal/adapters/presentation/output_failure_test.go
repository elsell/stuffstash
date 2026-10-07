package presentation

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type rejectedNoticeWriter struct{}

func (rejectedNoticeWriter) Write([]byte) (int, error) { return 0, errors.New("private-notice-path") }
func TestNoticeFailureDoesNotClaimMutationSucceeded(t *testing.T) {
	output := Output{Stdout: io.Discard, Stderr: rejectedNoticeWriter{}}
	err := output.Notice("Confirm the change")
	var failure *ports.Error
	if !errors.As(err, &failure) || failure.Category != "output" {
		t.Fatalf("unclassified output failure: %v", err)
	}
	if strings.Contains(failure.Message, "private-notice") || strings.Contains(failure.Message, "change is complete") {
		t.Fatalf("unsafe notice guidance: %v", err)
	}
}

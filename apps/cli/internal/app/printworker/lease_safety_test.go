package printworker_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type safetyMarginPrinter struct {
	*printer
	safety time.Duration
}

func (p safetyMarginPrinter) Submit(ctx context.Context, label printing.Label) (printing.Submission, error) {
	submission, err := p.printer.Submit(ctx, label)
	if err == nil {
		p.api.clock.now = p.api.status.LeaseExpiresAt.Add(-p.safety)
	}
	return submission, err
}

func TestLeaseSafetyAfterSubmissionPreservesUncertainOutput(t *testing.T) {
	w, a, j, p := fixture(t)
	device := safetyMarginPrinter{printer: p, safety: w.Config.LeaseSafety}
	err := w.Step(context.Background(), j, device)
	var failure *ports.Error
	if !errors.As(err, &failure) || failure.Category != "lease_expired" {
		t.Fatalf("expected lease safety rejection, got %v", err)
	}
	if !a.clock.Now().Before(a.status.LeaseExpiresAt) {
		t.Fatal("test must reach safety margin before actual expiry")
	}
	if p.submissions != 1 || j.record == nil || j.record.Phase != printing.JournalSubmitting || j.record.CompletedCopies != 0 {
		t.Fatal("submission evidence was lost or output was repeated")
	}
	if err := w.Step(context.Background(), j, device); !errors.Is(err, ports.ErrRecoveryRequired) {
		t.Fatalf("expected recovery instead of another print, got %v", err)
	}
	if p.submissions != 1 || j.record == nil || j.record.Phase != printing.JournalSubmitting {
		t.Fatal("recovery repeated output or cleared uncertain evidence")
	}
	if strings.Contains(failure.Message, "no output") || strings.Contains(failure.Message, "expired") {
		t.Fatalf("unsafe output assurance: %s", failure.Message)
	}
}

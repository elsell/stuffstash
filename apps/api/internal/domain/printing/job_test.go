package printing

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"
)

func TestExpiredStartedPrintCannotBeReclaimedOrBlindlyRetried(t *testing.T) {
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	job := Job{ID: "job", PrinterID: "printer", Status: JobQueued, Revision: 1, Copies: 1}
	owner := AttemptAuthority{AttemptID: "attempt", ConnectorID: "connector", SessionID: "session", TokenDigest: sha256.Sum256([]byte("secret"))}
	if err := job.Claim("attempt", owner, now, time.Minute, 1); err != nil {
		t.Fatal(err)
	}
	stranger := owner
	stranger.SessionID = "other-session"
	if err := job.Start(stranger, now, job.Revision); !errors.Is(err, ErrAttemptOwnership) {
		t.Fatalf("foreign start: %v", err)
	}
	if err := job.Start(owner, now, job.Revision); err != nil {
		t.Fatal(err)
	}
	if !job.Expire(now.Add(time.Minute)) || job.Status != JobUncertain {
		t.Fatal("started expired attempt must remain uncertain")
	}
	if err := job.Claim("second", owner, now.Add(time.Minute), time.Minute, job.Revision); !errors.Is(err, ErrJobConflict) {
		t.Fatalf("uncertain job reclaimed: %v", err)
	}
	if err := job.Report(owner, now.Add(time.Minute), job.Revision, Outcome{Kind: OutcomeCompleted, CompletedCopies: 1}); !errors.Is(err, ErrLeaseExpired) {
		t.Fatalf("expired normal report: %v", err)
	}
	if err := job.Reconcile("attempt", "connector", Outcome{Kind: OutcomeCompleted, CompletedCopies: 1}, now.Add(2*time.Minute), job.Revision); err != nil {
		t.Fatal(err)
	}
	revision := job.Revision
	if err := job.Reconcile("attempt", "connector", Outcome{Kind: OutcomeCompleted, CompletedCopies: 1}, now.Add(3*time.Minute), revision); err != nil || job.Revision != revision {
		t.Fatal("identical outcome must be idempotent")
	}
}

func TestPreStartRecoveryAndExplicitNoOutputPermitSafeRetry(t *testing.T) {
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	owner := AttemptAuthority{AttemptID: "first", ConnectorID: "connector", SessionID: "session", TokenDigest: sha256.Sum256([]byte("secret"))}
	job := Job{ID: "job", PrinterID: "printer", Status: JobQueued, Revision: 1, Copies: 2}
	if err := job.Claim("first", owner, now, time.Minute, 1); err != nil {
		t.Fatal(err)
	}
	if !job.Expire(now.Add(time.Minute)) || job.Status != JobQueued {
		t.Fatal("expired unstarted claim should requeue")
	}
	owner.AttemptID = "second"
	if err := job.Claim("second", owner, now.Add(time.Minute), time.Minute, job.Revision); err != nil {
		t.Fatal(err)
	}
	if err := job.Start(owner, now.Add(time.Minute), job.Revision); err != nil {
		t.Fatal(err)
	}
	if err := job.Report(owner, now.Add(time.Minute), job.Revision, Outcome{Kind: OutcomeNoOutput, Retryable: true}); err != nil || job.Status != JobQueued {
		t.Fatal("authoritative zero-output transient failure should requeue")
	}
	if len(job.Attempts) != 2 {
		t.Fatal("retry must preserve attempt history")
	}
	owner.AttemptID = "third"
	if err := job.Claim("third", owner, now.Add(time.Minute), time.Minute, job.Revision); err != nil {
		t.Fatal(err)
	}
	if err := job.Cancel(now.Add(time.Minute), job.Revision); err != nil {
		t.Fatal(err)
	}
	if err := job.Start(owner, now.Add(time.Minute), job.Revision); !errors.Is(err, ErrJobConflict) {
		t.Fatalf("cancellation failed to fence start: %v", err)
	}
}

func TestKnownPartialOutputCannotBecomeAutomaticRetry(t *testing.T) {
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	owner := AttemptAuthority{AttemptID: "attempt", ConnectorID: "connector", SessionID: "session", TokenDigest: sha256.Sum256([]byte("secret"))}
	job := Job{ID: "job", PrinterID: "printer", Status: JobQueued, Revision: 1, Copies: 2}
	if err := job.Claim("attempt", owner, now, time.Minute, 1); err != nil {
		t.Fatal(err)
	}
	if err := job.Start(owner, now, job.Revision); err != nil {
		t.Fatal(err)
	}
	if err := job.Report(owner, now, job.Revision, Outcome{Kind: OutcomeUncertain, CompletedCopies: 1, Reason: ReasonPartialOutput}); err != nil {
		t.Fatal(err)
	}
	revision := job.Revision
	if err := job.Report(owner, now.Add(2*time.Minute), revision-1, Outcome{Kind: OutcomeUncertain, CompletedCopies: 1, Reason: ReasonPartialOutput}); err != nil || job.Revision != revision {
		t.Fatal("lost uncertain response must retry idempotently")
	}
	if err := job.Reconcile("attempt", "connector", Outcome{Kind: OutcomeNoOutput, Retryable: true}, now, job.Revision); !errors.Is(err, ErrInvalidOutcome) {
		t.Fatalf("partial print requeued: %v", err)
	}
	if job.Status != JobUncertain || !job.HoldsReservation() {
		t.Fatal("partial output must retain uncertainty and reservation")
	}
}

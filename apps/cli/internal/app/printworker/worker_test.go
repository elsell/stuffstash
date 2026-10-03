package printworker_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/app/printworker"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

type waiter struct{}

func (waiter) Wait(ctx context.Context, _ time.Duration) error { return ctx.Err() }

type identities struct{ sequence int }

func (i *identities) Attempt() (string, error) {
	i.sequence++
	return fmt.Sprintf("attempt-%d", i.sequence), nil
}
func (i *identities) ClaimToken() (string, error) {
	sum := sha256.Sum256([]byte(fmt.Sprintf("test-claim-%d", i.sequence)))
	return base64.RawURLEncoding.EncodeToString(sum[:]), nil
}

type observer struct{}

func (observer) Event(context.Context, string) {}

// journal represents acknowledged durable storage; rejected writes never replace
// its recoverable record. Printer output consults that committed state.
type journal struct {
	record *printing.JournalRecord
	fail   printing.JournalPhase
}

func (j *journal) Load(context.Context) (*printing.JournalRecord, error) {
	if j.record == nil {
		return nil, nil
	}
	copy := *j.record
	return &copy, nil
}
func (j *journal) Save(_ context.Context, value *printing.JournalRecord) error {
	if value != nil && value.Phase == j.fail {
		return io.ErrClosedPipe
	}
	if value == nil {
		j.record = nil
	} else {
		copy := *value
		j.record = &copy
	}
	return nil
}
func (j *journal) Close() error { return nil }

type api struct {
	idleConfirmed          bool
	clock                  *clock
	claim                  printing.Claim
	status                 *printing.AttemptStatus
	bytes                  []byte
	dropStart, dropOutcome bool
	evidence               printing.Evidence
	claims                 int
}

func (a *api) Unsettled(context.Context, string) ([]printing.AttemptStatus, error) {
	if a.status == nil || a.status.Phase == printing.RemoteCompleted || a.status.Phase == printing.RemoteFailed {
		return nil, nil
	}
	return []printing.AttemptStatus{*a.status}, nil
}
func (a *api) Claim(_ context.Context, _ string, control printing.AttemptControl) (*printing.Claim, error) {
	if a.status != nil && a.status.Phase != printing.RemoteCompleted && a.status.Phase != printing.RemoteFailed {
		return nil, errors.New("printer reserved")
	}
	a.claims++
	a.idleConfirmed = false
	a.claim.Control = control
	a.claim.Control.Revision = 1
	a.claim.LeaseExpiresAt = a.clock.now.Add(time.Minute)
	a.status = &printing.AttemptStatus{AttemptID: control.AttemptID, SessionID: control.SessionID, JobID: a.claim.JobID, Phase: printing.RemoteClaimed, Revision: 1, LeaseExpiresAt: a.claim.LeaseExpiresAt}
	copy := a.claim
	return &copy, nil
}
func (a *api) owned(control printing.AttemptControl) bool {
	return a.status != nil && control.AttemptID == a.status.AttemptID && control.SessionID == a.status.SessionID && control.ClaimToken == a.claim.Control.ClaimToken && control.Revision == a.status.Revision && a.status.LeaseExpiresAt.After(a.clock.Now())
}
func (a *api) Artifact(_ context.Context, control printing.AttemptControl, maximum int64) ([]byte, string, error) {
	if !a.owned(control) || int64(len(a.bytes)) > maximum {
		return nil, "", errors.New("artifact unavailable")
	}
	return append([]byte(nil), a.bytes...), "image/png", nil
}
func (a *api) Start(_ context.Context, control printing.AttemptControl) (printing.AttemptStatus, error) {
	if !a.owned(control) || a.status.Phase != printing.RemoteClaimed {
		return printing.AttemptStatus{}, errors.New("start conflict")
	}
	a.status.Phase = printing.RemotePrinting
	a.status.Revision++
	if a.dropStart {
		a.dropStart = false
		return printing.AttemptStatus{}, io.ErrUnexpectedEOF
	}
	return *a.status, nil
}
func (a *api) Renew(_ context.Context, control printing.AttemptControl) (printing.AttemptStatus, error) {
	if !a.owned(control) || (a.status.Phase != printing.RemoteClaimed && a.status.Phase != printing.RemotePrinting) {
		return printing.AttemptStatus{}, errors.New("lease conflict")
	}
	a.status.Revision++
	a.status.LeaseExpiresAt = a.clock.now.Add(time.Minute)
	return *a.status, nil
}
func (a *api) Outcome(_ context.Context, control printing.AttemptControl, evidence printing.Evidence) error {
	if !a.owned(control) {
		return errors.New("outcome conflict")
	}
	a.evidence = evidence
	a.status.Outcome = evidence.Outcome
	a.status.CompletedCopies = evidence.CompletedCopies
	a.status.Revision++
	if evidence.Outcome == printing.Completed {
		a.status.Phase = printing.RemoteCompleted
	} else if evidence.Outcome == printing.NoOutput {
		a.status.Phase = printing.RemoteFailed
	} else {
		a.status.Phase = printing.RemoteUncertain
	}
	if a.dropOutcome {
		a.dropOutcome = false
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (a *api) Attempt(_ context.Context, attemptID string) (printing.AttemptStatus, error) {
	if a.status == nil || a.status.AttemptID != attemptID {
		return printing.AttemptStatus{}, ports.ErrAttemptNotFound
	}
	return *a.status, nil
}
func (a *api) ConfirmIdle(_ context.Context, id string, revision uint64) error {
	if a.status == nil || a.status.AttemptID != id || a.status.Phase != printing.RemoteUncertain {
		return errors.New("idle confirmation conflict")
	}
	if a.idleConfirmed {
		return nil
	}
	if a.status.Revision != revision {
		return errors.New("idle confirmation conflict")
	}
	a.status.Revision++
	a.idleConfirmed = true
	return nil
}
func (a *api) Reconcile(_ context.Context, _ string, revision uint64, evidence printing.Evidence) error {
	if a.status.Phase != printing.RemoteUncertain || revision != a.status.Revision {
		return errors.New("reconciliation conflict")
	}
	a.evidence = evidence
	a.status.Outcome = evidence.Outcome
	a.status.CompletedCopies = evidence.CompletedCopies
	a.status.Revision++
	if evidence.Outcome == printing.Completed {
		a.status.Phase = printing.RemoteCompleted
	} else if evidence.Outcome == printing.NoOutput {
		a.status.Phase = printing.RemoteFailed
	}
	return nil
}

type printer struct {
	api         *api
	journal     *journal
	submissions int
	uncertainAt int
	active      *printing.Submission
	ready       bool
}

func (p *printer) ConfirmIdle(context.Context) error {
	if p.active != nil || !p.ready {
		return ports.ErrRecoveryRequired
	}
	return nil
}
func (p *printer) Readiness(context.Context) (printing.Readiness, error) {
	p.ready = true
	return printing.Readiness{State: printing.Ready}, nil
}
func (p *printer) Submit(_ context.Context, label printing.Label) (printing.Submission, error) {
	if !p.ready || p.api.status.Phase != printing.RemotePrinting || p.journal.record == nil || p.journal.record.Phase != printing.JournalSubmitting || p.journal.record.Copy != label.Copy {
		return printing.Submission{}, errors.New("output bypassed durable intent or API start")
	}
	p.ready = false
	p.submissions++
	if p.submissions == p.uncertainAt {
		return printing.Submission{}, &printing.SubmissionError{Outcome: printing.Uncertain, Reason: printing.Disconnected}
	}
	submission := printing.Submission{Reference: fmt.Sprintf("hardware-%d", p.submissions), AttemptID: label.AttemptID, Copy: label.Copy}
	p.active = &submission
	return submission, nil
}
func (p *printer) Observe(_ context.Context, submission printing.Submission) (printing.Observation, error) {
	if p.active == nil || *p.active != submission {
		return printing.Observation{}, errors.New("unknown hardware submission")
	}
	p.active = nil
	return printing.Observation{Outcome: printing.Completed}, nil
}
func (p *printer) Close() error { return nil }

func fixture(t *testing.T) (*printworker.Worker, *api, *journal, *printer) {
	t.Helper()
	c := &clock{time.Unix(100, 0)}
	j := &journal{}
	var body bytes.Buffer
	img := image.NewGray(image.Rect(0, 0, 3, 3))
	img.SetGray(1, 1, color.Gray{Y: 255})
	if err := png.Encode(&body, img); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body.Bytes())
	media := printing.Media{PresetID: "test-media", Version: 1, RasterWidth: 3, RasterHeight: 3}
	a := &api{clock: c, bytes: body.Bytes(), claim: printing.Claim{JobID: "job", PrinterID: "printer", ContractVersion: 1, Copies: 2, Media: media, Artifact: printing.Artifact{SHA256: hex.EncodeToString(sum[:]), ContentType: "image/png", ByteLength: int64(body.Len()), Width: 3, Height: 3}}}
	p := &printer{api: a, journal: j}
	w := &printworker.Worker{Jobs: a, Clock: c, Waiter: waiter{}, Identity: &identities{}, Observer: observer{}, Config: printworker.Config{Binding: "binding", PrinterID: "printer", SessionID: "process", Media: media, MaxArtifactBytes: 1 << 20, LeaseSafety: time.Second, ObserveInterval: time.Millisecond, ReadinessTimeout: time.Second}}
	return w, a, j, p
}

func TestNoBytesWithoutAcknowledgedStartAndDurableSubmissionIntent(t *testing.T) {
	for _, test := range []string{"start-reply-lost", "journal-fsync-failed"} {
		t.Run(test, func(t *testing.T) {
			w, a, j, p := fixture(t)
			if test == "start-reply-lost" {
				a.dropStart = true
			} else {
				j.fail = printing.JournalSubmitting
			}
			if err := w.Step(context.Background(), j, p); err == nil {
				t.Fatal("failure hidden")
			}
			if p.submissions != 0 {
				t.Fatal("unacknowledged/unrecorded physical output")
			}
			if j.record == nil {
				t.Fatal("lost recovery evidence")
			}
		})
	}
}
func TestLostOutcomeResponseRecoversWithoutRepeatingCompletedCopies(t *testing.T) {
	w, a, j, p := fixture(t)
	a.dropOutcome = true
	if err := w.Step(context.Background(), j, p); err == nil {
		t.Fatal("lost reply hidden")
	}
	if p.submissions != 2 || j.record == nil || j.record.Phase != printing.JournalCompleted {
		t.Fatal("completion evidence lost")
	}
	w.Config.SessionID = "restarted-process"
	if err := w.Step(context.Background(), j, p); err != nil {
		t.Fatal(err)
	}
	if p.submissions != 2 || j.record != nil || a.claims != 1 {
		t.Fatal("recovery replayed output or claimed before returning")
	}
}
func TestPartialOutputAndRestartedSubmissionStayUncertain(t *testing.T) {
	w, a, j, p := fixture(t)
	p.uncertainAt = 2
	if err := w.Step(context.Background(), j, p); !errors.Is(err, ports.ErrRecoveryRequired) {
		t.Fatalf("uncertainty not surfaced: %v", err)
	}
	if a.evidence.Outcome != printing.Uncertain || a.evidence.CompletedCopies != 1 {
		t.Fatal("partial output became safe to retry")
	}
	w.Config.SessionID = "restarted-process"
	if err := w.Step(context.Background(), j, p); !errors.Is(err, ports.ErrRecoveryRequired) {
		t.Fatalf("recovery uncertainty: %v", err)
	}
	if p.submissions != 2 || a.claims != 1 {
		t.Fatal("replayed ambiguous physical attempt")
	}
}

func TestCorruptArtifactAndUnsettledAttemptPreventOutput(t *testing.T) {
	for _, scenario := range []string{"artifact-changed", "journal-missing"} {
		t.Run(scenario, func(t *testing.T) {
			w, a, j, p := fixture(t)
			if scenario == "artifact-changed" {
				a.bytes[len(a.bytes)-1] ^= 1
			} else {
				a.status = &printing.AttemptStatus{AttemptID: "prior-attempt", Phase: printing.RemoteUncertain}
			}
			if err := w.Step(context.Background(), j, p); err == nil {
				t.Fatal("unsafe print accepted")
			}
			if p.submissions != 0 {
				t.Fatal("unsafe output")
			}
			if scenario == "journal-missing" && a.claims != 0 {
				t.Fatal("claimed despite unresolved prior attempt")
			}
		})
	}
}

func TestUncertainRecoveryAttestsIdleWithoutReplayingAndWaitsForHuman(t *testing.T) {
	w, a, j, p := fixture(t)
	p.uncertainAt = 1
	if err := w.Step(context.Background(), j, p); err == nil {
		t.Fatal("expected uncertain submission")
	}
	if a.status.Phase != printing.RemoteUncertain {
		t.Fatal("missing uncertain status")
	}
	if err := w.Step(context.Background(), j, p); !errors.Is(err, ports.ErrRecoveryRequired) {
		t.Fatal(err)
	}
	if a.idleConfirmed {
		t.Fatal("unavailable device attested")
	}
	// A fresh locked connection observes physical idle after the ambiguous write.
	recoveredDevice := &printer{api: a, journal: j, ready: true}
	if err := w.Step(context.Background(), j, recoveredDevice); !errors.Is(err, ports.ErrRecoveryRequired) {
		t.Fatal(err)
	}
	if !a.idleConfirmed || j.record == nil || recoveredDevice.submissions != 0 {
		t.Fatal("idle confirmation lost journal or replayed")
	}
	a.status.Phase = printing.RemoteFailed // Authorized server-side human resolution.
	if err := w.Step(context.Background(), j, recoveredDevice); err != nil {
		t.Fatal(err)
	}
	if j.record != nil || recoveredDevice.submissions != 0 {
		t.Fatal("resolution recovery replayed output")
	}
}

func TestMissingJournalConfirmsScopedUncertaintyWithoutInventingEvidence(t *testing.T) {
	w, a, j, p := fixture(t)
	p.uncertainAt = 1
	if err := w.Step(context.Background(), j, p); err == nil {
		t.Fatal("expected ambiguous output")
	}
	claims := a.claims
	originalEvidence := a.evidence
	j.record = nil // Lost local state; the server still owns the uncertain attempt.
	if err := w.Step(context.Background(), j, nil); !errors.Is(err, ports.ErrRecoveryRequired) {
		t.Fatal(err)
	}
	if a.idleConfirmed {
		t.Fatal("confirmed without device lock/connection")
	}
	recoveredDevice := &printer{api: a, journal: j, ready: true}
	if err := w.Step(context.Background(), j, recoveredDevice); !errors.Is(err, ports.ErrRecoveryRequired) {
		t.Fatal(err)
	}
	if !a.idleConfirmed || a.evidence != originalEvidence || a.claims != claims || j.record != nil || recoveredDevice.submissions != 0 {
		t.Fatal("lost journal prevented idle proof or fabricated evidence/output")
	}
	a.status.Phase = printing.RemoteFailed // Human acknowledgement releases the queue.
	w.Config.RecoveryOnly = true
	if err := w.Step(context.Background(), j, recoveredDevice); err != nil {
		t.Fatal("resolved attempt still blocked", err)
	}
	if a.claims != claims || recoveredDevice.submissions != 0 {
		t.Fatal("recovery itself printed")
	}
	w.Config.RecoveryOnly = false
	if err := w.Step(context.Background(), j, recoveredDevice); err != nil {
		t.Fatal("subsequent queued work remained blocked", err)
	}
	if a.claims != claims+1 {
		t.Fatal("next job was not consumed")
	}
}

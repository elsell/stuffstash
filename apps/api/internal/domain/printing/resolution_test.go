package printing

import (
	"testing"
	"time"
)

func TestHumanResolutionRequiresServingConnectorIdleAndPreservesUncertainty(t *testing.T) {
	now := time.Now().UTC()
	j := Job{ID: "job", Status: JobUncertain, Revision: 4, Attempts: []Attempt{{ID: "attempt", Authority: AttemptAuthority{ConnectorID: "connector"}, Outcome: Outcome{Kind: OutcomeUncertain, Reason: ReasonPartialOutput, CompletedCopies: 1}}}}
	if j.Resolve("human", ReportedUnknown, true, now, j.Revision) == nil {
		t.Fatal("resolved without idle evidence")
	}
	if j.ConfirmIdle("attempt", "other", now, j.Revision) == nil {
		t.Fatal("foreign connector attested")
	}
	if err := j.ConfirmIdle("attempt", "connector", now, j.Revision); err != nil {
		t.Fatal(err)
	}
	if !j.HoldsReservation() {
		t.Fatal("idle confirmation released reservation")
	}
	revision := j.Revision
	if err := j.Resolve("human", ReportedUnknown, true, now, revision); err != nil {
		t.Fatal(err)
	}
	if j.Status != JobFailed || j.HoldsReservation() || j.Attempts[0].Outcome.Kind != OutcomeUncertain || j.Attempts[0].Outcome.CompletedCopies != 1 {
		t.Fatal("human report rewrote physical evidence")
	}
	if err := j.Resolve("human", ReportedUnknown, true, now, revision); err != nil {
		t.Fatal("lost-response retry", err)
	}
	if j.Resolve("human", ReportedPrinted, true, now, revision) == nil {
		t.Fatal("changed resolution accepted")
	}
	if j.Reconcile("attempt", "connector", Outcome{Kind: OutcomeCompleted, CompletedCopies: 1}, now, j.Revision) == nil {
		t.Fatal("late reconciliation overwrote human resolution")
	}
}

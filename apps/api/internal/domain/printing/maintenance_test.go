package printing

import (
	"testing"
	"time"
)

func TestMaintenanceNeverTreatsPossibleOutputAsSafeToRetry(t *testing.T) {
	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	for _, status := range []JobStatus{JobQueued, JobClaimed, JobPrinting, JobUncertain} {
		j := Job{Status: status, Revision: 1, Artifact: Artifact{ExpiresAt: now}, Attempts: []Attempt{{LeaseExpiresAt: now}}}
		j.Maintain(now)
		if status == JobPrinting || status == JobUncertain {
			if j.Status != JobUncertain || !j.HoldsReservation() || j.Terminal() {
				t.Fatalf("possible output released: %+v", j)
			}
		} else if j.Status != JobFailed || !j.Terminal() {
			t.Fatalf("expired unstarted artifact reclaimable: %+v", j)
		}
	}
}

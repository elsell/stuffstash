package printing

import "time"

func (j Job) Terminal() bool {
	return j.Status == JobCompleted || j.Status == JobFailed || j.Status == JobCanceled
}

// Maintain never converts possible physical output into a retryable outcome.
func (j *Job) Maintain(now time.Time) {
	j.Expire(now)
	if !j.Artifact.ExpiresAt.IsZero() && !j.Artifact.ExpiresAt.After(now) && (j.Status == JobQueued || j.Status == JobClaimed) {
		if j.Status == JobClaimed {
			j.apply(Outcome{Kind: OutcomeNoOutput, Reason: ReasonInvalidArtifact}, now)
		} else {
			j.Status = JobFailed
			j.changed(now)
		}
	}
}

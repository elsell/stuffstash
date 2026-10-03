package printing

import "time"

// DeleteScope preserves any possible physical output; it never makes started
// work retryable merely because the scope or credentials have been removed.
func (j *Job) DeleteScope(now time.Time) bool {
	switch j.Status {
	case JobQueued, JobClaimed:
		return j.Cancel(now, j.Revision) == nil
	case JobPrinting:
		if attempt := j.current(); attempt != nil {
			outcome := attempt.Outcome
			outcome.Kind = OutcomeUncertain
			outcome.Retryable = false
			if outcome.Reason == "" {
				outcome.Reason = ReasonUnknown
			}
			j.apply(outcome, now)
		} else {
			j.Status = JobUncertain
			j.changed(now)
		}
		return true
	default:
		return false
	}
}

func (c *Connector) RevokeDeletedScope(now time.Time) bool {
	if c.State == ConnectorRevoked {
		return false
	}
	c.State = ConnectorRevoked
	c.CredentialHash = ""
	c.PendingCredentialHash = ""
	c.PendingCredentialVersion = 0
	c.PendingPublicKey = nil
	c.PendingCredentialExpiresAt = time.Time{}
	c.PendingActivationDeadline = time.Time{}
	c.CredentialVersion++
	c.Generation++
	c.UpdatedAt = now
	return true
}

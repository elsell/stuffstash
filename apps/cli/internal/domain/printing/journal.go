package printing

import "errors"

type JournalPhase string

const (
	JournalClaiming   JournalPhase = "claiming"
	JournalPrepared   JournalPhase = "prepared"
	JournalStarted    JournalPhase = "started"
	JournalSubmitting JournalPhase = "submitting"
	JournalCompleted  JournalPhase = "completed"
	JournalUncertain  JournalPhase = "uncertain"
	JournalNoOutput   JournalPhase = "no_output"
)

// JournalRecord is evidence, never authorization to replay physical output.
// A claiming record is durable before the API can grant its attempt.
type JournalRecord struct {
	Binding         string       `json:"binding"`
	Revision        uint64       `json:"revision"`
	AttemptID       string       `json:"attemptId"`
	JobID           string       `json:"jobId,omitempty"`
	SessionID       string       `json:"sessionId"`
	ClaimToken      string       `json:"claimToken"`
	Phase           JournalPhase `json:"phase"`
	Copy            int          `json:"copy"`
	CompletedCopies int          `json:"completedCopies"`
}

func (r JournalRecord) Validate() error {
	if r.Binding == "" || r.AttemptID == "" || r.SessionID == "" || r.ClaimToken == "" || r.CompletedCopies < 0 || r.Copy < 0 {
		return errors.New("invalid attempt journal")
	}
	switch r.Phase {
	case JournalClaiming:
		if r.JobID != "" || r.Copy != 0 || r.CompletedCopies != 0 {
			return errors.New("invalid unclaimed journal")
		}
	case JournalPrepared, JournalStarted, JournalSubmitting, JournalCompleted, JournalUncertain, JournalNoOutput:
		if r.JobID == "" {
			return errors.New("journal requires a job identity")
		}
	default:
		return errors.New("invalid journal phase")
	}
	if r.Phase == JournalSubmitting && r.Copy != r.CompletedCopies+1 {
		return errors.New("invalid copy submission evidence")
	}
	return nil
}

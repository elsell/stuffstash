package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printjobs/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"time"
)

func optionalTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
func Job(j printing.Job) dto.PrintJob {
	attempts := make([]dto.PrintJobAttempt, 0, len(j.Attempts))
	for _, a := range j.Attempts {
		attempts = append(attempts, dto.PrintJobAttempt{ID: string(a.ID), IdleConfirmedAt: optionalTime(a.IdleConfirmedAt), ConnectorID: string(a.Authority.ConnectorID), ClaimedAt: a.ClaimedAt, LeaseExpiresAt: a.LeaseExpiresAt, StartedAt: optionalTime(a.StartedAt), SettledAt: optionalTime(a.SettledAt), Outcome: string(a.Outcome.Kind), Reason: string(a.Outcome.Reason), CompletedCopies: a.Outcome.CompletedCopies})
	}
	var resolution *dto.PrintJobResolution
	if j.Resolution != nil {
		resolution = &dto.PrintJobResolution{ReportedOutcome: string(j.Resolution.ReportedOutcome), ResolvedBy: j.Resolution.ResolvedBy, ResolvedAt: j.Resolution.ResolvedAt}
	}
	return dto.PrintJob{Resolution: resolution, ID: string(j.ID), Predecessor: string(j.Predecessor), PrinterID: string(j.PrinterID), AssetID: j.AssetID, Kind: string(j.Kind), Status: string(j.Status), Revision: j.Revision, Copies: j.Copies, MediaFingerprint: j.MediaFingerprint, RequestedBy: j.RequestedBy, CreatedAt: j.CreatedAt, UpdatedAt: j.UpdatedAt, Attempts: attempts}
}
func Jobs(values []printing.Job) []dto.PrintJob {
	out := make([]dto.PrintJob, 0, len(values))
	for _, j := range values {
		out = append(out, Job(j))
	}
	return out
}

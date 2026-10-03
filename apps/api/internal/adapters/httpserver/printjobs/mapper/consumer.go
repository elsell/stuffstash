package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printjobs/dto"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"time"
)

func ConsumerAttempt(j p.Job, id p.AttemptID, now time.Time, artifact bool) *dto.PrintConsumerAttempt {
	if j.ID == "" {
		return nil
	}
	var a p.Attempt
	for _, attempt := range j.Attempts {
		if attempt.ID == id {
			a = attempt
			break
		}
	}
	result := &dto.PrintConsumerAttempt{ProtocolVersion: 1, JobID: string(j.ID), PrinterID: string(j.PrinterID), AttemptID: string(a.ID), SessionID: string(a.Authority.SessionID), Status: string(j.Status), Revision: j.Revision, Copies: j.Copies, LeaseExpiresAt: a.LeaseExpiresAt, LeaseValid: a.LeaseExpiresAt.After(now) && (j.Status == p.JobClaimed || j.Status == p.JobPrinting) && len(j.Attempts) > 0 && j.Attempts[len(j.Attempts)-1].ID == a.ID, Outcome: dto.PrintOutcome{Kind: string(a.Outcome.Kind), Reason: string(a.Outcome.Reason), CompletedCopies: a.Outcome.CompletedCopies, Retryable: a.Outcome.Retryable}, MediaFingerprint: j.MediaFingerprint}
	if !a.StartedAt.IsZero() {
		result.StartedAt = &a.StartedAt
	}
	if !a.SettledAt.IsZero() {
		result.SettledAt = &a.SettledAt
	}
	if artifact {
		m := j.Media
		result.Media = &dto.PrintConsumerMedia{MarginsMicrometers: dto.PrintMediaMargins{Left: m.MarginsMicrometers.Left, Right: m.MarginsMicrometers.Right, Top: m.MarginsMicrometers.Top, Bottom: m.MarginsMicrometers.Bottom}, DisplayRotation: m.DisplayRotation, PresetID: m.PresetID, Version: m.Version, WidthMicrometers: m.WidthMicrometers, HeightMicrometers: m.HeightMicrometers, ResolutionDPI: m.ResolutionDPI, RasterWidth: m.RasterWidth, RasterHeight: m.RasterHeight, Orientation: string(m.Orientation), ColorMode: string(m.ColorMode), CutPolicy: string(m.CutPolicy)}
		result.Artifact = &dto.PrintArtifact{SHA256: j.Artifact.SHA256, ContentType: j.Artifact.ContentType, ByteLength: j.Artifact.ByteLength, WidthPixels: j.Artifact.WidthPixels, HeightPixels: j.Artifact.HeightPixels, ExpiresAt: j.Artifact.ExpiresAt}
	}
	return result
}

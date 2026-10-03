package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func Printer(p printing.Printer) dto.Printer {
	return dto.Printer{ID: string(p.ID), Name: p.Name, AdapterID: p.AdapterID, Revision: p.Revision, Retired: p.Retired, Media: Media(printing.MediaProfile{Media: p.Media}), MediaFingerprint: p.MediaFingerprint, Readiness: string(p.Readiness), ReadinessReason: p.ReadinessReason, ReportedAt: p.ReportedAt}
}
func Printers(values []printing.Printer) []dto.Printer {
	out := make([]dto.Printer, 0, len(values))
	for _, p := range values {
		out = append(out, Printer(p))
	}
	return out
}

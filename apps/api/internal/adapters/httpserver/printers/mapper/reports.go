package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func ConnectorReportInput(in *dto.ConnectorReport) *printing.ConnectorReport {
	if in == nil {
		return nil
	}
	out := &printing.ConnectorReport{Version: in.Version, Commit: in.Commit, Platform: in.Platform, Architecture: in.Architecture, Adapters: []printing.AdapterCapability{}}
	for _, a := range in.Adapters {
		value := printing.AdapterCapability{ID: a.ID, ContractVersions: a.ContractVersions, Formats: a.Formats, CompletionEvidence: a.CompletionEvidence, Wake: a.Wake, Media: []printing.MediaCapability{}}
		for _, m := range a.Media {
			value.Media = append(value.Media, printing.MediaCapability{ID: m.ID, Version: m.Version})
		}
		out.Adapters = append(out.Adapters, value)
	}
	return out
}
func ConnectorReport(in *printing.ConnectorReport) *dto.ConnectorReport {
	if in == nil {
		return nil
	}
	out := &dto.ConnectorReport{Version: in.Version, Commit: in.Commit, Platform: in.Platform, Architecture: in.Architecture, Adapters: []dto.ConnectorAdapterCapability{}}
	for _, a := range in.Adapters {
		value := dto.ConnectorAdapterCapability{ID: a.ID, ContractVersions: a.ContractVersions, Formats: a.Formats, CompletionEvidence: a.CompletionEvidence, Wake: a.Wake, Media: []dto.ConnectorMediaCapability{}}
		for _, m := range a.Media {
			value.Media = append(value.Media, dto.ConnectorMediaCapability{ID: m.ID, Version: m.Version})
		}
		out.Adapters = append(out.Adapters, value)
	}
	return out
}

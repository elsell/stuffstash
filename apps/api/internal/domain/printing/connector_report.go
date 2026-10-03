package printing

import "slices"

type MediaCapability struct {
	ID      string
	Version uint32
}
type AdapterCapability struct {
	ID                 string
	ContractVersions   []uint32
	Formats            []string
	Media              []MediaCapability
	CompletionEvidence string
	Wake               bool
}
type ConnectorReport struct {
	Version, Commit, Platform, Architecture string
	Adapters                                []AdapterCapability
}

func reportText(s string, max int) bool {
	if len(s) == 0 || len(s) > max {
		return false
	}
	for _, r := range s {
		if r < 32 || r > 126 {
			return false
		}
	}
	return true
}
func (r ConnectorReport) Valid() bool {
	if !reportText(r.Version, 80) || !reportText(r.Commit, 80) || !reportText(r.Platform, 32) || !reportText(r.Architecture, 32) || len(r.Adapters) > 32 {
		return false
	}
	seen := map[string]bool{}
	for _, a := range r.Adapters {
		if !reportText(a.ID, 100) || seen[a.ID] || len(a.ContractVersions) < 1 || len(a.ContractVersions) > 16 || len(a.Formats) < 1 || len(a.Formats) > 16 || len(a.Media) > 64 || !reportText(a.CompletionEvidence, 100) {
			return false
		}
		seen[a.ID] = true
		for _, v := range a.ContractVersions {
			if v == 0 {
				return false
			}
		}
		for _, f := range a.Formats {
			if !reportText(f, 80) {
				return false
			}
		}
		for _, m := range a.Media {
			if !reportText(m.ID, 100) || m.Version == 0 {
				return false
			}
		}
	}
	return true
}
func (r *ConnectorReport) Clone() *ConnectorReport {
	if r == nil {
		return nil
	}
	out := *r
	out.Adapters = slices.Clone(r.Adapters)
	for i, a := range out.Adapters {
		a.ContractVersions = slices.Clone(a.ContractVersions)
		a.Formats = slices.Clone(a.Formats)
		a.Media = slices.Clone(a.Media)
		out.Adapters[i] = a
	}
	return &out
}

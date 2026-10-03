package printing

// ConnectorReport contains public software capabilities, never device identity.
type ConnectorReport struct {
	Version, Commit, Platform, Architecture string
	Adapters                                []Descriptor
}

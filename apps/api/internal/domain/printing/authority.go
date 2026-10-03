package printing

import "time"

// AcceptsAuthority is a deny fence, not an authorization grant. The application
// must first check the service principal's current SpiceDB permissions.
func AcceptsAuthority(connector Connector, binding PrinterBinding, authority ConsumerAuthority, now time.Time) bool {
	return connector.ID == authority.ConnectorID && connector.Scope == authority.Scope &&
		connector.ServiceAccountID == authority.ServiceAccountID && connector.State == ConnectorActive &&
		connector.CredentialVersion == authority.CredentialVersion && connector.CredentialExpiresAt.After(now) &&
		connector.Generation == connector.SyncedGeneration &&
		binding.Scope == authority.Scope && binding.ConnectorID == authority.ConnectorID &&
		binding.PrinterID == authority.PrinterID && !binding.Revoked &&
		binding.Generation == authority.BindingGeneration && binding.Generation == binding.SyncedGeneration
}

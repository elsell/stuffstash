package printing

import "time"

type ConnectorAvailability string

const (
	ConnectorOnline  ConnectorAvailability = "online"
	ConnectorOffline ConnectorAvailability = "offline"
	ConnectorUnknown ConnectorAvailability = "unknown"
)

// Availability describes observed connectivity, not permission or device readiness.
func (c Connector) Availability(now time.Time, maxAge time.Duration) ConnectorAvailability {
	if c.State == ConnectorRevoked || (!c.CredentialExpiresAt.IsZero() && !c.CredentialExpiresAt.After(now)) {
		return ConnectorOffline
	}
	if c.LastSeenAt == nil {
		return ConnectorUnknown
	}
	if c.State != ConnectorActive || maxAge <= 0 || now.Before(*c.LastSeenAt) || now.Sub(*c.LastSeenAt) > maxAge {
		return ConnectorOffline
	}
	return ConnectorOnline
}

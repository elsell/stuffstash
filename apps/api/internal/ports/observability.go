package ports

import "context"

type EventName string

const (
	EventHealthChecked            EventName = "health.checked"
	EventHTTPServerStartFailed    EventName = "http.server.start_failed"
	EventHTTPServerShutdownFailed EventName = "http.server.shutdown_failed"
	EventAuthenticationFailed     EventName = "authentication.failed"
	EventAuthorizationDenied      EventName = "authorization.denied"
	EventTenantCreated            EventName = "tenant.created"
	EventInventoryCreated         EventName = "inventory.created"
	EventInventoriesListed        EventName = "inventory.listed"
)

type Event struct {
	Name    EventName
	Message string
	Fields  map[string]string
}

type Observer interface {
	Record(ctx context.Context, event Event)
}

package ports

import (
	"context"
	"errors"
	"time"
)

var ErrConnectorNotRegistered = errors.New("connector is not registered; run connectors print register")

// ConnectorRegistration is secret local state, never a presentation DTO.
type ConnectorRegistration struct {
	Server      string    `json:"server"`
	TenantID    string    `json:"tenantId"`
	InventoryID string    `json:"inventoryId"`
	ConnectorID string    `json:"connectorId"`
	Credential  string    `json:"credential"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type ConnectorCredentials interface {
	Load(context.Context, string, string) (ConnectorRegistration, error)
	Save(context.Context, ConnectorRegistration) error
	Delete(context.Context, string, string) error
}

package presentation

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"time"
)

func (o Output) pairingStartedReceipt(v ports.PairingStartedReceipt) error {
	return o.details([][2]string{{"Pairing", v.Data.ID}, {"Approval code", v.Data.UserCode}, {"Verify", v.Data.VerificationURL}, {"Expires", v.Data.ExpiresAt.Format(time.RFC3339Nano)}})
}
func (o Output) pairingStatusReceipt(v ports.PairingStatusReceipt) error {
	return o.details([][2]string{{"Pairing", v.Data.ID}, {"State", v.Data.State}, {"Expires", v.Data.ExpiresAt.Format(time.RFC3339Nano)}})
}
func (o Output) pairingCredentialReceipt(v ports.PairingCredentialReceipt) error {
	return o.details([][2]string{{"Connector", v.Data.ConnectorID}, {"Household", v.Data.TenantID}, {"Inventory", v.Data.InventoryID}, {"Credential expires", v.Data.ExpiresAt.Format(time.RFC3339Nano)}, {"Activate before", v.Data.ActivationDeadline.Format(time.RFC3339Nano)}})
}

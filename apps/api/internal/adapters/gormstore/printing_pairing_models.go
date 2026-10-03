package gormstore

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"time"
)

type printingPairingModel struct {
	RotationVersion                    uint64
	Rotation                           bool
	ID                                 string `gorm:"primaryKey"`
	Name                               string
	PublicKey                          []byte
	PollHash                           string
	CodeHash                           string `gorm:"uniqueIndex"`
	CandidatesJSON                     []byte
	State                              string
	TenantID, InventoryID, ConnectorID string
	ExpiresAt, CreatedAt               time.Time
}

func (printingPairingModel) TableName() string { return "print_connector_pairings" }
func (m printingPairingModel) domain() (printing.Pairing, error) {
	var candidates []printing.PairingCandidate
	if err := json.Unmarshal(m.CandidatesJSON, &candidates); err != nil {
		return printing.Pairing{}, err
	}
	return printing.Pairing{Rotation: m.Rotation, RotationVersion: m.RotationVersion, ID: printing.PairingID(m.ID), Name: m.Name, PublicKey: append([]byte(nil), m.PublicKey...), PollHash: m.PollHash, CodeHash: m.CodeHash, Candidates: candidates, State: printing.PairingState(m.State), Scope: printing.Scope{TenantID: m.TenantID, InventoryID: m.InventoryID}, ConnectorID: printing.ConnectorID(m.ConnectorID), ExpiresAt: m.ExpiresAt, CreatedAt: m.CreatedAt}, nil
}
func printingPairingFromDomain(p printing.Pairing) (printingPairingModel, error) {
	candidates, err := json.Marshal(p.Candidates)
	return printingPairingModel{Rotation: p.Rotation, RotationVersion: p.RotationVersion, ID: string(p.ID), Name: p.Name, PublicKey: append([]byte(nil), p.PublicKey...), PollHash: p.PollHash, CodeHash: p.CodeHash, CandidatesJSON: candidates, State: string(p.State), TenantID: p.Scope.TenantID, InventoryID: p.Scope.InventoryID, ConnectorID: string(p.ConnectorID), ExpiresAt: p.ExpiresAt, CreatedAt: p.CreatedAt}, err
}
func printingConnectorFromDomain(c printing.Connector) printingConnectorModel {
	return printingConnectorModel{PendingCredentialHash: c.PendingCredentialHash, PendingCredentialVersion: c.PendingCredentialVersion, PendingCredentialExpiresAt: c.PendingCredentialExpiresAt, PendingActivationDeadline: c.PendingActivationDeadline, PendingPublicKey: append([]byte(nil), c.PendingPublicKey...), ID: string(c.ID), TenantID: c.Scope.TenantID, InventoryID: c.Scope.InventoryID, ServiceAccountID: string(c.ServiceAccountID), Name: c.Name, State: string(c.State), PublicKey: append([]byte(nil), c.PublicKey...), CredentialHash: c.CredentialHash, CredentialVersion: c.CredentialVersion, CredentialExpiresAt: c.CredentialExpiresAt, ActivationDeadline: c.ActivationDeadline, Generation: c.Generation, SyncedGeneration: c.SyncedGeneration, LastSeenAt: c.LastSeenAt, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
}
func printingBindingFromDomain(b printing.PrinterBinding) printingBindingModel {
	return printingBindingModel{ConnectorID: string(b.ConnectorID), PrinterID: string(b.PrinterID), TenantID: b.Scope.TenantID, InventoryID: b.Scope.InventoryID, DeviceID: b.DeviceID, Generation: b.Generation, SyncedGeneration: b.SyncedGeneration, Revoked: b.Revoked}
}

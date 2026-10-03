package printregistry

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type RotateConnectorCredential struct {
	Actor       Actor
	ConnectorID printing.ConnectorID
	Generation  uint64
	PairingID   printing.PairingID
	UserCode    string
}

func (s ConnectorService) RotateCredential(ctx context.Context, input RotateConnectorCredential) (printing.Pairing, error) {
	if err := s.Registry.access(ctx, input.Actor, ports.InventoryPermissionConfigure); err != nil {
		return printing.Pairing{}, err
	}
	p, err := s.Repository.GetPrintPairing(ctx, input.PairingID)
	if err != nil {
		return printing.Pairing{}, connectorError(err)
	}
	if input.Generation == 0 || p.State != printing.PairingPending || !p.ExpiresAt.After(s.Registry.Clock.Now()) || !s.Secrets.Matches(input.UserCode, p.CodeHash) {
		return printing.Pairing{}, apperrors.ErrInvalidInput
	}
	record, err := s.connectorAudit(input.Actor, audit.ActionPrintConnectorRotationRequested, input.ConnectorID)
	if err != nil {
		return printing.Pairing{}, err
	}
	result, err := s.Repository.ApprovePrintCredentialRotation(ctx, ports.RotationApproval{PairingID: p.ID, CodeHash: p.CodeHash, Scope: input.Actor.Scope, ConnectorID: input.ConnectorID, Generation: input.Generation, Now: s.Registry.Clock.Now(), Audit: record})
	return result, connectorError(err)
}

func (s ConnectorService) machineAudit(c printing.Connector, action audit.Action) (audit.Record, error) {
	record, ok := audit.NewRecord(audit.ID(s.Registry.IDs.NewID()), audit.TenantID(c.Scope.TenantID), audit.InventoryID(c.Scope.InventoryID), audit.PrincipalID(c.ServiceAccountID), action, audit.SourceAPI, audit.TargetInventory, c.Scope.InventoryID, s.Registry.Clock.Now(), "", map[string]string{"connector_id": string(c.ID), "actor_kind": "service_account"})
	if !ok {
		return audit.Record{}, apperrors.ErrInvalidInput
	}
	return record, nil
}

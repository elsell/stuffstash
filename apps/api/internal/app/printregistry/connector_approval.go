package printregistry

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s ConnectorService) Approve(ctx context.Context, input ApprovePairing) (ports.ConnectorRegistration, error) {
	if err := s.Registry.access(ctx, input.Actor, ports.InventoryPermissionConfigure); err != nil {
		return ports.ConnectorRegistration{}, err
	}
	p, err := s.Repository.GetPrintPairing(ctx, input.PairingID)
	if err != nil {
		return ports.ConnectorRegistration{}, connectorError(err)
	}
	now := s.Registry.Clock.Now()
	if p.State != printing.PairingPending || !p.ExpiresAt.After(now) || !s.Secrets.Matches(input.UserCode, p.CodeHash) || len(input.Bindings) < 1 || len(input.Bindings) > len(p.Candidates) {
		return ports.ConnectorRegistration{}, apperrors.ErrInvalidInput
	}
	c := printing.Connector{ID: printing.ConnectorID(s.Registry.IDs.NewID()), Scope: input.Actor.Scope, ServiceAccountID: printing.ServiceAccountID(s.Registry.IDs.NewID()), Name: p.Name, State: printing.ConnectorPending, PublicKey: append([]byte(nil), p.PublicKey...), CredentialVersion: 1, Generation: 1, CreatedAt: now, UpdatedAt: now}
	bindings := make([]printing.PrinterBinding, 0, len(input.Bindings))
	seen := map[string]bool{}
	for _, mapping := range input.Bindings {
		var candidate printing.PairingCandidate
		for _, v := range p.Candidates {
			if v.ID == mapping.CandidateID {
				candidate = v
				break
			}
		}
		if candidate.ID == "" || seen[candidate.ID] {
			return ports.ConnectorRegistration{}, apperrors.ErrInvalidInput
		}
		seen[candidate.ID] = true
		printer, err := s.Registry.Printers.GetPrinter(ctx, input.Actor.Scope, mapping.PrinterID)
		if err != nil {
			return ports.ConnectorRegistration{}, registryError(err)
		}
		if printer.AdapterID != candidate.AdapterID || printer.Retired {
			return ports.ConnectorRegistration{}, apperrors.ErrConflict
		}
		bindings = append(bindings, printing.PrinterBinding{Scope: c.Scope, PrinterID: printer.ID, ConnectorID: c.ID, DeviceID: candidate.DeviceID, Generation: 1})
	}
	record, ok := audit.NewRecord(audit.ID(s.Registry.IDs.NewID()), audit.TenantID(c.Scope.TenantID), audit.InventoryID(c.Scope.InventoryID), audit.PrincipalID(input.Actor.Principal.ID), audit.ActionPrintConnectorApproved, audit.SourceAPI, audit.TargetInventory, c.Scope.InventoryID, now, input.Actor.RequestID, map[string]string{"connector_id": string(c.ID)})
	if !ok {
		return ports.ConnectorRegistration{}, apperrors.ErrInvalidInput
	}
	registration, err := s.Repository.ApprovePrintPairing(ctx, ports.PairingApproval{PairingID: p.ID, CodeHash: p.CodeHash, Now: now, Registration: ports.ConnectorRegistration{Connector: c, Bindings: bindings}, Audit: record})
	if err != nil {
		return ports.ConnectorRegistration{}, connectorError(err)
	}
	// Approval is durable even if SpiceDB is temporarily unavailable. Pending
	// grants stay denied and the reconciliation worker retries the stored intent.
	if err := s.Reconcile(ctx, c.Scope, c.ID); err != nil {
		return registration, nil
	}
	return s.Repository.GetPrintConnector(ctx, c.Scope, c.ID)
}

package printregistry

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strings"
)

type UpdateConnector struct {
	Actor      Actor
	ID         printing.ConnectorID
	Generation uint64
	Name       *string
	Revoked    *bool
	PrinterIDs *[]printing.PrinterID
}

func (s ConnectorService) connectorAudit(a Actor, action audit.Action, id printing.ConnectorID) (audit.Record, error) {
	record, ok := audit.NewRecord(audit.ID(s.Registry.IDs.NewID()), audit.TenantID(a.Scope.TenantID), audit.InventoryID(a.Scope.InventoryID), audit.PrincipalID(a.Principal.ID), action, audit.SourceAPI, audit.TargetInventory, a.Scope.InventoryID, s.Registry.Clock.Now(), a.RequestID, map[string]string{"connector_id": string(id)})
	if !ok {
		return audit.Record{}, apperrors.ErrInvalidInput
	}
	return record, nil
}
func (s ConnectorService) Update(ctx context.Context, input UpdateConnector) (ports.ConnectorRegistration, error) {
	if err := s.Registry.access(ctx, input.Actor, ports.InventoryPermissionConfigure); err != nil {
		return ports.ConnectorRegistration{}, err
	}
	if input.Generation == 0 || input.Name == nil && input.Revoked == nil && input.PrinterIDs == nil {
		return ports.ConnectorRegistration{}, apperrors.ErrInvalidInput
	}
	r, err := s.Repository.UpdatePrintConnector(ctx, input.Actor.Scope, input.ID, input.Generation, func(r *ports.ConnectorRegistration) error {
		if input.Name != nil {
			name := strings.TrimSpace(*input.Name)
			if !validPrinterName(name) {
				return apperrors.ErrInvalidInput
			}
			r.Connector.Name = name
		}
		if r.Connector.State == printing.ConnectorRevoked {
			return apperrors.ErrConflict
		}
		if input.Revoked != nil {
			if !*input.Revoked {
				return apperrors.ErrInvalidInput
			}
			r.Connector.State = printing.ConnectorRevoked
			r.Connector.CredentialHash = ""
			r.Connector.PendingCredentialHash = ""
			r.Connector.PendingCredentialVersion = 0
			r.Connector.PendingPublicKey = nil
			r.Connector.CredentialVersion++
		}
		r.Connector.Generation++
		r.Connector.UpdatedAt = s.Registry.Clock.Now()
		selected := map[printing.PrinterID]bool{}
		if input.PrinterIDs != nil {
			for _, id := range *input.PrinterIDs {
				if selected[id] {
					return apperrors.ErrInvalidInput
				}
				selected[id] = true
			}
		}
		for i := range r.Bindings {
			b := &r.Bindings[i]
			if input.PrinterIDs != nil {
				b.Revoked = !selected[b.PrinterID]
				delete(selected, b.PrinterID)
			}
			if r.Connector.State == printing.ConnectorRevoked {
				b.Revoked = true
			}
			b.Generation = r.Connector.Generation
		}
		if len(selected) > 0 {
			return apperrors.ErrInvalidInput
		}
		return nil
	}, func(r ports.ConnectorRegistration) (audit.Record, error) {
		return s.connectorAudit(input.Actor, audit.ActionPrintConnectorUpdated, r.Connector.ID)
	})
	if err != nil {
		return ports.ConnectorRegistration{}, registryError(err)
	}
	if err := s.Reconcile(ctx, r.Connector.Scope, r.Connector.ID); err != nil {
		return r, nil
	}
	return s.Repository.GetPrintConnector(ctx, r.Connector.Scope, r.Connector.ID)
}
func (s ConnectorService) Get(ctx context.Context, a Actor, id printing.ConnectorID) (ports.ConnectorRegistration, error) {
	if err := s.Registry.access(ctx, a, ports.InventoryPermissionView); err != nil {
		return ports.ConnectorRegistration{}, err
	}
	r, err := s.Repository.GetPrintConnector(ctx, a.Scope, id)
	if err != nil {
		return r, registryError(err)
	}
	record, err := s.connectorAudit(a, audit.ActionPrintConnectorViewed, id)
	if err != nil {
		return r, err
	}
	return r, s.Registry.Audit.SaveAuditRecord(ctx, record)
}

type ConnectorPage struct {
	Items      []ports.ConnectorRegistration
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (s ConnectorService) List(ctx context.Context, a Actor, limit int, cursor string) (ConnectorPage, error) {
	if err := s.Registry.access(ctx, a, ports.InventoryPermissionView); err != nil {
		return ConnectorPage{}, err
	}
	if limit < 1 || limit > 100 {
		return ConnectorPage{}, apperrors.ErrInvalidInput
	}
	scope := a.Scope.TenantID + ":" + a.Scope.InventoryID
	after, err := appsupport.DecodePageCursor("print_connectors", scope, cursor)
	if err != nil {
		return ConnectorPage{}, err
	}
	records, err := s.Repository.ListPrintConnectors(ctx, a.Scope, limit+1, after)
	if err != nil {
		return ConnectorPage{}, registryError(err)
	}
	result := ConnectorPage{Items: records, Limit: limit, HasMore: len(records) > limit}
	if result.HasMore {
		result.Items = records[:limit]
		result.NextCursor = appsupport.EncodePageCursor("print_connectors", scope, string(result.Items[len(result.Items)-1].Connector.ID))
	}
	record, err := s.connectorAudit(a, audit.ActionPrintConnectorListed, "")
	if err != nil {
		return ConnectorPage{}, err
	}
	if err := s.Registry.Audit.SaveAuditRecord(ctx, record); err != nil {
		return ConnectorPage{}, err
	}
	return result, nil
}

type PairingReview struct {
	Rotation                   bool
	ID                         printing.PairingID
	Name, PublicKeyFingerprint string
	Candidates                 []printing.PairingCandidate
}

func (s ConnectorService) Review(ctx context.Context, a Actor, id printing.PairingID, code string) (PairingReview, error) {
	if err := s.Registry.access(ctx, a, ports.InventoryPermissionConfigure); err != nil {
		return PairingReview{}, err
	}
	p, err := s.Repository.GetPrintPairing(ctx, id)
	if err != nil {
		return PairingReview{}, connectorError(err)
	}
	if p.State != printing.PairingPending || !p.ExpiresAt.After(s.Registry.Clock.Now()) || !s.Secrets.Matches(code, p.CodeHash) {
		return PairingReview{}, apperrors.ErrInvalidInput
	}
	candidates := make([]printing.PairingCandidate, len(p.Candidates))
	for i, c := range p.Candidates {
		c.DeviceID = ""
		candidates[i] = c
	}
	record, err := s.connectorAudit(a, audit.ActionPrintPairingReviewed, "")
	if err != nil {
		return PairingReview{}, err
	}
	if err := s.Registry.Audit.SaveAuditRecord(ctx, record); err != nil {
		return PairingReview{}, err
	}
	return PairingReview{Rotation: p.Rotation, ID: p.ID, Name: p.Name, PublicKeyFingerprint: s.Secrets.Digest(string(p.PublicKey)), Candidates: candidates}, nil
}

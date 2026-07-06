package appsupport

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type AuditRecordInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	RequestID   string
	Source      audit.Source
	Action      audit.Action
	TargetType  audit.TargetType
	TargetID    string
	Metadata    map[string]string
}

func NewAuditRecord(ids ports.IDGenerator, clock ports.Clock, input AuditRecordInput) (audit.Record, error) {
	if ids == nil {
		return audit.Record{}, apperrors.ErrInvalidInput
	}
	if clock == nil {
		clock = ports.SystemClock{}
	}
	id, ok := audit.NewID(ids.NewID())
	if !ok {
		return audit.Record{}, apperrors.ErrInvalidInput
	}
	source := input.Source
	if source.String() == "" {
		source = audit.SourceAPI
	}
	record, ok := audit.NewRecord(
		id,
		audit.TenantID(input.TenantID.String()),
		audit.InventoryID(input.InventoryID.String()),
		audit.PrincipalID(input.Principal.ID.String()),
		input.Action,
		source,
		input.TargetType,
		input.TargetID,
		clock.Now(),
		input.RequestID,
		input.Metadata,
	)
	if !ok {
		return audit.Record{}, apperrors.ErrInvalidInput
	}
	return record, nil
}

func SaveReadAuditRecord(ctx context.Context, repo ports.AuditRepository, ids ports.IDGenerator, clock ports.Clock, input AuditRecordInput) error {
	if repo == nil {
		return apperrors.ErrInvalidInput
	}
	record, err := NewAuditRecord(ids, clock, input)
	if err != nil {
		return err
	}
	return repo.SaveAuditRecord(ctx, record)
}

package inventories

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) GetTenant(ctx context.Context, input GetTenantInput) (tenant.Tenant, error) {
	item, found, err := a.tenants.TenantByID(ctx, input.TenantID)
	if err != nil {
		return tenant.Tenant{}, err
	}
	if !found || !item.IsActive() {
		return tenant.Tenant{}, apperrors.ErrNotFound
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionView, input.TenantID); err != nil {
		a.RecordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return tenant.Tenant{}, err
	}
	if err := appsupport.SaveReadAuditRecord(ctx, a.audit, a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:  input.Principal,
		TenantID:   input.TenantID,
		Source:     input.Source,
		RequestID:  input.RequestID,
		Action:     audit.ActionTenantViewed,
		TargetType: audit.TargetTenant,
		TargetID:   item.ID.String(),
		Metadata:   map[string]string{},
	}); err != nil {
		return tenant.Tenant{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventTenantViewed,
		Message: "tenant viewed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return item, nil
}

func (a Service) UpdateTenant(ctx context.Context, input UpdateTenantInput) (tenant.Tenant, error) {
	if input.Name == nil {
		return tenant.Tenant{}, apperrors.ErrInvalidInput
	}
	current, found, err := a.tenants.TenantByID(ctx, input.TenantID)
	if err != nil {
		return tenant.Tenant{}, err
	}
	if !found || !current.IsActive() {
		return tenant.Tenant{}, apperrors.ErrNotFound
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.RecordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return tenant.Tenant{}, err
	}
	name, ok := tenant.NewName(*input.Name)
	if !ok {
		return tenant.Tenant{}, apperrors.ErrInvalidInput
	}
	updated := current
	updated.Name = name
	if updated.Name == current.Name {
		return current, nil
	}
	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:  input.Principal,
		TenantID:   input.TenantID,
		Source:     input.Source,
		RequestID:  input.RequestID,
		Action:     audit.ActionTenantUpdated,
		TargetType: audit.TargetTenant,
		TargetID:   updated.ID.String(),
		Metadata: map[string]string{
			"name": updated.Name.String(),
		},
	})
	if err != nil {
		return tenant.Tenant{}, err
	}
	if err := a.tenantUnitOfWork.UpdateTenant(ctx, updated, auditRecord); err != nil {
		return tenant.Tenant{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventTenantUpdated,
		Message: "tenant updated",
		Fields: map[string]string{
			"tenant_id":    updated.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return updated, nil
}

func (a Service) ArchiveTenant(ctx context.Context, input UpdateTenantLifecycleInput) (tenant.Tenant, error) {
	return a.updateTenantLifecycle(ctx, input, tenant.LifecycleStateActive, tenant.LifecycleStateArchived, audit.ActionTenantArchived, ports.EventTenantArchived, "tenant archived")
}

func (a Service) RestoreTenant(ctx context.Context, input UpdateTenantLifecycleInput) (tenant.Tenant, error) {
	return a.updateTenantLifecycle(ctx, input, tenant.LifecycleStateArchived, tenant.LifecycleStateActive, audit.ActionTenantRestored, ports.EventTenantRestored, "tenant restored")
}

func (a Service) updateTenantLifecycle(ctx context.Context, input UpdateTenantLifecycleInput, from tenant.LifecycleState, to tenant.LifecycleState, action audit.Action, eventName ports.EventName, eventMessage string) (tenant.Tenant, error) {
	current, found, err := a.tenants.TenantByID(ctx, input.TenantID)
	if err != nil {
		return tenant.Tenant{}, err
	}
	if !found {
		return tenant.Tenant{}, apperrors.ErrNotFound
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.RecordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return tenant.Tenant{}, err
	}
	if current.LifecycleState != from {
		return tenant.Tenant{}, apperrors.ErrInvalidInput
	}
	updated := current
	updated.LifecycleState = to
	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:  input.Principal,
		TenantID:   input.TenantID,
		Source:     input.Source,
		RequestID:  input.RequestID,
		Action:     action,
		TargetType: audit.TargetTenant,
		TargetID:   updated.ID.String(),
		Metadata: map[string]string{
			"previous_state":  current.LifecycleState.String(),
			"lifecycle_state": updated.LifecycleState.String(),
		},
	})
	if err != nil {
		return tenant.Tenant{}, err
	}
	if err := a.tenantUnitOfWork.UpdateTenantLifecycle(ctx, updated, auditRecord); err != nil {
		return tenant.Tenant{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    eventName,
		Message: eventMessage,
		Fields: map[string]string{
			"tenant_id":       updated.ID.String(),
			"principal_id":    input.Principal.ID.String(),
			"lifecycle_state": updated.LifecycleState.String(),
		},
	})
	return updated, nil
}

func (a Service) DeleteTenant(ctx context.Context, input UpdateTenantLifecycleInput) error {
	current, found, err := a.tenants.TenantByID(ctx, input.TenantID)
	if err != nil {
		return err
	}
	if !found {
		return apperrors.ErrNotFound
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.RecordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return err
	}
	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:  input.Principal,
		TenantID:   input.TenantID,
		Source:     input.Source,
		RequestID:  input.RequestID,
		Action:     audit.ActionTenantDeleted,
		TargetType: audit.TargetTenant,
		TargetID:   current.ID.String(),
		Metadata: map[string]string{
			"lifecycle_state": current.LifecycleState.String(),
		},
	})
	if err != nil {
		return err
	}
	if err := a.tenantUnitOfWork.DeleteTenant(ctx, input.TenantID, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return apperrors.ErrInvalidInput
		}
		return err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventTenantDeleted,
		Message: "tenant deleted",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return nil
}

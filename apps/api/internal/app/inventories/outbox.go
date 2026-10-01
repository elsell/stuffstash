package inventories

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) DrainAuthorizationOutboxBestEffort(ctx context.Context, limit int) {
	if err := a.DrainAuthorizationOutbox(ctx, limit); err != nil {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventAuthorizationOutboxFailed,
			Message: "authorization outbox drain failed",
			Fields:  map[string]string{"error": err.Error()},
		})
	}
}

func (a Service) DrainAuthorizationOutbox(ctx context.Context, limit int) error {
	if limit <= 0 {
		limit = a.AuthorizationOutboxDrainLimit()
	}

	claimID := a.ids.NewID()
	now := a.clock.Now().UTC()
	events, err := a.outbox.ClaimPendingAuthorizationOutboxEvents(ctx, claimID, limit, now, now.Add(a.AuthorizationOutboxClaimLease()))
	if err != nil {
		return err
	}

	var drainErr error
	processedCount := 0
	failedCount := 0
	deadLetteredCount := 0
	for _, event := range events {
		if err := a.applyAuthorizationOutboxEvent(ctx, event); err != nil {
			if isUnrecoverableAuthorizationOutboxError(err) {
				if markErr := a.outbox.MarkAuthorizationOutboxEventDeadLettered(ctx, event.ID, claimID, err.Error()); markErr != nil {
					failedCount++
					a.recordAuthorizationOutboxEventFailed(ctx, event, markErr)
					drainErr = errors.Join(drainErr, markErr)
					continue
				}
				deadLetteredCount++
				a.recordAuthorizationOutboxEventDeadLettered(ctx, event, err)
				continue
			}
			failedCount++
			if markErr := a.outbox.MarkAuthorizationOutboxEventFailed(ctx, event.ID, claimID, err.Error()); markErr != nil {
				a.recordAuthorizationOutboxEventFailed(ctx, event, markErr)
				drainErr = errors.Join(drainErr, markErr)
				continue
			}
			a.recordAuthorizationOutboxEventFailed(ctx, event, err)
			drainErr = errors.Join(drainErr, err)
			continue
		}
		if err := a.outbox.MarkAuthorizationOutboxEventProcessed(ctx, event.ID, claimID); err != nil {
			failedCount++
			drainErr = errors.Join(drainErr, err)
			continue
		}
		processedCount++
	}

	if len(events) > 0 {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventAuthorizationOutboxDrained,
			Message: "authorization outbox drained",
			Fields: map[string]string{
				"event_count":         strconv.Itoa(len(events)),
				"processed_count":     strconv.Itoa(processedCount),
				"failed_count":        strconv.Itoa(failedCount),
				"dead_lettered_count": strconv.Itoa(deadLetteredCount),
			},
		})
	}

	return drainErr
}

func (a Service) DrainAuthorizationOutboxEvent(ctx context.Context, eventID string) error {
	claimID := a.ids.NewID()
	event, found, err := a.outbox.ClaimAuthorizationOutboxEvent(ctx, eventID, claimID, a.clock.Now().Add(a.AuthorizationOutboxClaimLease()))
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%w: authorization outbox event %q is not claimable", ports.ErrOutboxClaimLost, eventID)
	}
	return a.ProcessClaimedAuthorizationOutboxEvent(ctx, event, claimID)
}

func (a Service) ProcessClaimedAuthorizationOutboxEvent(ctx context.Context, event ports.AuthorizationOutboxEvent, claimID string) error {
	if err := a.applyAuthorizationOutboxEvent(ctx, event); err != nil {
		if isUnrecoverableAuthorizationOutboxError(err) {
			if markErr := a.outbox.MarkAuthorizationOutboxEventDeadLettered(ctx, event.ID, claimID, err.Error()); markErr != nil {
				a.recordAuthorizationOutboxEventFailed(ctx, event, markErr)
				return markErr
			}
			a.recordAuthorizationOutboxEventDeadLettered(ctx, event, err)
			return err
		}
		if markErr := a.outbox.MarkAuthorizationOutboxEventFailed(ctx, event.ID, claimID, err.Error()); markErr != nil {
			a.recordAuthorizationOutboxEventFailed(ctx, event, markErr)
			return markErr
		}
		a.recordAuthorizationOutboxEventFailed(ctx, event, err)
		return err
	}
	if err := a.outbox.MarkAuthorizationOutboxEventProcessed(ctx, event.ID, claimID); err != nil {
		return err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAuthorizationOutboxDrained,
		Message: "authorization outbox event drained",
		Fields: map[string]string{
			"event_count":         "1",
			"processed_count":     "1",
			"failed_count":        "0",
			"dead_lettered_count": "0",
		},
	})
	return nil
}

func (a Service) applyAuthorizationOutboxEvent(ctx context.Context, event ports.AuthorizationOutboxEvent) error {
	return ApplyAuthorizationOutboxEvent(ctx, a.authorizer, event)
}

func ApplyAuthorizationOutboxEvent(ctx context.Context, authorizer ports.Authorizer, event ports.AuthorizationOutboxEvent) error {
	if err := validateAuthorizationOutboxEvent(event); err != nil {
		return err
	}

	principal := identity.Principal{ID: event.PrincipalID}
	switch event.Kind {
	case ports.AuthorizationOutboxGrantTenantOwner:
		return authorizer.GrantTenantOwner(ctx, principal, event.TenantID)
	case ports.AuthorizationOutboxGrantInventoryOwner:
		return authorizer.GrantInventoryOwner(ctx, principal, event.TenantID, event.InventoryID)
	case ports.AuthorizationOutboxGrantInventoryViewer:
		return authorizer.GrantInventoryViewer(ctx, principal, event.TenantID, event.InventoryID)
	case ports.AuthorizationOutboxGrantInventoryEditor:
		return authorizer.GrantInventoryEditor(ctx, principal, event.TenantID, event.InventoryID)
	case ports.AuthorizationOutboxRevokeInventoryViewer:
		return authorizer.RevokeInventoryViewer(ctx, principal, event.TenantID, event.InventoryID)
	case ports.AuthorizationOutboxRevokeInventoryEditor:
		return authorizer.RevokeInventoryEditor(ctx, principal, event.TenantID, event.InventoryID)
	default:
		return apperrors.ErrInvalidInput
	}
}

func validateAuthorizationOutboxEvent(event ports.AuthorizationOutboxEvent) error {
	if _, ok := identity.NewPrincipalID(event.PrincipalID.String()); !ok {
		return fmt.Errorf("%w: authorization outbox event principal id is invalid", apperrors.ErrInvalidInput)
	}
	if _, ok := tenant.NewID(event.TenantID.String()); !ok {
		return fmt.Errorf("%w: authorization outbox event tenant id is invalid", apperrors.ErrInvalidInput)
	}

	switch event.Kind {
	case ports.AuthorizationOutboxGrantTenantOwner:
		if event.InventoryID.String() != "" {
			return fmt.Errorf("%w: tenant owner grant must not include inventory id", apperrors.ErrInvalidInput)
		}
	case ports.AuthorizationOutboxGrantInventoryOwner, ports.AuthorizationOutboxGrantInventoryViewer, ports.AuthorizationOutboxGrantInventoryEditor, ports.AuthorizationOutboxRevokeInventoryViewer, ports.AuthorizationOutboxRevokeInventoryEditor:
		if _, ok := inventory.NewID(event.InventoryID.String()); !ok {
			return fmt.Errorf("%w: inventory grant inventory id is invalid", apperrors.ErrInvalidInput)
		}
	default:
		return fmt.Errorf("%w: authorization outbox event kind is unsupported", apperrors.ErrInvalidInput)
	}
	return nil
}

func isUnrecoverableAuthorizationOutboxError(err error) bool {
	return errors.Is(err, apperrors.ErrInvalidInput)
}

func (a Service) recordAuthorizationOutboxEventFailed(ctx context.Context, event ports.AuthorizationOutboxEvent, err error) {
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAuthorizationOutboxFailed,
		Message: "authorization outbox event failed",
		Fields: map[string]string{
			"event_id":     event.ID,
			"event_kind":   string(event.Kind),
			"tenant_id":    event.TenantID.String(),
			"inventory_id": event.InventoryID.String(),
			"attempts":     strconv.Itoa(event.Attempts + 1),
			"error":        err.Error(),
		},
	})
}

func (a Service) recordAuthorizationOutboxEventDeadLettered(ctx context.Context, event ports.AuthorizationOutboxEvent, err error) {
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAuthorizationOutboxDeadLettered,
		Message: "authorization outbox event dead-lettered",
		Fields: map[string]string{
			"event_id":     event.ID,
			"event_kind":   string(event.Kind),
			"tenant_id":    event.TenantID.String(),
			"inventory_id": event.InventoryID.String(),
			"reason":       err.Error(),
		},
	})
}

func (a Service) AuthorizationOutboxDrainLimit() int {
	if a.outboxDrainLimit <= 0 {
		return 25
	}
	return a.outboxDrainLimit
}

func (a Service) AuthorizationOutboxClaimLease() time.Duration {
	if a.outboxClaimLease <= 0 {
		return 30 * time.Second
	}
	return a.outboxClaimLease
}

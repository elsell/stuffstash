package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateTenantInput struct {
	Principal identity.Principal
	Name      string
}

type CreateInventoryInput struct {
	Principal identity.Principal
	TenantID  tenant.ID
	Name      string
}

type ListInventoriesInput struct {
	Principal identity.Principal
	TenantID  tenant.ID
	Limit     int
	Cursor    string
}

type ListInventoriesResult struct {
	Items      []inventory.Inventory
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (a App) CurrentPrincipal(principal identity.Principal) identity.Principal {
	return principal
}

func (a App) CreateTenant(ctx context.Context, input CreateTenantInput) (tenant.Tenant, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return tenant.Tenant{}, ErrInvalidInput
	}

	id := a.ids.NewID()

	tenantName, ok := tenant.NewName(name)
	if !ok {
		return tenant.Tenant{}, ErrInvalidInput
	}
	tenantID, ok := tenant.NewID(id)
	if !ok {
		return tenant.Tenant{}, ErrInvalidInput
	}

	item := tenant.Tenant{
		ID:   tenantID,
		Name: tenantName,
	}

	if err := a.outbox.SaveTenantAndEnqueueOwnerGrant(ctx, a.ids.NewID(), item, input.Principal); err != nil {
		return tenant.Tenant{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventTenantCreated,
		Message: "tenant created",
		Fields: map[string]string{
			"tenant_id":    item.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	a.drainAuthorizationOutboxBestEffort(ctx, a.authorizationOutboxDrainLimit())

	return item, nil
}

func (a App) CreateInventory(ctx context.Context, input CreateInventoryInput) (inventory.Inventory, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return inventory.Inventory{}, ErrInvalidInput
	}

	exists, err := a.tenants.TenantExists(ctx, input.TenantID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !exists {
		return inventory.Inventory{}, ErrNotFound
	}

	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionCreateInventory, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return inventory.Inventory{}, err
	}

	id := a.ids.NewID()
	inventoryID, ok := inventory.NewID(id)
	if !ok {
		return inventory.Inventory{}, ErrInvalidInput
	}
	inventoryName, ok := inventory.NewName(name)
	if !ok {
		return inventory.Inventory{}, ErrInvalidInput
	}

	item := inventory.Inventory{
		ID:       inventoryID,
		TenantID: inventory.TenantID(input.TenantID.String()),
		Name:     inventoryName,
	}

	if err := a.outbox.SaveInventoryAndEnqueueOwnerGrant(ctx, a.ids.NewID(), item, input.TenantID, input.Principal); err != nil {
		return inventory.Inventory{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryCreated,
		Message: "inventory created",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": item.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	a.drainAuthorizationOutboxBestEffort(ctx, a.authorizationOutboxDrainLimit())

	return item, nil
}

func (a App) drainAuthorizationOutboxBestEffort(ctx context.Context, limit int) {
	if err := a.DrainAuthorizationOutbox(ctx, limit); err != nil {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventAuthorizationOutboxFailed,
			Message: "authorization outbox drain failed",
			Fields:  map[string]string{"error": err.Error()},
		})
	}
}

func (a App) DrainAuthorizationOutbox(ctx context.Context, limit int) error {
	if limit <= 0 {
		limit = a.authorizationOutboxDrainLimit()
	}

	claimID := a.ids.NewID()
	events, err := a.outbox.ClaimPendingAuthorizationOutboxEvents(ctx, claimID, limit, time.Now().Add(a.authorizationOutboxClaimLease()))
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

func (a App) applyAuthorizationOutboxEvent(ctx context.Context, event ports.AuthorizationOutboxEvent) error {
	if err := validateAuthorizationOutboxEvent(event); err != nil {
		return err
	}

	principal := identity.Principal{ID: event.PrincipalID}
	switch event.Kind {
	case ports.AuthorizationOutboxGrantTenantOwner:
		return a.authorizer.GrantTenantOwner(ctx, principal, event.TenantID)
	case ports.AuthorizationOutboxGrantInventoryOwner:
		return a.authorizer.GrantInventoryOwner(ctx, principal, event.TenantID, event.InventoryID)
	case ports.AuthorizationOutboxGrantInventoryViewer:
		return a.authorizer.GrantInventoryViewer(ctx, principal, event.TenantID, event.InventoryID)
	case ports.AuthorizationOutboxGrantInventoryEditor:
		return a.authorizer.GrantInventoryEditor(ctx, principal, event.TenantID, event.InventoryID)
	default:
		return ErrInvalidInput
	}
}

func validateAuthorizationOutboxEvent(event ports.AuthorizationOutboxEvent) error {
	if _, ok := identity.NewPrincipalID(event.PrincipalID.String()); !ok {
		return fmt.Errorf("%w: authorization outbox event principal id is invalid", ErrInvalidInput)
	}
	if _, ok := tenant.NewID(event.TenantID.String()); !ok {
		return fmt.Errorf("%w: authorization outbox event tenant id is invalid", ErrInvalidInput)
	}

	switch event.Kind {
	case ports.AuthorizationOutboxGrantTenantOwner:
		if event.InventoryID.String() != "" {
			return fmt.Errorf("%w: tenant owner grant must not include inventory id", ErrInvalidInput)
		}
	case ports.AuthorizationOutboxGrantInventoryOwner, ports.AuthorizationOutboxGrantInventoryViewer, ports.AuthorizationOutboxGrantInventoryEditor:
		if _, ok := inventory.NewID(event.InventoryID.String()); !ok {
			return fmt.Errorf("%w: inventory grant inventory id is invalid", ErrInvalidInput)
		}
	default:
		return fmt.Errorf("%w: authorization outbox event kind is unsupported", ErrInvalidInput)
	}
	return nil
}

func isUnrecoverableAuthorizationOutboxError(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

func (a App) recordAuthorizationOutboxEventFailed(ctx context.Context, event ports.AuthorizationOutboxEvent, err error) {
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

func (a App) recordAuthorizationOutboxEventDeadLettered(ctx context.Context, event ports.AuthorizationOutboxEvent, err error) {
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

func (a App) authorizationOutboxDrainLimit() int {
	if a.outboxDrainLimit <= 0 {
		return 25
	}
	return a.outboxDrainLimit
}

func (a App) authorizationOutboxClaimLease() time.Duration {
	if a.outboxClaimLease <= 0 {
		return 30 * time.Second
	}
	return a.outboxClaimLease
}

func (a App) ListInventories(ctx context.Context, input ListInventoriesInput) (ListInventoriesResult, error) {
	exists, err := a.tenants.TenantExists(ctx, input.TenantID)
	if err != nil {
		return ListInventoriesResult{}, err
	}
	if !exists {
		return ListInventoriesResult{}, ErrNotFound
	}

	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionView, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return ListInventoriesResult{}, err
	}

	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterInventoryID, err := decodeInventoryCursor(input.TenantID, input.Cursor)
	if err != nil {
		return ListInventoriesResult{}, ErrInvalidInput
	}

	visible := make([]inventory.Inventory, 0, limit+1)
	items, err := a.inventories.ListInventoriesByTenant(ctx, inventory.TenantID(input.TenantID.String()), ports.InventoryListPageRequest{
		AfterInventoryID: afterInventoryID,
		Limit:            inventoryScanLimit(limit),
	})
	if err != nil {
		return ListInventoriesResult{}, err
	}

	lastScannedID := inventory.InventoryID("")
	for _, item := range items {
		lastScannedID = item.ID
		err := a.authorizer.CheckInventory(ctx, input.Principal, ports.InventoryPermissionView, item.ID)
		if err == nil {
			visible = append(visible, item)
			continue
		}
		if !errors.Is(err, ports.ErrForbidden) {
			return ListInventoriesResult{}, err
		}
	}

	hasMore := len(visible) > limit
	var nextCursor *string
	if hasMore {
		visible = visible[:limit]
		nextCursor = encodeInventoryCursor(input.TenantID, visible[len(visible)-1].ID)
	} else if len(items) == inventoryScanLimit(limit) {
		hasMore = true
		nextCursor = encodeInventoryCursor(input.TenantID, lastScannedID)
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoriesListed,
		Message: "inventories listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"principal_id": input.Principal.ID.String(),
			"limit":        strconv.Itoa(limit),
		},
	})

	return ListInventoriesResult{
		Items:      visible,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func inventoryScanLimit(limit int) int {
	return limit*2 + 1
}

func encodeInventoryCursor(tenantID tenant.ID, id inventory.InventoryID) *string {
	return encodePageCursor("inventories", tenantID.String(), id.String())
}

func decodeInventoryCursor(tenantID tenant.ID, cursor string) (inventory.InventoryID, error) {
	decoded, err := decodePageCursor("inventories", tenantID.String(), cursor)
	if err != nil {
		return inventory.InventoryID(""), err
	}
	if decoded == "" {
		return inventory.InventoryID(""), nil
	}
	id, ok := inventory.NewID(decoded)
	if !ok {
		return inventory.InventoryID(""), ErrInvalidInput
	}
	return id, nil
}

func (a App) recordAuthorizationDenied(ctx context.Context, principal identity.Principal, tenantID tenant.ID) {
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAuthorizationDenied,
		Message: "authorization denied",
		Fields: map[string]string{
			"tenant_id":    tenantID.String(),
			"principal_id": principal.ID.String(),
		},
	})
}

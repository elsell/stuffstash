package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Store struct {
	mu          sync.RWMutex
	tenants     map[tenant.ID]tenant.Tenant
	inventories map[inventory.InventoryID]inventory.Inventory
	outbox      map[string]ports.AuthorizationOutboxEvent
}

func NewStore() *Store {
	return &Store{
		tenants:     map[tenant.ID]tenant.Tenant{},
		inventories: map[inventory.InventoryID]inventory.Inventory{},
		outbox:      map[string]ports.AuthorizationOutboxEvent{},
	}
}

func (s *Store) SaveTenant(_ context.Context, item tenant.Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tenants[item.ID] = item
	return nil
}

func (s *Store) TenantExists(_ context.Context, tenantID tenant.ID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.tenants[tenantID]
	return ok, nil
}

func (s *Store) SaveInventory(_ context.Context, item inventory.Inventory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.inventories[item.ID] = item
	return nil
}

func (s *Store) SaveTenantAndEnqueueOwnerGrant(_ context.Context, eventID string, item tenant.Tenant, principal identity.Principal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tenants[item.ID] = item
	s.outbox[eventID] = ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        ports.AuthorizationOutboxGrantTenantOwner,
		PrincipalID: principal.ID,
		TenantID:    item.ID,
		CreatedAt:   time.Now(),
	}
	return nil
}

func (s *Store) SaveInventoryAndEnqueueOwnerGrant(_ context.Context, eventID string, item inventory.Inventory, tenantID tenant.ID, principal identity.Principal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.inventories[item.ID] = item
	s.outbox[eventID] = ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        ports.AuthorizationOutboxGrantInventoryOwner,
		PrincipalID: principal.ID,
		TenantID:    tenantID,
		InventoryID: item.ID,
		CreatedAt:   time.Now(),
	}
	return nil
}

func (s *Store) ClaimPendingAuthorizationOutboxEvents(_ context.Context, claimID string, limit int, leaseUntil time.Time) ([]ports.AuthorizationOutboxEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limit <= 0 {
		limit = len(s.outbox)
	}
	now := time.Now()
	events := []ports.AuthorizationOutboxEvent{}
	for _, event := range s.outbox {
		if !event.ClaimedUntil.IsZero() && event.ClaimedUntil.After(now) {
			continue
		}
		events = append(events, event)
	}
	sort.Slice(events, func(left int, right int) bool {
		if events[left].CreatedAt.Equal(events[right].CreatedAt) {
			return events[left].ID < events[right].ID
		}
		return events[left].CreatedAt.Before(events[right].CreatedAt)
	})
	if len(events) > limit {
		events = events[:limit]
	}
	for index, event := range events {
		event.ClaimID = claimID
		event.ClaimedUntil = leaseUntil
		s.outbox[event.ID] = event
		events[index] = event
	}
	return events, nil
}

func (s *Store) MarkAuthorizationOutboxEventProcessed(_ context.Context, eventID string, claimID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.outbox[eventID]
	if !ok || event.ClaimID != claimID {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	delete(s.outbox, eventID)
	return nil
}

func (s *Store) MarkAuthorizationOutboxEventFailed(_ context.Context, eventID string, claimID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.outbox[eventID]
	if !ok || event.ClaimID != claimID {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	event.Attempts++
	event.LastError = reason
	event.ClaimID = ""
	event.ClaimedUntil = time.Time{}
	s.outbox[eventID] = event
	return nil
}

func (s *Store) ListInventoriesByTenant(_ context.Context, tenantID inventory.TenantID) ([]inventory.Inventory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []inventory.Inventory{}
	for _, item := range s.inventories {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	return items, nil
}

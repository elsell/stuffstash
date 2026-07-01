package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
	"time"
)

func (s *Store) SaveAuditRecord(_ context.Context, record audit.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[tenant.ID(record.TenantID.String())]; !exists {
		return ports.ErrForbidden
	}
	if record.InventoryID.String() != "" {
		item, ok := s.inventories[inventory.InventoryID(record.InventoryID.String())]
		if !ok || item.TenantID.String() != record.TenantID.String() {
			return ports.ErrForbidden
		}
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.ErrConflict
	}
	s.auditRecords[record.ID] = record
	return nil
}

func (s *Store) ListTenantAuditRecords(_ context.Context, tenantID tenant.ID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []audit.Record{}
	for _, record := range s.auditRecords {
		if record.TenantID.String() == tenantID.String() && auditRecordAfter(record, page.AfterOccurredAt, page.AfterRecordID) {
			items = append(items, record)
		}
	}
	return pagedAuditRecords(items, page.Limit), nil
}

func (s *Store) ListInventoryAuditRecords(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []audit.Record{}
	for _, record := range s.auditRecords {
		if record.TenantID.String() == tenantID.String() && record.InventoryID.String() == inventoryID.String() && auditRecordAfter(record, page.AfterOccurredAt, page.AfterRecordID) {
			items = append(items, record)
		}
	}
	return pagedAuditRecords(items, page.Limit), nil
}

func (s *Store) ListAssetAuditRecords(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, targetID string, request ports.AssetAuditRecordListRequest) ([]audit.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []audit.Record{}
	for _, record := range s.auditRecords {
		if record.TenantID.String() == tenantID.String() &&
			record.InventoryID.String() == inventoryID.String() &&
			record.TargetType == audit.TargetAsset &&
			record.TargetID == targetID {
			items = append(items, record)
		}
	}
	return limitedNewestFirstAuditRecords(items, request.Limit), nil
}

func pagedAuditRecords(items []audit.Record, limit int) []audit.Record {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].Before(items[right])
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func limitedNewestFirstAuditRecords(items []audit.Record, limit int) []audit.Record {
	sort.Slice(items, func(left int, right int) bool {
		return items[right].Before(items[left])
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func auditRecordAfter(record audit.Record, occurredAt time.Time, id audit.ID) bool {
	if occurredAt.IsZero() || id.String() == "" {
		return true
	}
	if record.OccurredAt.After(occurredAt) {
		return true
	}
	return record.OccurredAt.Equal(occurredAt) && record.ID.String() > id.String()
}

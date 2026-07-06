package memory

import (
	"context"
	"sort"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) ImportSourceLinkByKey(_ context.Context, key ports.ImportSourceLinkKey) (ports.ImportSourceLink, bool, error) {
	if err := validateImportSourceLinkKey(key); err != nil {
		return ports.ImportSourceLink{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	link, ok := s.importLinks[importSourceLinkKey(key)]
	if !ok {
		return ports.ImportSourceLink{}, false, nil
	}
	return link, true, nil
}

func (s *Store) SaveImportSourceLink(_ context.Context, link ports.ImportSourceLink) error {
	if err := validateImportSourceLink(link); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	key := importSourceLinkKey(link.Key)
	if _, exists := s.importLinks[key]; exists {
		return ports.ErrConflict
	}
	s.importLinks[key] = link
	return nil
}

func (s *Store) SaveImportJobResource(_ context.Context, record ports.ImportJobResource) error {
	if err := validateImportJobResource(record); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	key := importJobResourceKey(record)
	if _, exists := s.importResources[key]; exists {
		return ports.ErrConflict
	}
	s.importResources[key] = record
	return nil
}

func (s *Store) ListImportJobResources(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, page ports.ImportJobResourcePageRequest) ([]ports.ImportJobResource, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || jobID.String() == "" {
		return nil, ports.ErrInvalidProviderInput
	}
	limit := page.Limit
	if limit <= 0 {
		limit = 50
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := s.listImportJobResourcesLocked(tenantID, inventoryID, jobID)
	if len(records) > limit {
		records = records[:limit]
	}
	return records, nil
}

func (s *Store) ListAllImportJobResources(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) ([]ports.ImportJobResource, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || jobID.String() == "" {
		return nil, ports.ErrInvalidProviderInput
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.listImportJobResourcesLocked(tenantID, inventoryID, jobID), nil
}

func (s *Store) listImportJobResourcesLocked(tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) []ports.ImportJobResource {
	records := []ports.ImportJobResource{}
	for _, record := range s.importResources {
		if record.TenantID == tenantID && record.InventoryID == inventoryID && record.JobID == jobID {
			records = append(records, record)
		}
	}
	sort.SliceStable(records, func(left, right int) bool {
		if !records[left].CreatedAt.Equal(records[right].CreatedAt) {
			return records[left].CreatedAt.Before(records[right].CreatedAt)
		}
		if records[left].ResourceType != records[right].ResourceType {
			return records[left].ResourceType < records[right].ResourceType
		}
		return records[left].ResourceID < records[right].ResourceID
	})
	return records
}

func (s *Store) DeleteImportSourceLinksForJob(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (int, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || jobID.String() == "" {
		return 0, ports.ErrInvalidProviderInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	deleted := 0
	for key, link := range s.importLinks {
		if link.Key.TenantID == tenantID && link.Key.InventoryID == inventoryID && link.JobID == jobID {
			delete(s.importLinks, key)
			deleted++
		}
	}
	return deleted, nil
}

func validateImportSourceLink(link ports.ImportSourceLink) error {
	if err := validateImportSourceLinkKey(link.Key); err != nil {
		return err
	}
	if link.ResourceType == "" || strings.TrimSpace(link.ResourceID) == "" || link.JobID.String() == "" || link.CreatedAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func validateImportSourceLinkKey(key ports.ImportSourceLinkKey) error {
	if key.TenantID.String() == "" || key.InventoryID.String() == "" || key.SourceType == "" || strings.TrimSpace(key.SourceInstanceKey) == "" || key.SourceEntityType == "" || strings.TrimSpace(key.SourceEntityID) == "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func validateImportJobResource(record ports.ImportJobResource) error {
	if record.TenantID.String() == "" || record.InventoryID.String() == "" || record.JobID.String() == "" || record.ResourceType == "" || strings.TrimSpace(record.ResourceID) == "" || record.SourceType == "" || strings.TrimSpace(record.SourceInstanceKey) == "" || record.SourceEntityType == "" || strings.TrimSpace(record.SourceEntityID) == "" || record.CreatedAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func importSourceLinkKey(key ports.ImportSourceLinkKey) string {
	return key.TenantID.String() + "\x00" + key.InventoryID.String() + "\x00" + string(key.SourceType) + "\x00" + strings.TrimSpace(key.SourceInstanceKey) + "\x00" + string(key.SourceEntityType) + "\x00" + strings.TrimSpace(key.SourceEntityID)
}

func importJobResourceKey(record ports.ImportJobResource) string {
	return record.TenantID.String() + "\x00" + record.InventoryID.String() + "\x00" + record.JobID.String() + "\x00" + string(record.ResourceType) + "\x00" + strings.TrimSpace(record.ResourceID)
}

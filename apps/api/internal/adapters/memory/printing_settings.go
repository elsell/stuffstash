package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) GetPrintSettings(_ context.Context, scope printing.Scope) (printing.InventoryPrintSettings, error) {
	if scope.TenantID == "" || scope.InventoryID == "" {
		return printing.InventoryPrintSettings{}, ports.ErrPrintNotFound
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.printSettings[scope]
	if !found {
		return printing.DefaultInventoryPrintSettings(scope), nil
	}
	return value, nil
}
func (s *Store) SavePrintSettings(_ context.Context, next printing.InventoryPrintSettings, expected uint64, destination *printing.SettingsDestination, record audit.Record) (printing.InventoryPrintSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.printingInventoryExistsLocked(next.Scope) {
		return printing.InventoryPrintSettings{}, ports.ErrPrintNotFound
	}
	if !next.ValidReplacement(expected) {
		return printing.InventoryPrintSettings{}, ports.ErrPrintConflict
	}
	current := s.printSettings[next.Scope]
	if current.Revision != expected {
		return printing.InventoryPrintSettings{}, ports.ErrPrintConflict
	}
	if next.DefaultPrinterID != "" {
		printer, found := s.printingPrinters[next.DefaultPrinterID]
		if destination == nil || !found || destination.ID != next.DefaultPrinterID || printer.Scope != next.Scope || printer.Retired || printer.Revision != destination.Revision || printer.MediaFingerprint != destination.MediaFingerprint {
			return printing.InventoryPrintSettings{}, ports.ErrPrintConflict
		}
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return printing.InventoryPrintSettings{}, ports.ErrPrintConflict
	}
	if s.printSettings == nil {
		s.printSettings = make(map[printing.Scope]printing.InventoryPrintSettings)
	}
	s.printSettings[next.Scope] = next
	s.auditRecords[record.ID] = record
	return next, nil
}

var _ ports.PrintSettingsRepository = (*Store)(nil)

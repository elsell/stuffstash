package gormstore

import (
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"gorm.io/gorm"
)

// Caller holds the connector and printer locks, shared with health updates.
func printingDispatchReady(tx *gorm.DB, authority printing.ConsumerAuthority, now time.Time, maxAge time.Duration) (bool, error) {
	var connector printingConnectorModel
	if err := tx.Where(&printingConnectorModel{ID: string(authority.ConnectorID), TenantID: authority.Scope.TenantID, InventoryID: authority.Scope.InventoryID}).First(&connector).Error; err != nil {
		return false, err
	}
	if connector.LastSeenAt == nil || now.Before(*connector.LastSeenAt) || now.Sub(*connector.LastSeenAt) > maxAge {
		return false, nil
	}
	var report printingReportModel
	err := tx.Where(&printingReportModel{ConnectorID: connector.ID, PrinterID: string(authority.PrinterID), TenantID: authority.Scope.TenantID, InventoryID: authority.Scope.InventoryID}).First(&report).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return report.State == string(printing.PrinterReady) && !now.Before(report.ReportedAt) && now.Sub(report.ReportedAt) <= maxAge, nil
}

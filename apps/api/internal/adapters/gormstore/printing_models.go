package gormstore

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"time"
)

type printingPrinterModel struct {
	RequestKey                    string `gorm:"uniqueIndex:idx_printer_request,priority:3"`
	RequestFingerprint            string
	ID                            string `gorm:"primaryKey"`
	TenantID                      string `gorm:"uniqueIndex:idx_printer_request,priority:1"`
	InventoryID                   string `gorm:"uniqueIndex:idx_printer_request,priority:2"`
	Name, AdapterID, DeviceID     string
	MediaJSON                     []byte
	MediaFingerprint              string
	Revision                      uint64
	Retired                       bool
	ActiveJobID, ReservationState string
	Readiness, ReadinessReason    string
	ReportedAt                    *time.Time
	CreatedAt, UpdatedAt          time.Time
}

func (printingPrinterModel) TableName() string { return "printers" }
func (m printingPrinterModel) domain() (printing.Printer, error) {
	var media printing.MediaSnapshot
	if err := json.Unmarshal(m.MediaJSON, &media); err != nil {
		return printing.Printer{}, err
	}
	return printing.Printer{RequestKey: m.RequestKey, RequestFingerprint: m.RequestFingerprint, ID: printing.PrinterID(m.ID), Scope: printing.Scope{TenantID: m.TenantID, InventoryID: m.InventoryID}, Name: m.Name, AdapterID: m.AdapterID, DeviceID: m.DeviceID, Media: media, MediaFingerprint: m.MediaFingerprint, Revision: m.Revision, Retired: m.Retired, ActiveJobID: m.ActiveJobID, ReservationState: m.ReservationState, Readiness: printing.PrinterReadiness(m.Readiness), ReadinessReason: m.ReadinessReason, ReportedAt: m.ReportedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}
func printingPrinterFromDomain(p printing.Printer) (printingPrinterModel, error) {
	media, err := json.Marshal(p.Media)
	return printingPrinterModel{RequestKey: p.RequestKey, RequestFingerprint: p.RequestFingerprint, ID: string(p.ID), TenantID: p.Scope.TenantID, InventoryID: p.Scope.InventoryID, Name: p.Name, AdapterID: p.AdapterID, DeviceID: p.DeviceID, MediaJSON: media, MediaFingerprint: p.MediaFingerprint, Revision: p.Revision, Retired: p.Retired, ActiveJobID: p.ActiveJobID, ReservationState: p.ReservationState, Readiness: string(p.Readiness), ReadinessReason: p.ReadinessReason, ReportedAt: p.ReportedAt, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}, err
}

type printingConnectorModel struct {
	ReportJSON                                            []byte
	ReportReceivedAt                                      *time.Time
	PendingCredentialHash                                 string `gorm:"uniqueIndex:idx_print_connector_pending_credential,where:pending_credential_hash <> ''"`
	PendingCredentialVersion                              uint64
	PendingCredentialExpiresAt, PendingActivationDeadline time.Time
	PendingPublicKey                                      []byte
	ID                                                    string `gorm:"primaryKey"`
	TenantID, InventoryID, ServiceAccountID, Name, State  string
	PublicKey                                             []byte
	CredentialHash                                        string `gorm:"uniqueIndex:idx_print_connector_credential,where:credential_hash <> ''"`
	CredentialVersion                                     uint64
	CredentialExpiresAt, ActivationDeadline               time.Time
	Generation, SyncedGeneration                          uint64
	LastSeenAt                                            *time.Time
	CreatedAt, UpdatedAt                                  time.Time
}

func (printingConnectorModel) TableName() string { return "print_connectors" }
func (m printingConnectorModel) domain() printing.Connector {
	var report *printing.ConnectorReport
	if len(m.ReportJSON) > 0 {
		if json.Unmarshal(m.ReportJSON, &report) != nil {
			report = nil
		}
	}
	return printing.Connector{Report: report, ReportReceivedAt: m.ReportReceivedAt, PendingCredentialHash: m.PendingCredentialHash, PendingCredentialVersion: m.PendingCredentialVersion, PendingCredentialExpiresAt: m.PendingCredentialExpiresAt, PendingActivationDeadline: m.PendingActivationDeadline, PendingPublicKey: append([]byte(nil), m.PendingPublicKey...), ID: printing.ConnectorID(m.ID), Scope: printing.Scope{TenantID: m.TenantID, InventoryID: m.InventoryID}, ServiceAccountID: printing.ServiceAccountID(m.ServiceAccountID), Name: m.Name, State: printing.ConnectorState(m.State), PublicKey: append([]byte(nil), m.PublicKey...), CredentialHash: m.CredentialHash, CredentialVersion: m.CredentialVersion, CredentialExpiresAt: m.CredentialExpiresAt, ActivationDeadline: m.ActivationDeadline, Generation: m.Generation, SyncedGeneration: m.SyncedGeneration, LastSeenAt: m.LastSeenAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

type printingBindingModel struct {
	ConnectorID                     string `gorm:"primaryKey"`
	PrinterID                       string `gorm:"primaryKey"`
	TenantID, InventoryID, DeviceID string
	Generation, SyncedGeneration    uint64
	Revoked                         bool
}

func (printingBindingModel) TableName() string { return "print_connector_bindings" }
func (m printingBindingModel) domain() printing.PrinterBinding {
	return printing.PrinterBinding{Scope: printing.Scope{TenantID: m.TenantID, InventoryID: m.InventoryID}, ConnectorID: printing.ConnectorID(m.ConnectorID), PrinterID: printing.PrinterID(m.PrinterID), DeviceID: m.DeviceID, Generation: m.Generation, SyncedGeneration: m.SyncedGeneration, Revoked: m.Revoked}
}

type printingReportModel struct {
	ConnectorID                          string `gorm:"primaryKey"`
	PrinterID                            string `gorm:"primaryKey"`
	TenantID, InventoryID, State, Reason string
	ReportedAt                           time.Time
}

func (printingReportModel) TableName() string { return "print_printer_reports" }

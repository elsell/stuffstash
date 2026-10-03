package printregistry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strings"
	"time"
)

type Service struct {
	Settings              ports.PrintSettingsRepository
	SelectionValidator    ports.LabelSelectionValidator
	Health                ports.ConnectorRepository
	PrintingAuthorization ports.PrintingAuthorization
	ReportMaxAge          time.Duration
	Authorizer            ports.Authorizer
	Inventories           ports.InventoryRepository
	Printers              ports.PrinterRepository
	Catalog               ports.PrinterCatalog
	Audit                 ports.AuditRepository
	IDs                   ports.IDGenerator
	Clock                 ports.Clock
	Observer              ports.Observer
}
type Actor struct {
	Principal identity.Principal
	Scope     printing.Scope
	RequestID string
}
type RegisterPrinter struct {
	Actor                                 Actor
	RequestKey, Name, AdapterID, PresetID string
	PresetVersion                         uint32
}
type UpdatePrinter struct {
	Actor         Actor
	ID            printing.PrinterID
	Revision      uint64
	Name          *string
	PresetID      *string
	PresetVersion *uint32
	Retired       *bool
}

func (s Service) access(ctx context.Context, a Actor, permission ports.InventoryPermission) error {
	if a.Principal.ID == "" {
		return apperrors.ErrUnauthenticated
	}
	if a.Scope.TenantID == "" || a.Scope.InventoryID == "" {
		return apperrors.ErrInvalidInput
	}
	i, found, err := s.Inventories.InventoryByID(ctx, tenant.ID(a.Scope.TenantID), inventory.InventoryID(a.Scope.InventoryID))
	if err != nil {
		return err
	}
	if !found || !i.IsActive() {
		return apperrors.ErrNotFound
	}
	return s.Authorizer.CheckInventory(ctx, a.Principal, permission, inventory.InventoryID(a.Scope.InventoryID))
}
func registryError(err error) error {
	if errors.Is(err, ports.ErrPrintNotFound) {
		return apperrors.ErrNotFound
	}
	if errors.Is(err, ports.ErrPrintConflict) {
		return apperrors.ErrConflict
	}
	return err
}
func digest(v any) string {
	data, _ := json.Marshal(v)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
func validPrinterName(name string) bool {
	return name != "" && len(name) <= 100 && !strings.ContainsAny(name, "\r\n\x00")
}
func (s Service) auditRecord(a Actor, action audit.Action, id string) (audit.Record, error) {
	record, ok := audit.NewRecord(audit.ID(s.IDs.NewID()), audit.TenantID(a.Scope.TenantID), audit.InventoryID(a.Scope.InventoryID), audit.PrincipalID(a.Principal.ID), action, audit.SourceAPI, audit.TargetInventory, a.Scope.InventoryID, s.Clock.Now(), a.RequestID, map[string]string{"printer_id": id})
	if !ok {
		return audit.Record{}, errors.New("invalid printer audit record")
	}
	return record, nil
}
func (s Service) observe(ctx context.Context, event ports.EventName) {
	if s.Observer != nil {
		s.Observer.Record(ctx, ports.Event{Name: event})
	}
}
func (s Service) Register(ctx context.Context, input RegisterPrinter) (printing.Printer, bool, error) {
	if err := s.access(ctx, input.Actor, ports.InventoryPermissionConfigure); err != nil {
		return printing.Printer{}, false, err
	}
	input.Name = strings.TrimSpace(input.Name)
	if !validPrinterName(input.Name) || len(input.RequestKey) < 1 || len(input.RequestKey) > 200 {
		return printing.Printer{}, false, apperrors.ErrInvalidInput
	}
	media, err := s.Catalog.ResolvePrinterMedia(input.AdapterID, input.PresetID, input.PresetVersion)
	if err != nil {
		return printing.Printer{}, false, apperrors.ErrInvalidInput
	}
	now := s.Clock.Now()
	p := printing.Printer{ID: printing.PrinterID(s.IDs.NewID()), Scope: input.Actor.Scope, Name: input.Name, AdapterID: input.AdapterID, Media: media, MediaFingerprint: media.Fingerprint(), Revision: 1, Readiness: printing.PrinterUnknown, CreatedAt: now, UpdatedAt: now, RequestKey: digest([]string{input.Actor.Scope.TenantID, input.Actor.Scope.InventoryID, string(input.Actor.Principal.ID), input.RequestKey}), RequestFingerprint: digest([]any{input.Name, input.AdapterID, media})}
	record, err := s.auditRecord(input.Actor, audit.ActionPrinterRegistered, string(p.ID))
	if err != nil {
		return printing.Printer{}, false, err
	}
	result, created, err := s.Printers.CreatePrinter(ctx, p, record)
	if err != nil {
		return printing.Printer{}, false, registryError(err)
	}
	if created {
		s.observe(ctx, ports.EventPrinterRegistered)
	}
	return result, created, nil
}
func (s Service) Get(ctx context.Context, a Actor, id printing.PrinterID) (printing.Printer, error) {
	if err := s.access(ctx, a, ports.InventoryPermissionView); err != nil {
		return printing.Printer{}, err
	}
	p, err := s.Printers.GetPrinter(ctx, a.Scope, id)
	if err != nil {
		return printing.Printer{}, registryError(err)
	}
	record, err := s.auditRecord(a, audit.ActionPrinterViewed, string(id))
	if err != nil {
		return printing.Printer{}, err
	}
	if err := s.Audit.SaveAuditRecord(ctx, record); err != nil {
		return printing.Printer{}, err
	}
	return s.withHealth(ctx, p)
}

type PrinterPage struct {
	Items      []printing.Printer
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (s Service) List(ctx context.Context, a Actor, limit int, cursor string) (PrinterPage, error) {
	if err := s.access(ctx, a, ports.InventoryPermissionView); err != nil {
		return PrinterPage{}, err
	}
	if limit < 1 || limit > 100 {
		return PrinterPage{}, apperrors.ErrInvalidInput
	}
	scope := a.Scope.TenantID + ":" + a.Scope.InventoryID
	after, err := appsupport.DecodePageCursor("printers", scope, cursor)
	if err != nil {
		return PrinterPage{}, err
	}
	values, err := s.Printers.ListPrinters(ctx, a.Scope, limit+1, after)
	if err != nil {
		return PrinterPage{}, registryError(err)
	}
	for i := range values {
		projected, err := s.withHealth(ctx, values[i])
		if err != nil {
			return PrinterPage{}, err
		}
		values[i] = projected
	}
	result := PrinterPage{Items: values, Limit: limit, HasMore: len(values) > limit}
	if result.HasMore {
		result.Items = values[:limit]
		result.NextCursor = appsupport.EncodePageCursor("printers", scope, string(result.Items[len(result.Items)-1].ID))
	}
	record, err := s.auditRecord(a, audit.ActionPrintersListed, "")
	if err != nil {
		return PrinterPage{}, err
	}
	if err := s.Audit.SaveAuditRecord(ctx, record); err != nil {
		return PrinterPage{}, err
	}
	return result, nil
}
func (s Service) Update(ctx context.Context, input UpdatePrinter) (printing.Printer, error) {
	if err := s.access(ctx, input.Actor, ports.InventoryPermissionConfigure); err != nil {
		return printing.Printer{}, err
	}
	if input.Revision == 0 || (input.Name == nil && input.PresetID == nil && input.Retired == nil) || (input.PresetID == nil) != (input.PresetVersion == nil) {
		return printing.Printer{}, apperrors.ErrInvalidInput
	}
	change := func(p *printing.Printer) error {
		if input.Name != nil {
			name := strings.TrimSpace(*input.Name)
			if !validPrinterName(name) {
				return apperrors.ErrInvalidInput
			}
			p.Name = name
		}
		if input.PresetID != nil {
			media, err := s.Catalog.ResolvePrinterMedia(p.AdapterID, *input.PresetID, *input.PresetVersion)
			if err != nil {
				return apperrors.ErrInvalidInput
			}
			if media.Fingerprint() != p.MediaFingerprint && (p.ReservationState == printing.PrinterReservationPrinting || p.ReservationState == printing.PrinterReservationUncertain) {
				return ports.ErrPrintConflict
			}
			p.Media = media
			p.MediaFingerprint = media.Fingerprint()
		}
		if input.Retired != nil {
			p.Retired = *input.Retired
		}
		p.Revision++
		p.UpdatedAt = s.Clock.Now()
		return nil
	}
	makeAudit := func(p printing.Printer) (audit.Record, error) {
		return s.auditRecord(input.Actor, audit.ActionPrinterUpdated, string(p.ID))
	}
	p, err := s.Printers.UpdatePrinter(ctx, input.Actor.Scope, input.ID, input.Revision, change, makeAudit, &ports.PrinterRetirement{
		Now: s.Clock.Now(),
		JobAudit: func(_, job printing.Job) (audit.Record, error) {
			record, err := s.auditRecord(input.Actor, audit.ActionPrintJobCanceled, string(input.ID))
			record.TargetID = string(job.ID)
			record.TargetType = audit.TargetPrintJob
			return record, err
		},
		SettingsAudit: func(printing.InventoryPrintSettings) (audit.Record, error) {
			return s.auditRecord(input.Actor, audit.ActionPrintSettingsUpdated, string(input.ID))
		},
	})
	if err != nil {
		return printing.Printer{}, registryError(err)
	}
	s.observe(ctx, ports.EventPrinterUpdated)
	return p, nil
}

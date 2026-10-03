package gormstore

import (
	"context"
	"crypto/subtle"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
	"time"
)

func lockPrintConnector(tx *gorm.DB, scope printing.Scope, id printing.ConnectorID) (printingConnectorModel, error) {
	if scope.TenantID == "" || scope.InventoryID == "" || id == "" {
		return printingConnectorModel{}, ports.ErrPrintNotFound
	}
	var model printingConnectorModel
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printingConnectorModel{ID: string(id), TenantID: scope.TenantID, InventoryID: scope.InventoryID}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ports.ErrPrintNotFound
	}
	return model, err
}
func printRegistration(tx *gorm.DB, c printingConnectorModel) (ports.ConnectorRegistration, error) {
	result := ports.ConnectorRegistration{Connector: c.domain(), Bindings: []printing.PrinterBinding{}}
	var bindings []printingBindingModel
	if err := tx.Where(&printingBindingModel{ConnectorID: c.ID, TenantID: c.TenantID, InventoryID: c.InventoryID}).Order(clause.OrderByColumn{Column: clause.Column{Name: "printer_id"}}).Find(&bindings).Error; err != nil {
		return result, err
	}
	for _, b := range bindings {
		result.Bindings = append(result.Bindings, b.domain())
	}
	return result, nil
}
func (s Store) GetPrintConnector(ctx context.Context, scope printing.Scope, id printing.ConnectorID) (ports.ConnectorRegistration, error) {
	if scope.TenantID == "" || scope.InventoryID == "" || id == "" {
		return ports.ConnectorRegistration{}, ports.ErrPrintNotFound
	}
	var model printingConnectorModel
	tx := s.db.WithContext(ctx)
	if err := tx.Where(&printingConnectorModel{ID: string(id), TenantID: scope.TenantID, InventoryID: scope.InventoryID}).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ConnectorRegistration{}, ports.ErrPrintNotFound
		}
		return ports.ConnectorRegistration{}, err
	}
	return printRegistration(tx, model)
}
func (s Store) ListPrintConnectors(ctx context.Context, scope printing.Scope, limit int, after string) ([]ports.ConnectorRegistration, error) {
	out := []ports.ConnectorRegistration{}
	if scope.TenantID == "" || scope.InventoryID == "" {
		return out, ports.ErrPrintNotFound
	}
	tx := s.db.WithContext(ctx)
	query := tx.Where(&printingConnectorModel{TenantID: scope.TenantID, InventoryID: scope.InventoryID}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(limit)
	if after != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: after})
	}
	var models []printingConnectorModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	for _, model := range models {
		registration, err := printRegistration(tx, model)
		if err != nil {
			return nil, err
		}
		out = append(out, registration)
	}
	return out, nil
}
func (s Store) FindPrintConnectorCredential(ctx context.Context, hash string) (printing.Connector, error) {
	if hash == "" {
		return printing.Connector{}, ports.ErrPrintDenied
	}
	var model printingConnectorModel
	if err := s.db.WithContext(ctx).Where(clause.Or(clause.Eq{Column: clause.Column{Name: "credential_hash"}, Value: hash}, clause.Eq{Column: clause.Column{Name: "pending_credential_hash"}, Value: hash})).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return printing.Connector{}, ports.ErrPrintDenied
		}
		return printing.Connector{}, err
	}
	c := model.domain()
	if c.PendingCredentialHash == hash && c.State != printing.ConnectorRevoked {
		return c.PendingCredentialIdentity(), nil
	}
	return c, nil
}
func (s Store) SynchronizePrintConnector(ctx context.Context, scope printing.Scope, id printing.ConnectorID, sync ports.ConnectorAuthorizationSync) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		c, err := lockPrintConnector(tx, scope, id)
		if err != nil {
			return err
		}
		registration, err := printRegistration(tx, c)
		if err != nil {
			return err
		}
		printers := []printing.Printer{}
		for _, b := range registration.Bindings {
			model, err := printerByScope(tx, scope, b.PrinterID)
			if err != nil {
				return err
			}
			p, err := model.domain()
			if err != nil {
				return err
			}
			printers = append(printers, p)
		}
		if err := sync(ctx, registration, printers); err != nil {
			return err
		}
		c.SyncedGeneration = c.Generation
		if err := tx.Save(&c).Error; err != nil {
			return err
		}
		for _, b := range registration.Bindings {
			b.SyncedGeneration = b.Generation
			model := printingBindingFromDomain(b)
			if err := tx.Save(&model).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (s Store) PendingPrintConnectorScopes(ctx context.Context, limit int) ([]printing.Connector, error) {
	var models []printingConnectorModel
	if err := s.db.WithContext(ctx).Where(clause.Neq{Column: clause.Column{Name: "generation"}, Value: clause.Column{Name: "synced_generation"}}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]printing.Connector, 0, len(models))
	for _, m := range models {
		out = append(out, m.domain())
	}
	return out, nil
}
func (s Store) HeartbeatPrintConnector(ctx context.Context, authenticated printing.Connector, now time.Time, makeAudit ports.ConnectorActivationAudit, report *printing.ConnectorReport) (printing.Connector, error) {
	if report != nil && !report.Valid() {
		return printing.Connector{}, ports.ErrPrintConflict
	}
	var result printing.Connector
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model, err := lockPrintConnector(tx, authenticated.Scope, authenticated.ID)
		if err != nil {
			return err
		}
		c := model.domain()
		if c.ServiceAccountID != authenticated.ServiceAccountID || c.Generation != c.SyncedGeneration || c.State == printing.ConnectorRevoked {
			return ports.ErrPrintDenied
		}
		activating := c.State == printing.ConnectorAwaitingActivation
		pending := c.PendingCredentialHash != "" && c.PendingCredentialVersion == authenticated.CredentialVersion && subtle.ConstantTimeCompare([]byte(c.PendingCredentialHash), []byte(authenticated.CredentialHash)) == 1
		if pending {
			if !c.PendingCredentialExpiresAt.After(now) || !c.PendingActivationDeadline.After(now) {
				return ports.ErrPrintDenied
			}
			c.ActivatePending()
		} else {
			if c.CredentialVersion != authenticated.CredentialVersion || subtle.ConstantTimeCompare([]byte(c.CredentialHash), []byte(authenticated.CredentialHash)) != 1 || !c.CredentialExpiresAt.After(now) {
				return ports.ErrPrintDenied
			}
			if c.State == printing.ConnectorAwaitingActivation && c.ActivationDeadline.After(now) {
				c.State = printing.ConnectorActive
			} else if c.State != printing.ConnectorActive {
				return ports.ErrPrintDenied
			}
		}
		if pending || activating || (report != nil && !reflect.DeepEqual(c.Report, report)) {
			if makeAudit == nil {
				return ports.ErrPrintConflict
			}
			record, err := makeAudit(c, pending)
			if err != nil {
				return err
			}
			if err := createAuditRecord(tx, record); err != nil {
				return err
			}
		}
		if report != nil {
			c.Report = report.Clone()
			c.ReportReceivedAt = &now
		}
		c.LastSeenAt = &now
		c.UpdatedAt = now
		model = printingConnectorFromDomain(c)
		if err := tx.Save(&model).Error; err != nil {
			return err
		}
		result = c
		return nil
	})
	return result, err
}
func (s Store) ReportPrintPrinter(ctx context.Context, authority printing.ConsumerAuthority, report printing.PrinterReport, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := printingConsumerFence(tx, authority, now); err != nil {
			return err
		}
		if report.Scope != authority.Scope || report.ConnectorID != authority.ConnectorID || report.PrinterID != authority.PrinterID {
			return ports.ErrPrintDenied
		}
		model := printingReportModel{ConnectorID: string(report.ConnectorID), PrinterID: string(report.PrinterID), TenantID: report.Scope.TenantID, InventoryID: report.Scope.InventoryID, State: string(report.State), Reason: report.Reason, ReportedAt: now}
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "connector_id"}, {Name: "printer_id"}}, DoUpdates: clause.AssignmentColumns([]string{"state", "reason", "reported_at"})}).Create(&model).Error
	})
}

var _ ports.ConnectorRepository = Store{}

func (s Store) UpdatePrintConnector(ctx context.Context, scope printing.Scope, id printing.ConnectorID, generation uint64, change ports.ConnectorMutation, makeAudit ports.ConnectorAudit) (ports.ConnectorRegistration, error) {
	var result ports.ConnectorRegistration
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		c, err := lockPrintConnector(tx, scope, id)
		if err != nil {
			return err
		}
		if c.Generation != generation {
			return ports.ErrPrintConflict
		}
		r, err := printRegistration(tx, c)
		if err != nil {
			return err
		}
		if err := change(&r); err != nil {
			return err
		}
		if r.Connector.ID != id || r.Connector.Scope != scope || string(r.Connector.ServiceAccountID) != c.ServiceAccountID {
			return ports.ErrPrintConflict
		}
		record, err := makeAudit(r)
		if err != nil {
			return err
		}
		model := printingConnectorFromDomain(r.Connector)
		if err := tx.Save(&model).Error; err != nil {
			return err
		}
		for _, b := range r.Bindings {
			bm := printingBindingFromDomain(b)
			if err := tx.Save(&bm).Error; err != nil {
				return err
			}
		}
		if err := createAuditRecord(tx, record); err != nil {
			return err
		}
		result = r
		return nil
	})
	return result, err
}

func (s Store) ListPrintPrinterHealth(ctx context.Context, scope printing.Scope, id printing.PrinterID) ([]ports.PrinterHealthReport, error) {
	result := []ports.PrinterHealthReport{}
	if scope.TenantID == "" || scope.InventoryID == "" || id == "" {
		return result, ports.ErrPrintNotFound
	}
	tx := s.db.WithContext(ctx)
	var reports []printingReportModel
	if err := tx.Where(&printingReportModel{TenantID: scope.TenantID, InventoryID: scope.InventoryID, PrinterID: string(id)}).Find(&reports).Error; err != nil {
		return nil, err
	}
	for _, report := range reports {
		var c printingConnectorModel
		if err := tx.Where(&printingConnectorModel{ID: report.ConnectorID, TenantID: scope.TenantID, InventoryID: scope.InventoryID}).First(&c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		var b printingBindingModel
		if err := tx.Where(&printingBindingModel{ConnectorID: report.ConnectorID, PrinterID: string(id), TenantID: scope.TenantID, InventoryID: scope.InventoryID}).First(&b).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		result = append(result, ports.PrinterHealthReport{Connector: c.domain(), Binding: b.domain(), Report: printing.PrinterReport{Scope: scope, PrinterID: id, ConnectorID: printing.ConnectorID(report.ConnectorID), State: printing.PrinterReadiness(report.State), Reason: report.Reason, ReportedAt: report.ReportedAt}})
	}
	return result, nil
}

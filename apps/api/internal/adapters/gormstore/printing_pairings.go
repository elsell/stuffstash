package gormstore

import (
	"context"
	"crypto/subtle"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func (s Store) CreatePrintPairing(ctx context.Context, p printing.Pairing) error {
	model, err := printingPairingFromDomain(p)
	if err != nil {
		return err
	}
	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrPrintConflict
	}
	return nil
}
func (s Store) GetPrintPairing(ctx context.Context, id printing.PairingID) (printing.Pairing, error) {
	if id == "" {
		return printing.Pairing{}, ports.ErrPrintNotFound
	}
	var model printingPairingModel
	err := s.db.WithContext(ctx).Where(&printingPairingModel{ID: string(id)}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ports.ErrPrintNotFound
	}
	if err != nil {
		return printing.Pairing{}, err
	}
	return model.domain()
}
func lockPrintPairing(tx *gorm.DB, id printing.PairingID) (printing.Pairing, error) {
	if id == "" {
		return printing.Pairing{}, ports.ErrPrintDenied
	}
	var model printingPairingModel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&printingPairingModel{ID: string(id)}).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return printing.Pairing{}, ports.ErrPrintDenied
		}
		return printing.Pairing{}, err
	}
	return model.domain()
}
func (s Store) ApprovePrintPairing(ctx context.Context, input ports.PairingApproval) (ports.ConnectorRegistration, error) {
	var result ports.ConnectorRegistration
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		p, err := lockPrintPairing(tx, input.PairingID)
		if err != nil {
			return err
		}
		c := input.Registration.Connector
		if subtle.ConstantTimeCompare([]byte(p.CodeHash), []byte(input.CodeHash)) != 1 || !p.Approve(c.Scope, c.ID, input.Now) {
			return ports.ErrPrintDenied
		}
		devices := map[string]bool{}
		printers := map[printing.PrinterID]bool{}
		for _, b := range input.Registration.Bindings {
			pm, err := printerByScope(tx, c.Scope, b.PrinterID)
			if err != nil {
				return err
			}
			if pm.Retired || b.Scope != c.Scope || b.ConnectorID != c.ID || b.DeviceID == "" || devices[b.DeviceID] || printers[b.PrinterID] {
				return ports.ErrPrintConflict
			}
			devices[b.DeviceID] = true
			printers[b.PrinterID] = true
		}
		cm := printingConnectorFromDomain(c)
		if err := tx.Create(&cm).Error; err != nil {
			return err
		}
		for _, b := range input.Registration.Bindings {
			bm := printingBindingFromDomain(b)
			if err := tx.Create(&bm).Error; err != nil {
				return err
			}
		}
		pm, err := printingPairingFromDomain(p)
		if err != nil {
			return err
		}
		if err := tx.Save(&pm).Error; err != nil {
			return err
		}
		if err := createAuditRecord(tx, input.Audit); err != nil {
			return err
		}
		result = input.Registration
		return nil
	})
	return result, err
}
func (s Store) ConsumePrintPairing(ctx context.Context, id printing.PairingID, now time.Time, hash string, expires, activation time.Time) (printing.Connector, error) {
	var result printing.Connector
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		p, err := lockPrintPairing(tx, id)
		if err != nil {
			return err
		}
		if !p.Consume(now) {
			return ports.ErrPrintDenied
		}
		cm, err := lockPrintConnector(tx, p.Scope, p.ConnectorID)
		if err != nil {
			return err
		}
		c := cm.domain()
		if c.State != printing.ConnectorPending || c.Generation != c.SyncedGeneration {
			return ports.ErrPrintDenied
		}
		if hash == "" || !expires.After(now) || !activation.After(now) {
			return ports.ErrPrintConflict
		}
		c.CredentialHash = hash
		c.CredentialExpiresAt = expires
		c.ActivationDeadline = activation
		c.State = printing.ConnectorAwaitingActivation
		c.UpdatedAt = now
		cm = printingConnectorFromDomain(c)
		if err := tx.Save(&cm).Error; err != nil {
			return err
		}
		pm, err := printingPairingFromDomain(p)
		if err != nil {
			return err
		}
		if err := tx.Save(&pm).Error; err != nil {
			return err
		}
		result = c
		return nil
	})
	return result, err
}

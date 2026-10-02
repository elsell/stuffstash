package gormstore

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ArchiveRestorePublisher struct {
	store      Store
	clock      ports.Clock
	maxRecords int
}

func NewArchiveRestorePublisher(store Store, clock ports.Clock, maxRecords int) ArchiveRestorePublisher {
	return ArchiveRestorePublisher{store: store, clock: clock, maxRecords: maxRecords}
}

var _ ports.ArchiveRestoreUnitOfWork = ArchiveRestorePublisher{}

func (p ArchiveRestorePublisher) PublishArchiveRestore(ctx context.Context, input ports.ArchiveRestorePublication) (archivejob.Record, error) {
	empty := archivejob.Record{}
	d := input.Plan.Document
	if p.clock == nil || p.maxRecords <= 0 || input.OwnerGrantEventID == "" || len(input.AuditRecords) == 0 || input.Job.Kind != archivejob.Restore || input.Job.TenantID != d.TenantID || input.Job.DestinationInventoryID != d.InventoryID || input.Job.DestinationName != d.InventoryName {
		return empty, archivejob.ErrInvalid
	}
	if err := dataportability.ValidateArchiveDocument(ctx, d, p.maxRecords); err != nil {
		return empty, err
	}
	if err := dataportability.ValidateArchiveRestoreAudits(d, input.AuditRecords, input.Job.PrincipalID, input.Job.ID); err != nil {
		return empty, err
	}

	var completed archivejob.Record
	err := p.store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var stored archiveJobModel
		err := archiveJobQuery(tx, ports.ArchiveJobScope{TenantID: d.TenantID}).Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Eq{Column: "id", Value: input.Job.ID}).Where(clause.Eq{Column: "revision", Value: input.Job.Revision}).First(&stored).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrArchiveJobConflict
		}
		if err != nil {
			return err
		}
		job, err := archiveJobRecordFromModel(stored)
		if err != nil {
			return err
		}
		if job != input.Job {
			return ports.ErrArchiveJobConflict
		}
		if _, err = job.Complete(input.Job.LeaseToken, p.clock.Now(), ""); err != nil {
			return err
		}
		if err = tx.Create(&inventoryModel{ID: d.InventoryID, TenantID: d.TenantID, Name: d.InventoryName, LifecycleState: "active"}).Error; err != nil {
			return err
		}
		if err = restoreArchiveSchema(tx, d); err != nil {
			return err
		}
		if err = restoreArchiveAssets(tx, d, job.PrincipalID); err != nil {
			return err
		}
		if err = restoreArchiveTypeLifecycles(tx, d); err != nil {
			return err
		}
		for _, record := range input.AuditRecords {
			if err = createAuditRecord(tx, record); err != nil {
				return err
			}
		}
		if err = tx.Create(&authorizationOutboxEventModel{ID: input.OwnerGrantEventID, Kind: string(ports.AuthorizationOutboxGrantInventoryOwner), PrincipalID: job.PrincipalID, TenantID: d.TenantID, InventoryID: &d.InventoryID}).Error; err != nil {
			return err
		}
		completed, err = job.Complete(job.LeaseToken, p.clock.Now(), "")
		if err != nil {
			return err
		}
		ok, err := NewStore(tx).UpdateArchiveJob(ctx, completed, job.Revision)
		if err != nil {
			return err
		}
		if !ok {
			return ports.ErrArchiveJobConflict
		}
		return nil
	})
	if err != nil {
		return empty, err
	}
	return completed, nil
}

package gormstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) Store {
	return Store{db: db}
}

func Migrate(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&tenantModel{}, &inventoryModel{}, &inventoryAccessGrantModel{}, &assetModel{}, &authorizationOutboxEventModel{})
}

func (s Store) SaveTenant(ctx context.Context, item tenant.Tenant) error {
	model := tenantModel{
		ID:   item.ID.String(),
		Name: item.Name.String(),
	}

	return s.db.WithContext(ctx).Save(&model).Error
}

func (s Store) TenantExists(ctx context.Context, tenantID tenant.ID) (bool, error) {
	var model tenantModel
	err := s.db.WithContext(ctx).Where(&tenantModel{ID: tenantID.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s Store) SaveInventory(ctx context.Context, item inventory.Inventory) error {
	model := inventoryModel{
		ID:       item.ID.String(),
		TenantID: item.TenantID.String(),
		Name:     item.Name.String(),
	}

	return s.db.WithContext(ctx).Save(&model).Error
}

func (s Store) InventoryByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error) {
	var model inventoryModel
	err := s.db.WithContext(ctx).Where(&inventoryModel{
		ID:       inventoryID.String(),
		TenantID: tenantID.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return inventory.Inventory{}, false, nil
	}
	if err != nil {
		return inventory.Inventory{}, false, err
	}
	item, ok := model.toDomain()
	return item, ok, nil
}

func (s Store) SaveTenantAndEnqueueOwnerGrant(ctx context.Context, eventID string, item tenant.Tenant, principal identity.Principal) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&tenantModel{
			ID:   item.ID.String(),
			Name: item.Name.String(),
		}).Error; err != nil {
			return err
		}

		return tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(ports.AuthorizationOutboxGrantTenantOwner),
			PrincipalID: principal.ID.String(),
			TenantID:    item.ID.String(),
		}).Error
	})
}

func (s Store) SaveInventoryAndEnqueueOwnerGrant(ctx context.Context, eventID string, item inventory.Inventory, tenantID tenant.ID, principal identity.Principal) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&inventoryModel{
			ID:       item.ID.String(),
			TenantID: item.TenantID.String(),
			Name:     item.Name.String(),
		}).Error; err != nil {
			return err
		}

		inventoryID := item.ID.String()
		return tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(ports.AuthorizationOutboxGrantInventoryOwner),
			PrincipalID: principal.ID.String(),
			TenantID:    tenantID.String(),
			InventoryID: &inventoryID,
		}).Error
	})
}

func (s Store) SaveInventoryAccessGrantAndEnqueue(ctx context.Context, eventID string, grant ports.InventoryAccessGrant) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var containingInventory inventoryModel
		err := tx.Where(&inventoryModel{
			ID:       grant.InventoryID.String(),
			TenantID: grant.TenantID.String(),
		}).First(&containingInventory).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}

		grantKey := grant.CursorKey()
		var existingGrant inventoryAccessGrantModel
		err = tx.Where(&inventoryAccessGrantModel{
			TenantID:    grant.TenantID.String(),
			InventoryID: grant.InventoryID.String(),
			GrantKey:    grantKey,
		}).First(&existingGrant).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := tx.Save(&inventoryAccessGrantModel{
			TenantID:     grant.TenantID.String(),
			InventoryID:  grant.InventoryID.String(),
			GrantKey:     grantKey,
			PrincipalID:  grant.PrincipalID.String(),
			Relationship: string(grant.Relationship),
		}).Error; err != nil {
			return err
		}

		inventoryID := grant.InventoryID.String()
		return tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(outboxKindForInventoryAccess(grant.Relationship)),
			PrincipalID: grant.PrincipalID.String(),
			TenantID:    grant.TenantID.String(),
			InventoryID: &inventoryID,
		}).Error
	})
}

func (s Store) ClaimPendingAuthorizationOutboxEvents(ctx context.Context, claimID string, limit int, leaseUntil time.Time) ([]ports.AuthorizationOutboxEvent, error) {
	if limit <= 0 {
		limit = 25
	}

	events := []ports.AuthorizationOutboxEvent{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var models []authorizationOutboxEventModel
		now := time.Now()
		if err := tx.
			Clauses(skipLockedForUpdate()).
			Where(claimableAuthorizationOutboxEvent(now)).
			Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).
			Limit(limit).
			Find(&models).Error; err != nil {
			return err
		}

		claimed := make([]authorizationOutboxEventModel, 0, len(models))
		for _, model := range models {
			model.ClaimID = claimID
			model.ClaimedUntil = &leaseUntil
			claimed = append(claimed, model)
		}

		for _, model := range claimed {
			if err := tx.
				Model(&authorizationOutboxEventModel{}).
				Where(&authorizationOutboxEventModel{ID: model.ID}).
				Updates(map[string]any{
					"claim_id":      model.ClaimID,
					"claimed_until": model.ClaimedUntil,
				}).Error; err != nil {
				return err
			}
			events = append(events, model.toPort())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}

func skipLockedForUpdate() clause.Locking {
	return clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}
}

func claimableAuthorizationOutboxEvent(now time.Time) clause.Expression {
	return clause.And(
		clause.Eq{Column: clause.Column{Name: "processed_at"}, Value: nil},
		clause.Eq{Column: clause.Column{Name: "dead_lettered_at"}, Value: nil},
		clause.Or(
			clause.Eq{Column: clause.Column{Name: "claim_id"}, Value: ""},
			clause.Lte{Column: clause.Column{Name: "claimed_until"}, Value: now},
		),
	)
}

func (s Store) MarkAuthorizationOutboxEventProcessed(ctx context.Context, eventID string, claimID string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&authorizationOutboxEventModel{}).
		Where(&authorizationOutboxEventModel{ID: eventID, ClaimID: claimID}).
		Updates(map[string]any{
			"processed_at":  now,
			"last_error":    "",
			"claim_id":      "",
			"claimed_until": nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	return nil
}

func (s Store) MarkAuthorizationOutboxEventFailed(ctx context.Context, eventID string, claimID string, reason string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model authorizationOutboxEventModel
		if err := tx.Where(&authorizationOutboxEventModel{ID: eventID, ClaimID: claimID}).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrAuthorizationOutboxClaimLost
			}
			return err
		}

		model.Attempts++
		model.LastError = reason
		model.ClaimID = ""
		model.ClaimedUntil = nil
		return tx.Save(&model).Error
	})
}

func (s Store) MarkAuthorizationOutboxEventDeadLettered(ctx context.Context, eventID string, claimID string, reason string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&authorizationOutboxEventModel{}).
		Where(&authorizationOutboxEventModel{ID: eventID, ClaimID: claimID}).
		Updates(map[string]any{
			"dead_lettered_at":   now,
			"dead_letter_reason": reason,
			"claim_id":           "",
			"claimed_until":      nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrAuthorizationOutboxClaimLost
	}
	return nil
}

func (s Store) ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID, page ports.InventoryListPageRequest) ([]inventory.Inventory, error) {
	var models []inventoryModel
	query := s.db.WithContext(ctx).Where(&inventoryModel{TenantID: tenantID.String()})
	if page.AfterInventoryID.String() != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterInventoryID.String()})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]inventory.Inventory, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid inventory row %q", model.ID)
		}
		items = append(items, item)
	}

	return items, nil
}

func (s Store) ListInventoryAccessGrants(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.InventoryAccessGrantPageRequest) ([]ports.InventoryAccessGrant, error) {
	var models []inventoryAccessGrantModel
	query := s.db.WithContext(ctx).Where(&inventoryAccessGrantModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	})
	if page.AfterGrantKey != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "grant_key"}, Value: page.AfterGrantKey})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "grant_key"}}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]ports.InventoryAccessGrant, 0, len(models))
	for _, model := range models {
		item, ok := model.toPort()
		if !ok {
			return nil, fmt.Errorf("invalid inventory access grant row %q", model.GrantKey)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s Store) CreateAsset(ctx context.Context, item asset.Asset) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var containingInventory inventoryModel
		err := tx.Where(&inventoryModel{
			ID:       item.InventoryID.String(),
			TenantID: item.TenantID.String(),
		}).First(&containingInventory).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}

		if item.ParentAssetID.String() != "" {
			var parent assetModel
			err = tx.Where(&assetModel{
				ID:          item.ParentAssetID.String(),
				TenantID:    item.TenantID.String(),
				InventoryID: item.InventoryID.String(),
			}).First(&parent).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrForbidden
			}
			if err != nil {
				return err
			}
			parentKind, ok := asset.NewKind(parent.Kind)
			if !ok || !parentKind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive.String() || parent.ID == item.ID.String() {
				return ports.ErrForbidden
			}
			if err := rejectAssetContainmentCycle(tx, item.ID, parent); err != nil {
				return err
			}
		}

		parentAssetID := stringPtrFromAssetID(item.ParentAssetID)
		customFields, err := json.Marshal(item.CustomFields.Values())
		if err != nil {
			return err
		}
		return tx.Create(&assetModel{
			ID:             item.ID.String(),
			TenantID:       item.TenantID.String(),
			InventoryID:    item.InventoryID.String(),
			ParentAssetID:  parentAssetID,
			Kind:           item.Kind.String(),
			Title:          item.Title.String(),
			Description:    item.Description.String(),
			CustomFields:   string(customFields),
			LifecycleState: item.LifecycleState.String(),
		}).Error
	})
}

func (s Store) AssetByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error) {
	var model assetModel
	err := s.db.WithContext(ctx).Where(&assetModel{
		ID:          assetID.String(),
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return asset.Asset{}, false, nil
	}
	if err != nil {
		return asset.Asset{}, false, err
	}
	item, ok := model.toDomain()
	if !ok {
		return asset.Asset{}, false, fmt.Errorf("invalid asset row %q", model.ID)
	}
	return item, true, nil
}

func (s Store) ListAssetsByInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AssetListPageRequest) ([]asset.Asset, error) {
	var models []assetModel
	query := s.db.WithContext(ctx).Where(&assetModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	})
	if page.AfterAssetID.String() != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterAssetID.String()})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]asset.Asset, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid asset row %q", model.ID)
		}
		items = append(items, item)
	}
	return items, nil
}

func rejectAssetContainmentCycle(tx *gorm.DB, assetID asset.ID, parent assetModel) error {
	for current := parent; ; {
		if current.ID == assetID.String() {
			return ports.ErrForbidden
		}
		if current.ParentAssetID == nil {
			return nil
		}

		nextID := *current.ParentAssetID
		err := tx.Where(&assetModel{
			ID:          nextID,
			TenantID:    current.TenantID,
			InventoryID: current.InventoryID,
		}).First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
	}
}

type tenantModel struct {
	ID        string `gorm:"primaryKey;size:26"`
	Name      string `gorm:"not null;size:120"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (tenantModel) TableName() string {
	return "tenants"
}

type inventoryModel struct {
	ID        string      `gorm:"primaryKey;size:26"`
	TenantID  string      `gorm:"not null;size:26;index"`
	Tenant    tenantModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	Name      string      `gorm:"not null;size:120"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type inventoryAccessGrantModel struct {
	TenantID     string         `gorm:"primaryKey;size:26;index:idx_inventory_access_grants_inventory"`
	Tenant       tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID  string         `gorm:"primaryKey;size:26;index:idx_inventory_access_grants_inventory"`
	Inventory    inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:InventoryID;references:ID"`
	GrantKey     string         `gorm:"primaryKey;size:180"`
	PrincipalID  string         `gorm:"not null;size:128;index"`
	Relationship string         `gorm:"not null;size:32;check:chk_inventory_access_grants_relationship,relationship IN ('viewer','editor')"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type assetModel struct {
	ID             string         `gorm:"primaryKey;size:26"`
	TenantID       string         `gorm:"not null;size:26;index:idx_assets_tenant_inventory"`
	Tenant         tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID    string         `gorm:"not null;size:26;index:idx_assets_tenant_inventory;index:idx_assets_inventory_parent;index:idx_assets_inventory_kind"`
	Inventory      inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	ParentAssetID  *string        `gorm:"size:26;index;index:idx_assets_inventory_parent"`
	ParentAsset    *assetModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:ParentAssetID;references:ID"`
	Kind           string         `gorm:"not null;size:32;index:idx_assets_inventory_kind;check:chk_assets_kind,kind IN ('item','container','location')"`
	Title          string         `gorm:"not null;size:160"`
	Description    string         `gorm:"not null;default:''"`
	CustomFields   string         `gorm:"type:jsonb;not null;default:'{}'"`
	LifecycleState string         `gorm:"not null;size:32;check:chk_assets_lifecycle_state,lifecycle_state IN ('active','archived')"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type authorizationOutboxEventModel struct {
	ID               string         `gorm:"primaryKey;size:26"`
	Kind             string         `gorm:"not null;size:80;index;check:chk_authorization_outbox_events_kind,kind IN ('grant_tenant_owner','grant_inventory_owner','grant_inventory_viewer','grant_inventory_editor');check:chk_authorization_outbox_events_inventory_required,(kind IN ('grant_inventory_owner','grant_inventory_viewer','grant_inventory_editor') AND inventory_id IS NOT NULL) OR (kind = 'grant_tenant_owner' AND inventory_id IS NULL)"`
	PrincipalID      string         `gorm:"not null;size:128;index"`
	TenantID         string         `gorm:"not null;size:26;index"`
	Tenant           tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID      *string        `gorm:"size:26;index"`
	Inventory        inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	Attempts         int            `gorm:"not null;default:0"`
	LastError        string         `gorm:"not null;default:''"`
	ClaimID          string         `gorm:"not null;default:'';size:26;index"`
	ClaimedUntil     *time.Time     `gorm:"index"`
	ProcessedAt      *time.Time
	DeadLetteredAt   *time.Time `gorm:"index"`
	DeadLetterReason string     `gorm:"not null;default:''"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (authorizationOutboxEventModel) TableName() string {
	return "authorization_outbox_events"
}

func (inventoryModel) TableName() string {
	return "inventories"
}

func (inventoryAccessGrantModel) TableName() string {
	return "inventory_access_grants"
}

func (m *inventoryAccessGrantModel) BeforeSave(*gorm.DB) error {
	m.GrantKey = ports.InventoryAccessGrant{
		PrincipalID:  identity.PrincipalID(m.PrincipalID),
		Relationship: ports.InventoryAccessRelationship(m.Relationship),
	}.CursorKey()
	return nil
}

func (assetModel) TableName() string {
	return "assets"
}

func (m inventoryModel) toDomain() (inventory.Inventory, bool) {
	id, ok := inventory.NewID(m.ID)
	if !ok {
		return inventory.Inventory{}, false
	}
	name, ok := inventory.NewName(m.Name)
	if !ok {
		return inventory.Inventory{}, false
	}

	return inventory.Inventory{
		ID:       id,
		TenantID: inventory.TenantID(m.TenantID),
		Name:     name,
	}, true
}

func (m authorizationOutboxEventModel) toPort() ports.AuthorizationOutboxEvent {
	inventoryID := inventory.InventoryID("")
	if m.InventoryID != nil {
		inventoryID = inventory.InventoryID(*m.InventoryID)
	}
	event := ports.AuthorizationOutboxEvent{
		ID:          m.ID,
		Kind:        ports.AuthorizationOutboxEventKind(m.Kind),
		PrincipalID: identity.PrincipalID(m.PrincipalID),
		TenantID:    tenant.ID(m.TenantID),
		InventoryID: inventoryID,
		Attempts:    m.Attempts,
		LastError:   m.LastError,
		ClaimID:     m.ClaimID,
		CreatedAt:   m.CreatedAt,
	}
	if m.ClaimedUntil != nil {
		event.ClaimedUntil = *m.ClaimedUntil
	}
	if m.DeadLetteredAt != nil {
		event.DeadLetteredAt = *m.DeadLetteredAt
	}
	event.DeadLetterReason = m.DeadLetterReason
	return event
}

func (m inventoryAccessGrantModel) toPort() (ports.InventoryAccessGrant, bool) {
	principalID, ok := identity.NewPrincipalID(m.PrincipalID)
	if !ok {
		return ports.InventoryAccessGrant{}, false
	}
	relationship := ports.InventoryAccessRelationship(m.Relationship)
	switch relationship {
	case ports.InventoryAccessViewer, ports.InventoryAccessEditor:
	default:
		return ports.InventoryAccessGrant{}, false
	}
	return ports.InventoryAccessGrant{
		TenantID:     tenant.ID(m.TenantID),
		InventoryID:  inventory.InventoryID(m.InventoryID),
		PrincipalID:  principalID,
		Relationship: relationship,
	}, true
}

func (m assetModel) toDomain() (asset.Asset, bool) {
	id, ok := asset.NewID(m.ID)
	if !ok {
		return asset.Asset{}, false
	}
	kind, ok := asset.NewKind(m.Kind)
	if !ok {
		return asset.Asset{}, false
	}
	title, ok := asset.NewTitle(m.Title)
	if !ok {
		return asset.Asset{}, false
	}
	var customFieldValues map[string]any
	if err := json.Unmarshal([]byte(m.CustomFields), &customFieldValues); err != nil {
		return asset.Asset{}, false
	}
	customFields, ok := asset.NewCustomFields(customFieldValues)
	if !ok {
		return asset.Asset{}, false
	}
	lifecycleState := asset.LifecycleState(m.LifecycleState)
	switch lifecycleState {
	case asset.LifecycleStateActive, asset.LifecycleStateArchived:
	default:
		return asset.Asset{}, false
	}
	parentID := asset.ID("")
	if m.ParentAssetID != nil {
		parentID, ok = asset.NewID(*m.ParentAssetID)
		if !ok {
			return asset.Asset{}, false
		}
	}
	return asset.Asset{
		ID:             id,
		TenantID:       asset.TenantID(m.TenantID),
		InventoryID:    asset.InventoryID(m.InventoryID),
		ParentAssetID:  parentID,
		Kind:           kind,
		Title:          title,
		Description:    asset.NewDescription(m.Description),
		CustomFields:   customFields,
		LifecycleState: lifecycleState,
	}, true
}

func stringPtrFromAssetID(id asset.ID) *string {
	if id.String() == "" {
		return nil
	}
	value := id.String()
	return &value
}

func outboxKindForInventoryAccess(relationship ports.InventoryAccessRelationship) ports.AuthorizationOutboxEventKind {
	switch relationship {
	case ports.InventoryAccessEditor:
		return ports.AuthorizationOutboxGrantInventoryEditor
	default:
		return ports.AuthorizationOutboxGrantInventoryViewer
	}
}

package gormstore

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) Store {
	return Store{db: db}
}

func Migrate(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&tenantModel{}, &inventoryModel{})
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

func (s Store) ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID) ([]inventory.Inventory, error) {
	var models []inventoryModel
	if err := s.db.WithContext(ctx).Where(&inventoryModel{TenantID: tenantID.String()}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]inventory.Inventory, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			continue
		}
		items = append(items, item)
	}

	sort.Slice(items, func(left int, right int) bool {
		return items[left].ID.String() < items[right].ID.String()
	})

	return items, nil
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

func (inventoryModel) TableName() string {
	return "inventories"
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

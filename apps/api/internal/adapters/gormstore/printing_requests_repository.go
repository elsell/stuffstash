package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

func (s Store) FindPrintJobRequest(ctx context.Context, scope printing.Scope, actor, key string) (printing.Job, string, error) {
	if scope.TenantID == "" || scope.InventoryID == "" || actor == "" || key == "" {
		return printing.Job{}, "", ports.ErrPrintJobNotFound
	}
	m, err := printRequest(s.db.WithContext(ctx), printing.Job{Scope: scope, RequestedBy: actor, IdempotencyKey: key})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ports.ErrPrintJobNotFound
	}
	if err != nil {
		return printing.Job{}, "", err
	}
	j, err := m.domain()
	return j, m.RequestFingerprint, err
}

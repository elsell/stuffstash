package gormstore

import (
	"context"
	"fmt"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"gorm.io/gorm/clause"
)

func (s Store) SaveUser(ctx context.Context, user identity.User) error {
	model := userModel{
		ID:    user.ID.String(),
		Email: user.Email.String(),
	}
	onConflict := clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: user.Email.String() == "",
	}
	if user.Email.String() != "" {
		onConflict.DoUpdates = clause.AssignmentColumns([]string{"email", "updated_at"})
	}
	return s.db.WithContext(ctx).Clauses(onConflict).Create(&model).Error
}

func (s Store) UsersByID(ctx context.Context, ids []identity.PrincipalID) (map[identity.PrincipalID]identity.User, error) {
	if len(ids) == 0 {
		return map[identity.PrincipalID]identity.User{}, nil
	}
	values := make([]any, 0, len(ids))
	seen := map[string]struct{}{}
	for _, id := range ids {
		if id.String() == "" {
			continue
		}
		if _, ok := seen[id.String()]; ok {
			continue
		}
		seen[id.String()] = struct{}{}
		values = append(values, id.String())
	}
	var models []userModel
	if err := s.db.WithContext(ctx).Where(clause.IN{Column: clause.Column{Name: "id"}, Values: values}).Find(&models).Error; err != nil {
		return nil, err
	}
	users := make(map[identity.PrincipalID]identity.User, len(models))
	for _, model := range models {
		user, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid user row %q", model.ID)
		}
		users[user.ID] = user
	}
	return users, nil
}

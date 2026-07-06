package gormstore

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

type userModel struct {
	ID        string `gorm:"primaryKey;size:128"`
	Email     string `gorm:"not null;default:'';size:320"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (userModel) TableName() string {
	return "users"
}

func (m userModel) toDomain() (identity.User, bool) {
	id, ok := identity.NewPrincipalID(m.ID)
	if !ok {
		return identity.User{}, false
	}
	email := identity.Email("")
	if m.Email != "" {
		parsed, ok := identity.NewEmail(m.Email)
		if !ok {
			return identity.User{}, false
		}
		email = parsed
	}
	return identity.NewUser(id, email)
}

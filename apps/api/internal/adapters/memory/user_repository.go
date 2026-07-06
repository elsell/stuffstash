package memory

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

func (s *Store) SaveUser(_ context.Context, user identity.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.users[user.ID]; ok && user.Email.String() == "" {
		user.Email = existing.Email
	}
	s.users[user.ID] = user
	return nil
}

func (s *Store) UsersByID(_ context.Context, ids []identity.PrincipalID) (map[identity.PrincipalID]identity.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := map[identity.PrincipalID]identity.User{}
	for _, id := range ids {
		if user, ok := s.users[id]; ok {
			users[id] = user
		}
	}
	return users, nil
}

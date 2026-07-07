package app

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) ResolveUsersByID(ctx context.Context, ids []identity.PrincipalID) map[identity.PrincipalID]identity.User {
	if a.users == nil || len(ids) == 0 {
		return map[identity.PrincipalID]identity.User{}
	}
	deduped := make([]identity.PrincipalID, 0, len(ids))
	seen := map[identity.PrincipalID]struct{}{}
	for _, id := range ids {
		if id.String() == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		deduped = append(deduped, id)
	}
	if len(deduped) == 0 {
		return map[identity.PrincipalID]identity.User{}
	}
	users, err := a.users.UsersByID(ctx, deduped)
	if err != nil {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventPrincipalResolutionFailed,
			Message: "principal resolution failed",
			Fields: map[string]string{
				"count": strconv.Itoa(len(deduped)),
				"error": err.Error(),
			},
		})
		return map[identity.PrincipalID]identity.User{}
	}
	return users
}

package audithistory

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) resolveAuditPrincipals(ctx context.Context, input ListAuditRecordsInput, items []audit.Record) map[identity.PrincipalID]identity.User {
	if a.deps.Users == nil || len(items) == 0 {
		return map[identity.PrincipalID]identity.User{}
	}
	ids := make([]identity.PrincipalID, 0, len(items))
	seen := map[identity.PrincipalID]struct{}{}
	for _, item := range items {
		id := identity.PrincipalID(item.PrincipalID.String())
		if id.String() == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	users, err := a.deps.Users.UsersByID(ctx, ids)
	if err != nil {
		a.deps.Observer.Record(ctx, ports.Event{
			Name:    ports.EventAuditPrincipalResolutionFailed,
			Message: "audit principal resolution failed",
			Fields: map[string]string{
				"tenant_id":    input.TenantID.String(),
				"inventory_id": input.InventoryID.String(),
				"principal_id": input.Principal.ID.String(),
				"count":        strconv.Itoa(len(ids)),
				"error":        err.Error(),
			},
		})
		return map[identity.PrincipalID]identity.User{}
	}
	return users
}

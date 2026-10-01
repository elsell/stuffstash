package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

// RealtimeInventoryAccess keeps session authorization and denial observability
// behind the inventory application's access boundary.
type RealtimeInventoryAccess interface {
	ActiveInventoryAccess
	RecordAuthorizationDenied(context.Context, identity.Principal, tenant.ID)
}

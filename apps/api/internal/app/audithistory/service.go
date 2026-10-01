package audithistory

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Dependencies struct {
	Access           ports.InventoryReadAccess
	Assets           ports.AssetRepository
	Audit            ports.AuditRepository
	Authorizer       ports.Authorizer
	Clock            ports.Clock
	DefaultPageLimit int
	IDs              ports.IDGenerator
	MaxPageLimit     int
	Observer         ports.Observer
	Undoables        ports.UndoableOperationRepository
	Users            ports.UserRepository
}
type Service struct{ deps Dependencies }

func New(deps Dependencies) Service { return Service{deps: deps} }
func (a Service) recordAuthorizationDenied(ctx context.Context, principal identity.Principal, tenantID tenant.ID) {
	a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventAuthorizationDenied, Message: "authorization denied", Fields: map[string]string{"tenant_id": tenantID.String(), "principal_id": principal.ID.String()}})
}

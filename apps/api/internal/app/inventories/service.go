package inventories

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Service struct {
	invitationAllowInsecureHTTP bool
	invitationPublicBaseURL     string
	invitationTTL               time.Duration
	inventoryAccessUnitOfWork   ports.InventoryAccessUnitOfWork
	inventoryAccess             ports.InventoryAccessRepository
	observer                    ports.Observer
	authorizer                  ports.Authorizer
	tenants                     ports.TenantRepository
	tenantUnitOfWork            ports.TenantUnitOfWork
	inventories                 ports.InventoryRepository
	inventoryUnitOfWork         ports.InventoryUnitOfWork
	audit                       ports.AuditRepository
	outbox                      ports.AuthorizationOutbox
	ids                         ports.IDGenerator
	clock                       ports.Clock
	outboxDrainLimit            int
	outboxClaimLease            time.Duration
	defaultPageLimit            int
	maxPageLimit                int
}

type Dependencies struct {
	InvitationAllowInsecureHTTP bool
	InvitationPublicBaseURL     string
	InvitationTTL               time.Duration
	InventoryAccessUnitOfWork   ports.InventoryAccessUnitOfWork
	InventoryAccess             ports.InventoryAccessRepository
	Observer                    ports.Observer
	Authorizer                  ports.Authorizer
	Tenants                     ports.TenantRepository
	TenantUnitOfWork            ports.TenantUnitOfWork
	Inventories                 ports.InventoryRepository
	InventoryUnitOfWork         ports.InventoryUnitOfWork
	Audit                       ports.AuditRepository
	Outbox                      ports.AuthorizationOutbox
	IDs                         ports.IDGenerator
	Clock                       ports.Clock
	OutboxDrainLimit            int
	OutboxClaimLease            time.Duration
	DefaultPageLimit            int
	MaxPageLimit                int
}

func New(deps Dependencies) Service {
	return Service{
		invitationAllowInsecureHTTP: deps.InvitationAllowInsecureHTTP,
		invitationPublicBaseURL:     deps.InvitationPublicBaseURL,
		invitationTTL:               deps.InvitationTTL,
		inventoryAccessUnitOfWork:   deps.InventoryAccessUnitOfWork,
		inventoryAccess:             deps.InventoryAccess,
		observer:                    deps.Observer,
		authorizer:                  deps.Authorizer,
		tenants:                     deps.Tenants,
		tenantUnitOfWork:            deps.TenantUnitOfWork,
		inventories:                 deps.Inventories,
		inventoryUnitOfWork:         deps.InventoryUnitOfWork,
		audit:                       deps.Audit,
		outbox:                      deps.Outbox,
		ids:                         deps.IDs,
		clock:                       deps.Clock,
		outboxDrainLimit:            deps.OutboxDrainLimit,
		outboxClaimLease:            deps.OutboxClaimLease,
		defaultPageLimit:            deps.DefaultPageLimit,
		maxPageLimit:                deps.MaxPageLimit,
	}
}

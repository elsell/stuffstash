package search

import "github.com/stuffstash/stuff-stash/internal/ports"

type Dependencies struct {
	Tenants          ports.TenantRepository
	Authorizer       ports.Authorizer
	Inventories      ports.InventoryRepository
	Search           ports.AssetSearchRepository
	Assets           ports.AssetRepository
	Attachments      ports.AttachmentRepository
	Audit            ports.AuditRepository
	IDs              ports.IDGenerator
	Clock            ports.Clock
	Observer         ports.Observer
	DefaultPageLimit int
	MaxPageLimit     int
}
type Service struct{ deps Dependencies }

func New(deps Dependencies) Service { return Service{deps: deps} }

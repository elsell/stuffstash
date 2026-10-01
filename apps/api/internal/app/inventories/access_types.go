package inventories

import (
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type AccessRelationship string

const (
	AccessRelationshipOwner  AccessRelationship = "owner"
	AccessRelationshipEditor AccessRelationship = "editor"
	AccessRelationshipViewer AccessRelationship = "viewer"
)

type AccessSummary struct {
	Relationship AccessRelationship
	Permissions  []string
}

type MyTenantAccess struct {
	Tenant tenant.Tenant
	Access AccessSummary
}

type ListMyTenantsInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	Limit     int
	Cursor    string
}

type ListMyTenantsResult struct {
	Items      []MyTenantAccess
	Limit      int
	NextCursor *string
	HasMore    bool
}

package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type ExpirationWorkspaceInput struct {
	Kind          string   `query:"kind" enum:"item,container,location"`
	CheckoutState string   `query:"checkoutState" enum:"any,available,checked_out"`
	Authorization string   `header:"Authorization"`
	RequestID     string   `header:"X-Request-ID"`
	TenantID      string   `path:"tenantId"`
	InventoryID   string   `path:"inventoryId"`
	Mode          string   `query:"mode" enum:"soon,expired,all" default:"all"`
	Query         string   `query:"q" maxLength:"120"`
	TypeID        string   `query:"customAssetTypeId" maxLength:"128"`
	TagIDs        []string `query:"tagIds"`
	LocationID    string   `query:"locationId" maxLength:"128"`
	FromDate      string   `query:"fromDate"`
	ThroughDate   string   `query:"throughDate"`
	Limit         int      `query:"limit" minimum:"1" maximum:"100" default:"50"`
	Cursor        string   `query:"cursor"`
}
type ExpirationWorkspaceCounts struct {
	Soon    int `json:"soon"`
	Expired int `json:"expired"`
	All     int `json:"all"`
}
type ExpirationWorkspaceAncestor struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
type ExpirationWorkspaceAsset struct {
	AssetResponse
	AncestorPath []ExpirationWorkspaceAncestor `json:"ancestorPath"`
}
type ExpirationWorkspaceData struct {
	Items    []ExpirationWorkspaceAsset `json:"items"`
	Counts   ExpirationWorkspaceCounts  `json:"counts"`
	Timezone string                     `json:"timezone"`
}
type ExpirationWorkspaceOutput struct {
	Body shared.SuccessEnvelope[ExpirationWorkspaceData]
}

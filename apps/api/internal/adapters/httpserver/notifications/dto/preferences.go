package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type ScopeInput struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
	TenantID      string `path:"tenantId"`
	InventoryID   string `path:"inventoryId"`
}
type ExpirationPolicy struct {
	Enabled     bool `json:"enabled"`
	Upcoming    bool `json:"upcoming"`
	Expired     bool `json:"expired"`
	AdvanceDays int  `json:"advanceDays" minimum:"0" maximum:"3650"`
}
type InitializeInput struct {
	ScopeInput
	Body InitializeBody
}
type InitializeBody struct {
	Timezone string `json:"timezone" minLength:"1" maxLength:"100"`
}
type UpdateInput struct {
	ScopeInput
	Body UpdateBody
}
type UpdateBody struct {
	Revision    int64            `json:"revision" minimum:"1"`
	Defaults    ExpirationPolicy `json:"defaults"`
	Timezone    string           `json:"timezone" minLength:"1" maxLength:"100"`
	PushEnabled bool             `json:"pushEnabled"`
}
type TypeOverrideInput struct {
	ScopeInput
	CustomAssetTypeID string `path:"customAssetTypeId"`
	Body              TypeOverrideBody
}
type TypeOverrideBody struct {
	Revision int64            `json:"revision" minimum:"1"`
	Settings ExpirationPolicy `json:"settings"`
}
type DeleteTypeOverrideInput struct {
	ScopeInput
	CustomAssetTypeID string `path:"customAssetTypeId"`
	Revision          int64  `query:"revision" minimum:"1" required:"true"`
}
type TypeOverrideResponse struct {
	CustomAssetTypeID string           `json:"customAssetTypeId"`
	Settings          ExpirationPolicy `json:"settings"`
}
type PreferencesResponse struct {
	Revision    int64                  `json:"revision"`
	Defaults    ExpirationPolicy       `json:"defaults"`
	Timezone    string                 `json:"timezone"`
	PushEnabled bool                   `json:"pushEnabled"`
	Overrides   []TypeOverrideResponse `json:"overrides"`
}
type PreferencesOutput struct {
	Body shared.SuccessEnvelope[PreferencesResponse]
}

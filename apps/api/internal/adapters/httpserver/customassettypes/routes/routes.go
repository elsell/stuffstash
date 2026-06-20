package routes

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func Register(api huma.API, application app.App) {
	RegisterCreateTenant(api, application)
	RegisterCreateInventory(api, application)
	RegisterDetailTenant(api, application)
	RegisterDetailInventory(api, application)
	RegisterUpdateTenant(api, application)
	RegisterUpdateInventory(api, application)
	RegisterArchiveTenant(api, application)
	RegisterArchiveInventory(api, application)
	RegisterRestoreTenant(api, application)
	RegisterRestoreInventory(api, application)
	RegisterDeleteTenant(api, application)
	RegisterDeleteInventory(api, application)
	RegisterListTenant(api, application)
	RegisterListInventory(api, application)
}

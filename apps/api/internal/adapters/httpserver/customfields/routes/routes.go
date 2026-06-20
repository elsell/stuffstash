package routes

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func Register(api huma.API, application app.App) {
	RegisterCreateTenant(api, application)
	RegisterDetailTenant(api, application)
	RegisterUpdateTenant(api, application)
	RegisterArchiveTenant(api, application)
	RegisterRestoreTenant(api, application)
	RegisterDeleteTenant(api, application)
	RegisterListTenant(api, application)
	RegisterCreateInventory(api, application)
	RegisterDetailInventory(api, application)
	RegisterUpdateInventory(api, application)
	RegisterArchiveInventory(api, application)
	RegisterRestoreInventory(api, application)
	RegisterDeleteInventory(api, application)
	RegisterListInventory(api, application)
}

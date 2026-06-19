package routes

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func Register(api huma.API, application app.App) {
	RegisterCreateTenant(api, application)
	RegisterListTenant(api, application)
	RegisterCreateInventory(api, application)
	RegisterListInventory(api, application)
}

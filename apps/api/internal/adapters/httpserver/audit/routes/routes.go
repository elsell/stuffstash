package routes

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func Register(api huma.API, application app.App) {
	RegisterListTenant(api, application)
	RegisterListInventory(api, application)
	RegisterListAsset(api, application)
	RegisterListAssetActivity(api, application)
}

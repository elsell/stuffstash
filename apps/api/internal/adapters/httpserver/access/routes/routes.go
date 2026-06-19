package routes

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func Register(api huma.API, application app.App) {
	RegisterGrant(api, application)
	RegisterList(api, application)
	RegisterRevoke(api, application)
}

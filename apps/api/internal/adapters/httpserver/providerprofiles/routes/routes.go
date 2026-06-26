package routes

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func Register(api huma.API, application app.App) {
	RegisterCreate(api, application)
	RegisterList(api, application)
	RegisterDetail(api, application)
	RegisterUpdate(api, application)
	RegisterLifecycle(api, application)
	RegisterCredential(api, application)
	RegisterTest(api, application)
}

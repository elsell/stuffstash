package routes

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func Register(api huma.API, application app.App) {
	RegisterCreate(api, application)
	RegisterDetail(api, application)
	RegisterList(api, application)
	RegisterDownload(api, application)
	RegisterArchive(api, application)
	RegisterRestore(api, application)
	RegisterDelete(api, application)
}

package httpserver

import (
	"github.com/danielgtaylor/huma/v2"
	accessroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/routes"
	assetroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/routes"
	attachmentroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/attachments/routes"
	auditroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/routes"
	customassettyperoutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/routes"
	customfieldroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customfields/routes"
	identityroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/routes"
	inventoryroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/routes"
	tenantroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tenants/routes"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func registerRoutes(api huma.API, application app.App) {
	identityroutes.Register(api, application)
	tenantroutes.Register(api, application)
	inventoryroutes.Register(api, application)
	customassettyperoutes.Register(api, application)
	customfieldroutes.Register(api, application)
	assetroutes.Register(api, application)
	attachmentroutes.Register(api, application)
	auditroutes.Register(api, application)
	accessroutes.Register(api, application)
}

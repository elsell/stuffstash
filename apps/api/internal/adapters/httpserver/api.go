package httpserver

import (
	"github.com/danielgtaylor/huma/v2"
	accessroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/routes"
	assetroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/routes"
	attachmentroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/attachments/routes"
	auditroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/routes"
	workflowroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/conversationworkflows/routes"
	customassettyperoutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/routes"
	customfieldroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customfields/routes"
	evaluationcaseroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationcases/routes"
	identityroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/routes"
	importroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/imports/routes"
	inventoryroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/routes"
	providerprofileroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/providerprofiles/routes"
	searchroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/search/routes"
	tagroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tags/routes"
	tenantroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tenants/routes"
	undoableoperationroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/undoableoperations/routes"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func registerRoutes(api huma.API, application app.App) {
	identityroutes.Register(api, application)
	tenantroutes.Register(api, application)
	inventoryroutes.Register(api, application)
	customassettyperoutes.Register(api, application)
	customfieldroutes.Register(api, application)
	assetroutes.Register(api, application)
	tagroutes.Register(api, application)
	attachmentroutes.Register(api, application)
	importroutes.Register(api, application)
	undoableoperationroutes.Register(api, application)
	auditroutes.Register(api, application)
	accessroutes.Register(api, application)
	searchroutes.Register(api, application)
	providerprofileroutes.Register(api, application)
	workflowroutes.Register(api, application)
	evaluationcaseroutes.Register(api, application)
}

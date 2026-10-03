package httpserver

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
	accessroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/routes"
	archiveroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/archives/routes"
	assetroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/routes"
	attachmentroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/attachments/routes"
	auditroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/routes"
	clienttelemetryroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/clienttelemetry/routes"
	workflowroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/conversationworkflows/routes"
	customassettyperoutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/routes"
	customfieldroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customfields/routes"
	evaluationcaseroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationcases/routes"
	evaluationrunroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationruns/routes"
	exportroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/exports/routes"
	identitydto "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/dto"
	identityroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/routes"
	importroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/imports/routes"
	inventoryroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/routes"
	notificationroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/notifications/routes"
	providerprofileroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/providerprofiles/routes"
	searchroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/search/routes"
	tagroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tags/routes"
	tenantroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tenants/routes"
	undoableoperationroutes "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/undoableoperations/routes"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
)

func registerRoutes(api huma.API, application app.App, archives *dataportability.ArchiveService, archiveTimeout time.Duration, cliAuth *identitydto.CLIAuthMetadata) {
	archiveroutes.Register(api, application, archives, archiveTimeout)
	exportroutes.Register(api, application)
	notificationroutes.Register(api, application)
	clienttelemetryroutes.Register(api, application)
	identityroutes.Register(api, application)
	identityroutes.RegisterCLIAuth(api, cliAuth)
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
	evaluationrunroutes.Register(api, application)
}

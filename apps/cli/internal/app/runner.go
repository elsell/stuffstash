package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Runner struct {
	InvitationWriter func(string, string) (ports.InvitationWriter, error)
	ProviderWrites   func(string, string) (ports.ProviderWrites, error)
	ArchiveAPI       func(string, string) (ports.ArchiveAPI, error)
	ImportSources    func(string, string) (ports.ImportSources, error)
	SearchAPI        func(string, string) (ports.AssetSearch, error)

	BinaryFiles    ports.BinaryFiles
	StreamFiles    ports.StreamFiles
	UploadFiles    ports.UploadFiles
	UploadTransfer ports.UploadTransfer
	BinaryAPI      func(string, string) (ports.BinaryAPI, error)

	AssetTypesAPI              func(string, string) (ports.AssetTypesAPI, error)
	FieldDefinitionsAPI        func(string, string) (ports.FieldDefinitionsAPI, error)
	PairingApprovalAPI         func(string, string) (ports.PairingApprovalAPI, error)
	PrinterAdministrationAPI   func(string, string) (ports.PrinterAdministrationAPI, error)
	TelemetryAPI               func(string, string) (ports.TelemetryAPI, error)
	VoiceProviderAPI           func(string, string) (ports.VoiceProviderAPI, error)
	EvaluationAPI              func(string, string) (ports.EvaluationAPI, error)
	WorkflowsAPI               func(string, string) (ports.WorkflowsAPI, error)
	ConnectorInspectionAPI     func(string, string) (ports.ConnectorInspectionAPI, error)
	PrintSettingsAPI           func(string, string) (ports.PrintSettingsAPI, error)
	ProviderProfilesAPI        func(string, string) (ports.ProviderProfilesAPI, error)
	ImportJobsAPI              func(string, string) (ports.ImportJobsAPI, error)
	ServerAPI                  func(string) (ports.ServerAPI, error)
	InvitationsAPI             func(string, string) (ports.InvitationsAPI, error)
	AccessGrantsAPI            func(string, string) (ports.AccessGrantsAPI, error)
	ActivityAPI                func(string, string) (ports.ActivityAPI, error)
	AuditAPI                   func(string, string) (ports.AuditAPI, error)
	OperationsAPI              func(string, string) (ports.OperationsAPI, error)
	SecretInput                ports.SecretInput
	NotificationPreferencesAPI func(string, string) (ports.NotificationPreferencesAPI, error)
	NotificationDevicesAPI     func(string, string) (ports.NotificationDevicesAPI, error)
	NotificationsAPI           func(string, string) (ports.NotificationsAPI, error)
	AttachmentUploads          func(string, string) (ports.AttachmentUploads, error)
	AttachmentsAPI             func(string, string) (ports.AttachmentsAPI, error)
	TagsAPI                    func(string, string) (ports.TagsAPI, error)
	DirectoryLifecycle         func(string, string) (ports.DirectoryLifecycle, error)
	TextInput                  ports.TextInput
	InputFiles                 ports.InputFiles
	DirectoryWriter            func(string, string) (ports.DirectoryWriter, error)
	DirectoryAPI               func(string, string) (ports.Directory, error)
	Picker                     ports.Selector
	ScopeAPI                   func(string, string) (ports.ScopeCatalog, error)
	Contexts                   contexts.Store
	LabelsAPI                  func(string, string) (ports.LabelsAPI, error)
	LabelFiles                 ports.LabelFiles
	PrintingAPI                func(string, string) (ports.HumanPrintingAPI, error)
	API                        func(string, string) (ports.API, error)
	Auth                       ports.Auth
	Credentials                ports.Credentials
	Output                     ports.Output
	Clock                      ports.Clock
	Observer                   ports.Observer
}

func (r Runner) Run(ctx context.Context, o Options) error {
	if o.InputPath != "" && !acceptsBody(o) {
		return ports.Failure("usage", "This command does not accept --input. Remove the option.")
	}
	if len(o.Command) == 0 {
		return ports.Failure("usage", "a command is required; use --help")
	}
	if o.Command[0] == "context" {
		return r.contextCommand(ctx, o.Command)
	}
	switch o.Command[0] {
	case "server":
		return r.serverCommand(ctx, o)
	case "login":
		if len(o.Command) != 1 {
			return ports.Failure("usage", "login takes no positional arguments")
		}
		api, err := r.API(o.Server, "")
		if err != nil {
			return err
		}
		metadata, err := api.AuthConfig(ctx)
		if err != nil {
			return err
		}
		session, err := r.Auth.Login(ctx, o.Server, metadata, o.DeviceCode)
		if err != nil {
			return err
		}
		if err := r.Credentials.Save(ctx, session); err != nil {
			return err
		}
		r.Observer.Event(ctx, "cli.login.completed")
		return r.Output.Result(map[string]string{"status": "signed in", "server": o.Server})
	case "logout":
		if len(o.Command) != 1 {
			return ports.Failure("usage", "logout takes no positional arguments")
		}
		if err := r.Credentials.Delete(ctx, o.Server); err != nil {
			return err
		}
		if r.Contexts != nil {
			if err := (contexts.Manager{Store: r.Contexts}).ClearServer(ctx, o.Server); err != nil {
				return err
			}
		}
		r.Observer.Event(ctx, "cli.logout.completed")
		return r.Output.Result(map[string]string{"status": "signed out"})
	}
	if err := validateCommandShape(o); err != nil {
		return err
	}
	o, err := r.chooseDefinitionLevel(ctx, o)
	if err != nil {
		return err
	}
	o, err = r.prepareInput(ctx, o)
	if err != nil {
		return err
	}
	session, err := r.Credentials.Load(ctx, o.Server)
	if err != nil {
		return err
	}
	if !session.ExpiresAt.After(r.Clock.Now().Add(30*time.Second)) || r.Contexts != nil && session.Subject == "" && missingResourceScope(o) {
		session, err = r.Auth.Refresh(ctx, session)
		if err != nil {
			return err
		}
		if err = r.Credentials.Save(ctx, session); err != nil {
			return err
		}
	}
	if isTelemetry(o) {
		return r.telemetryCommand(ctx, o, session.IDToken)
	}
	if isTenantCreate(o) {
		return r.writeDirectory(ctx, o, session.IDToken)
	}
	if isAccountCommand(o) {
		return r.directoryCommand(ctx, o, session.IDToken)
	}
	if r.Contexts != nil {
		config, configErr := r.Contexts.Load(ctx)
		if configErr != nil {
			return configErr
		}
		request := o.Selection
		request.Server = o.Server
		request.Tenant = o.Scope.Tenant
		request.Inventory = o.Scope.Inventory
		resolved, resolveErr := contexts.Resolve(config, request, contexts.Principal(session))
		if resolveErr != nil {
			return resolveErr
		}
		o.Scope = resolved.Scope
	}
	if isArchiveCommand(o) && o.Command[1] != "create" {
		o.Scope.Inventory = o.Selection.Inventory
	}
	o, err = r.chooseMissingScope(ctx, o, session)
	if err != nil {
		return err
	}
	if err := validateCommand(o); err != nil {
		return err
	}

	if isSearch(o) {
		return r.searchAssets(ctx, o, session.IDToken)
	}
	if isArchiveCommand(o) {
		return r.archiveCommand(ctx, o, session.IDToken)
	}
	if isImportSource(o) {
		return r.importSourceCommand(ctx, o, session.IDToken)
	}
	if isPairingApproval(o) {
		return r.pairingApprovalCommand(ctx, o, session.IDToken)
	}
	if isPrinterAdministration(o) {
		return r.printerAdministration(ctx, o, session.IDToken)
	}
	if isConnectorInspection(o) {
		return r.connectorInspection(ctx, o, session.IDToken)
	}
	if isPrintSettingsCommand(o) {
		return r.printSettingsCommand(ctx, o, session.IDToken)
	}
	if isEvaluationCommand(o) {
		return r.evaluationCommand(ctx, o, session.IDToken)
	}
	if isVoiceProviderCommand(o) {
		return r.voiceProviderCommand(ctx, o, session.IDToken)
	}
	if isCustomization(o) {
		return r.customizationCommand(ctx, o, session.IDToken)
	}
	if isWorkflowCommand(o) {
		return r.workflowCommand(ctx, o, session.IDToken)
	}
	if isProviderProfileCommand(o) {
		return r.providerProfileCommand(ctx, o, session.IDToken)
	}
	if isImportJobCommand(o) {
		return r.importJobCommand(ctx, o, session.IDToken)
	}
	if isInvitationCommand(o) {
		return r.invitationCommand(ctx, o, session.IDToken)
	}
	if isAccessGrantCommand(o) {
		return r.accessGrantCommand(ctx, o, session.IDToken)
	}
	if isActivityCommand(o) {
		return r.activityCommand(ctx, o, session.IDToken)
	}
	if isAuditCommand(o) {
		return r.auditCommand(ctx, o, session.IDToken)
	}
	if isOperationCommand(o) {
		return r.operationCommand(ctx, o, session.IDToken)
	}
	if isPreferenceCommand(o) {
		return r.preferencesCommand(ctx, o, session.IDToken)
	}
	if isNotificationDeviceCommand(o) {
		return r.notificationDevicesCommand(ctx, o, session.IDToken)
	}
	if isNotificationCommand(o) {
		return r.notificationsCommand(ctx, o, session.IDToken)
	}
	if isBinaryCommand(o) {
		return r.binaryCommand(ctx, o, session.IDToken)
	}
	if isAttachmentCommand(o) {
		return r.attachmentsCommand(ctx, o, session.IDToken)
	}
	if isTagCommand(o) {
		return r.tagsCommand(ctx, o, session.IDToken)
	}
	if isDirectoryLifecycle(o) {
		return r.directoryLifecycle(ctx, o, session)
	}
	if isDirectoryWrite(o) {
		return r.writeDirectory(ctx, o, session.IDToken)
	}
	if isDirectoryCommand(o) {
		return r.directoryCommand(ctx, o, session.IDToken)
	}
	if isAssetWrite(o) || isCheckoutWrite(o) {
		if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory)); err != nil {
			return err
		}
	}
	if isAssetWrite(o) && o.Command[1] == "create" && assetPrintRequested(o) && !o.PrintLabel {
		if o.IdempotencyKey == "" {
			var value [16]byte
			if _, err := rand.Read(value[:]); err != nil {
				return err
			}
			o.IdempotencyKey = hex.EncodeToString(value[:])
		}
		if err := r.Output.Notice("Print request key: " + strconv.Quote(o.IdempotencyKey) + "; keep this key and the unchanged request for a retry."); err != nil {
			return err
		}
	}
	if err := r.confirmPrintCancellation(ctx, o); err != nil {
		return err
	}
	if err := r.confirmAssetLifecycle(ctx, o); err != nil {
		return err
	}
	api, err := r.API(o.Server, session.IDToken)
	if err != nil {
		return err
	}
	var result any
	if isLabelCommand(o) {
		if r.LabelsAPI == nil {
			return ports.Failure("configuration", "label API is unavailable")
		}
		labelAPI, labelErr := r.LabelsAPI(o.Server, session.IDToken)
		if labelErr != nil {
			return labelErr
		}
		if o.Command[1] == "assign" {
			if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory) + "; asset: " + strconv.Quote(o.Command[2])); err != nil {
				return err
			}
			if err := r.confirmAction(ctx, o, "Assign asset label", "Assign", "Assign a stable label identity without rendering or printing."); err != nil {
				return err
			}
		}
		result, err = executeLabels(ctx, labelAPI, r.LabelFiles, o)
	} else if isPrintingCommand(o) {
		if r.PrintingAPI == nil {
			return ports.Failure("configuration", "printing API is unavailable")
		}
		printingAPI, printErr := r.PrintingAPI(o.Server, session.IDToken)
		if printErr != nil {
			return printErr
		}
		if o.IdempotencyKey == "" && (o.PrintLabel || o.Command[1] == "print" || o.Command[1] == "test" || o.Command[1] == "reprint") {
			var token [16]byte
			if _, err = rand.Read(token[:]); err != nil {
				return err
			}
			o.IdempotencyKey = hex.EncodeToString(token[:])
		}
		if o.IdempotencyKey != "" {
			if err = r.Output.Notice("Print request key: " + o.IdempotencyKey + "; reuse this key and selection if the response is lost."); err != nil {
				return err
			}
		}
		if o.PrintLabel {
			selection, selectErr := selectPrinter(ctx, printingAPI, o)
			if selectErr != nil {
				return selectErr
			}
			result, err = api.CreateAsset(ctx, o.Scope, ports.AssetInput{Kind: o.Kind, Title: o.Title, Parent: o.Parent, PrintLabel: &selection}, o.IdempotencyKey)
		} else {
			if isPrintResolution(o) {
				result, err = r.resolvePrint(ctx, o, printingAPI)
			} else {
				result, err = executePrinting(ctx, printingAPI, o)
			}
		}
	} else {
		result, err = execute(ctx, api, o)
	}
	if err != nil {
		if isAssetWrite(o) && o.Command[1] == "create" {
			return assetCreateFailure(o, err)
		}
		if isCheckoutWrite(o) {
			return checkoutFailure(err)
		}
		return err
	}
	if isLabelCommand(o) {
		r.Observer.Event(ctx, "cli.label.command.completed")
	} else if isPrintingCommand(o) {
		r.Observer.Event(ctx, "cli.print.command.completed")
	} else {
		r.Observer.Event(ctx, "cli.inventory.command.completed")
	}
	return r.Output.Result(result)
}
func validateCommand(o Options) error      { return validateCommandOptions(o, true) }
func validateCommandShape(o Options) error { return validateCommandOptions(o, false) }
func validateCommandOptions(o Options, requireScope bool) error {
	if isBinaryCommand(o) {
		return validateBinary(o, requireScope)
	}
	if isCheckoutWrite(o) && (o.ConnectorName != "" || o.Title != "" || o.Kind != "" || o.Parent != "") {
		return ports.Failure("usage", "Use --details or --input for checkout notes. Remove unrelated asset field options.")
	}
	if len(o.Command) > 1 && o.Command[0] == "assets" && (o.Command[1] == "update" || o.Command[1] == "move" || isAssetLifecycle(o)) && o.IdempotencyKey != "" {
		return ports.Failure("usage", "This asset action does not support retry keys. Remove --idempotency-key.")
	}
	if o.PrintLabel && (len(o.Command) != 2 || o.Command[0] != "assets" || o.Command[1] != "create") {
		return ports.Failure("usage", "--print-label is only available for assets create")
	}
	if isSearch(o) {
		return validateSearch(o, requireScope)
	}
	if isPairingApproval(o) {
		return validatePairingCommand(o, requireScope)
	}
	if isPrinterAdministration(o) {
		return validatePrinterAdministration(o, requireScope)
	}
	if isTelemetry(o) {
		return validateTelemetry(o)
	}
	if isConnectorInspection(o) {
		return validateConnectorInspection(o, requireScope)
	}
	if isPrintSettingsCommand(o) {
		return validatePrintSettings(o, requireScope)
	}
	if isEvaluationCommand(o) {
		return validateEvaluation(o, requireScope)
	}
	if isVoiceProviderCommand(o) {
		return validateVoiceProvider(o, requireScope)
	}
	if isCustomization(o) {
		return validateCustomization(o, requireScope)
	}
	if isWorkflowCommand(o) {
		return validateWorkflows(o, requireScope)
	}
	if isProviderProfileCommand(o) {
		return validateProviderProfiles(o, requireScope)
	}
	if isArchiveCommand(o) {
		return validateArchive(o, requireScope)
	}
	if isImportJobCommand(o) {
		return validateImportJobs(o, requireScope)
	}
	if isInvitationCommand(o) {
		return validateInvitations(o, requireScope)
	}
	if isAccessGrantCommand(o) {
		return validateAccessGrants(o, requireScope)
	}
	if isActivityCommand(o) {
		return validateActivity(o, requireScope)
	}
	if isAuditCommand(o) {
		return validateAudit(o, requireScope)
	}
	if isOperationCommand(o) {
		return validateOperations(o, requireScope)
	}
	if isPreferenceCommand(o) {
		return validatePreferences(o, requireScope)
	}
	if isNotificationDeviceCommand(o) {
		return validateNotificationDevices(o, requireScope)
	}
	if isNotificationCommand(o) {
		return validateNotifications(o, requireScope)
	}
	if isAttachmentCommand(o) {
		return validateAttachments(o, requireScope)
	}
	if isTagCommand(o) {
		return validateTags(o, requireScope)
	}
	if isLabelCommand(o) {
		return validateLabelCommandOptions(o, requireScope)
	}
	if !o.PrintLabel && isPrintingCommand(o) {
		return validatePrintingCommandOptions(o, requireScope)
	}
	if isDirectoryCommand(o) || isDirectoryWrite(o) || isDirectoryLifecycle(o) {
		if requireScope && missingResourceScope(o) {
			return ports.Failure("usage", "Supply the required scope with --tenant and, for inventory commands, --inventory.")
		}
		return nil
	}
	if len(o.Command) < 2 {
		return ports.Failure("usage", "expected inventories list or assets <action>")
	}
	if requireScope && o.Scope.Tenant == "" {
		return ports.Failure("usage", "choose a tenant with --tenant or STUFF_STASH_CLI_TENANT")
	}
	if o.Command[0] == "inventories" && o.Command[1] == "list" && len(o.Command) == 2 {
		return nil
	}
	if o.Command[0] != "assets" {
		return ports.Failure("usage", "unknown command; use --help")
	}
	if requireScope && o.Scope.Inventory == "" {
		return ports.Failure("usage", "choose an inventory with --inventory or STUFF_STASH_CLI_INVENTORY")
	}
	switch o.Command[1] {
	case "list", "checked-out", "expiration":
		if len(o.Command) == 2 {
			return nil
		}
	case "create":
		if len(o.Command) == 2 && (!requireScope || len(o.RequestBody) > 0 || (o.Title != "" && o.Kind != "")) {
			return nil
		}
	case "show", "archive", "restore", "delete", "checkout", "return", "checkouts":
		if len(o.Command) == 3 {
			return nil
		}
	case "update":
		if len(o.Command) == 3 && (!requireScope || len(o.RequestBody) > 0 || o.Title != "") {
			return nil
		}
	case "return-details":
		if len(o.Command) == 4 {
			return nil
		}
	case "move":
		if len(o.Command) == 3 && o.Parent != "" {
			return nil
		}
	}
	return ports.Failure("usage", "invalid asset command arguments; use --help")
}
func execute(ctx context.Context, api ports.API, o Options) (any, error) {
	if len(o.Command) < 2 || (o.Command[0] != "assets" && !(o.Command[0] == "inventories" && o.Command[1] == "list")) {
		return nil, ports.Failure("usage", "This command has no executor. Use --help to choose a supported command.")
	}
	if o.Command[0] == "inventories" {
		return api.Inventories(ctx, o.Scope, o.Page)
	}
	action := o.Command[1]
	if action == "expiration" {
		q := o.Expiration
		q.Page = o.Page
		q.Kind = o.Kind
		return api.ExpirationAssets(ctx, o.Scope, q)
	}
	if action == "checked-out" {
		return api.CheckedOutAssets(ctx, o.Scope, o.Page)
	}
	if action == "list" {
		return api.Assets(ctx, o.Scope, ports.AssetQuery{Page: o.Page, Lifecycle: o.Lifecycle, Sort: o.Sort})
	}
	id := ""
	if len(o.Command) > 2 {
		id = o.Command[2]
	}
	if action == "show" {
		return api.Asset(ctx, o.Scope, id)
	}
	key := o.IdempotencyKey
	switch action {
	case "create":
		return api.CreateAsset(ctx, o.Scope, ports.AssetInput{Kind: o.Kind, Title: o.Title, Parent: o.Parent, RequestBody: o.RequestBody}, key)
	case "update":
		return api.UpdateAsset(ctx, o.Scope, id, ports.AssetChange{Title: &o.Title, RequestBody: o.RequestBody}, "")
	case "move":
		change := ports.AssetChange{Parent: &o.Parent}
		if o.Parent == "root" {
			change = ports.AssetChange{MoveToRoot: true}
		}
		return api.UpdateAsset(ctx, o.Scope, id, change, "")
	case "checkouts":
		return api.Checkouts(ctx, o.Scope, id, o.Page)
	case "checkout", "return", "return-details":
		checkoutID := ""
		if len(o.Command) == 4 {
			checkoutID = o.Command[3]
		}
		return api.ChangeCheckout(ctx, o.Scope, id, checkoutID, ports.CheckoutAction(action), o.RequestBody)
	case "delete":
		if err := api.DeleteAsset(ctx, o.Scope, id); err != nil {
			return nil, err
		}
		return map[string]string{"status": "deleted", "assetId": id, "tenantId": o.Scope.Tenant, "inventoryId": o.Scope.Inventory}, nil
	case "archive", "restore":
		return api.SetArchived(ctx, o.Scope, id, action == "archive", key)
	}
	return nil, ports.Failure("usage", "unknown action")
}

func isTenantList(o Options) bool {
	return len(o.Command) == 2 && o.Command[0] == "tenants" && o.Command[1] == "list"
}

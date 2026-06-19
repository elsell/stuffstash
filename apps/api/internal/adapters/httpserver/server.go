package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func init() {
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		details := make([]errorDetail, 0, len(errs))
		for _, err := range errs {
			if err == nil {
				continue
			}
			details = append(details, errorDetail{Message: err.Error()})
		}

		return &errorEnvelope{
			status: status,
			BodyError: errorBody{
				Code:    errorCode(status),
				Message: safeErrorMessage(status, msg),
				Details: details,
			},
			Meta: responseMeta{},
		}
	}
}

func NewServer(addr string, application app.App) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("GET /healthz", handleHealth(application))

	config := huma.DefaultConfig("Stuff Stash API", "0.1.0")
	config.DocsPath = "/docs"
	config.OpenAPIPath = "/openapi"
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "dev",
		},
	}

	api := humago.New(mux, config)
	registerRoutes(api, application)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(successEnvelope[indexResponse]{
		Data: indexResponse{
			Service: "stuff-stash",
			Links: indexLinksResponse{
				Health:  "/healthz",
				OpenAPI: "/openapi.json",
				Docs:    "/docs",
			},
		},
		Meta: responseMeta{},
	})
}

func registerRoutes(api huma.API, application app.App) {
	huma.Get(api, "/me", func(ctx context.Context, input *meInput) (*meOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		return &meOutput{
			Body: successEnvelope[principalResponse]{
				Data: principalResponse{ID: principal.ID.String()},
				Meta: responseMeta{},
			},
		}, nil
	}, huma.OperationTags("identity"), securedOperation)

	huma.Post(api, "/tenants", func(ctx context.Context, input *createTenantInput) (*createTenantOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		tenant, err := application.CreateTenant(ctx, app.CreateTenantInput{
			Principal: principal,
			Name:      input.Body.Name,
		})
		if err != nil {
			return nil, toHumaError(err)
		}

		return &createTenantOutput{
			Body: successEnvelope[tenantResponse]{
				Data: tenantResponse{
					ID:   tenant.ID.String(),
					Name: tenant.Name.String(),
				},
				Meta: responseMeta{TenantID: tenant.ID.String()},
			},
		}, nil
	}, huma.OperationTags("tenants"), createdOperation, securedOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories", func(ctx context.Context, input *createInventoryInput) (*createInventoryOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		item, err := application.CreateInventory(ctx, app.CreateInventoryInput{
			Principal: principal,
			TenantID:  tenant.ID(input.TenantID),
			Name:      input.Body.Name,
		})
		if err != nil {
			return nil, toHumaError(err)
		}

		return &createInventoryOutput{
			Body: successEnvelope[inventoryResponse]{
				Data: inventoryResponse{
					ID:       item.ID.String(),
					TenantID: item.TenantID.String(),
					Name:     item.Name.String(),
				},
				Meta: responseMeta{TenantID: item.TenantID.String()},
			},
		}, nil
	}, huma.OperationTags("inventories"), createdOperation, securedOperation)

	huma.Get(api, "/tenants/{tenantId}/inventories", func(ctx context.Context, input *listInventoriesInput) (*listInventoriesOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListInventories(ctx, app.ListInventoriesInput{
			Principal: principal,
			TenantID:  tenant.ID(input.TenantID),
			Limit:     input.Limit,
			Cursor:    input.Cursor,
		})
		if err != nil {
			return nil, toHumaError(err)
		}

		data := make([]inventoryResponse, 0, len(result.Items))
		for _, item := range result.Items {
			data = append(data, inventoryResponse{
				ID:       item.ID.String(),
				TenantID: item.TenantID.String(),
				Name:     item.Name.String(),
			})
		}

		return &listInventoriesOutput{
			Body: successEnvelope[[]inventoryResponse]{
				Data: data,
				Meta: responseMeta{
					TenantID: input.TenantID,
					Pagination: &paginationMeta{
						Limit:      result.Limit,
						NextCursor: result.NextCursor,
						HasMore:    result.HasMore,
					},
				},
			},
		}, nil
	}, huma.OperationTags("inventories"), securedOperation)

	huma.Post(api, "/tenants/{tenantId}/custom-field-definitions", func(ctx context.Context, input *createTenantCustomFieldDefinitionInput) (*createCustomFieldDefinitionOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		definition, err := application.CreateTenantCustomFieldDefinition(ctx, app.CreateCustomFieldDefinitionInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			Key:         input.Body.Key,
			DisplayName: input.Body.DisplayName,
			Type:        input.Body.Type,
			EnumOptions: input.Body.EnumOptions,
		})
		if err != nil {
			return nil, toHumaError(err)
		}

		return &createCustomFieldDefinitionOutput{
			Body: successEnvelope[customFieldDefinitionResponse]{
				Data: customFieldDefinitionToResponse(definition),
				Meta: responseMeta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("custom field definitions"), createdOperation, securedOperation)

	huma.Get(api, "/tenants/{tenantId}/custom-field-definitions", func(ctx context.Context, input *listTenantCustomFieldDefinitionsInput) (*listCustomFieldDefinitionsOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListTenantCustomFieldDefinitions(ctx, app.ListCustomFieldDefinitionsInput{
			Principal: principal,
			TenantID:  tenant.ID(input.TenantID),
			Limit:     input.Limit,
			Cursor:    input.Cursor,
		})
		if err != nil {
			return nil, toHumaError(err)
		}

		return customFieldDefinitionsOutput(input.TenantID, result), nil
	}, huma.OperationTags("custom field definitions"), securedOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets", func(ctx context.Context, input *createAssetInput) (*createAssetOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		item, err := application.CreateAsset(ctx, app.CreateAssetInput{
			Principal:     principal,
			TenantID:      tenant.ID(input.TenantID),
			InventoryID:   inventory.InventoryID(input.InventoryID),
			Kind:          input.Body.Kind,
			Title:         input.Body.Title,
			Description:   input.Body.Description,
			ParentAssetID: input.Body.ParentAssetID,
			CustomFields:  input.Body.CustomFields,
		})
		if err != nil {
			return nil, toHumaError(err)
		}

		return &createAssetOutput{
			Body: successEnvelope[assetResponse]{
				Data: assetToResponse(item),
				Meta: responseMeta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("assets"), createdOperation, securedOperation)

	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets", func(ctx context.Context, input *listAssetsInput) (*listAssetsOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListAssets(ctx, app.ListAssetsInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Limit:       input.Limit,
			Cursor:      input.Cursor,
		})
		if err != nil {
			return nil, toHumaError(err)
		}

		data := make([]assetResponse, 0, len(result.Items))
		for _, item := range result.Items {
			data = append(data, assetToResponse(item))
		}

		return &listAssetsOutput{
			Body: successEnvelope[[]assetResponse]{
				Data: data,
				Meta: responseMeta{
					TenantID: input.TenantID,
					Pagination: &paginationMeta{
						Limit:      result.Limit,
						NextCursor: result.NextCursor,
						HasMore:    result.HasMore,
					},
				},
			},
		}, nil
	}, huma.OperationTags("assets"), securedOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions", func(ctx context.Context, input *createInventoryCustomFieldDefinitionInput) (*createCustomFieldDefinitionOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		definition, err := application.CreateInventoryCustomFieldDefinition(ctx, app.CreateCustomFieldDefinitionInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Key:         input.Body.Key,
			DisplayName: input.Body.DisplayName,
			Type:        input.Body.Type,
			EnumOptions: input.Body.EnumOptions,
		})
		if err != nil {
			return nil, toHumaError(err)
		}

		return &createCustomFieldDefinitionOutput{
			Body: successEnvelope[customFieldDefinitionResponse]{
				Data: customFieldDefinitionToResponse(definition),
				Meta: responseMeta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("custom field definitions"), createdOperation, securedOperation)

	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions", func(ctx context.Context, input *listInventoryCustomFieldDefinitionsInput) (*listCustomFieldDefinitionsOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListInventoryCustomFieldDefinitions(ctx, app.ListCustomFieldDefinitionsInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Limit:       input.Limit,
			Cursor:      input.Cursor,
		})
		if err != nil {
			return nil, toHumaError(err)
		}

		return customFieldDefinitionsOutput(input.TenantID, result), nil
	}, huma.OperationTags("custom field definitions"), securedOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants", func(ctx context.Context, input *grantInventoryAccessInput) (*grantInventoryAccessOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		grant, err := application.GrantInventoryAccess(ctx, app.GrantInventoryAccessInput{
			Principal:    principal,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			TargetUserID: input.Body.PrincipalID,
			Relationship: input.Body.Relationship,
		})
		if err != nil {
			return nil, toHumaError(err)
		}

		return &grantInventoryAccessOutput{
			Body: successEnvelope[inventoryAccessGrantResponse]{
				Data: inventoryAccessGrantToResponse(grant),
				Meta: responseMeta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("inventory access"), createdOperation, securedOperation)

	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants", func(ctx context.Context, input *listInventoryAccessInput) (*listInventoryAccessOutput, error) {
		principal, err := authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListInventoryAccessGrants(ctx, app.ListInventoryAccessGrantsInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Limit:       input.Limit,
			Cursor:      input.Cursor,
		})
		if err != nil {
			return nil, toHumaError(err)
		}

		data := make([]inventoryAccessGrantResponse, 0, len(result.Items))
		for _, grant := range result.Items {
			data = append(data, inventoryAccessGrantToResponse(grant))
		}

		return &listInventoryAccessOutput{
			Body: successEnvelope[[]inventoryAccessGrantResponse]{
				Data: data,
				Meta: responseMeta{
					TenantID: input.TenantID,
					Pagination: &paginationMeta{
						Limit:      result.Limit,
						NextCursor: result.NextCursor,
						HasMore:    result.HasMore,
					},
				},
			},
		}, nil
	}, huma.OperationTags("inventory access"), securedOperation)
}

func securedOperation(operation *huma.Operation) {
	operation.Security = []map[string][]string{{"bearerAuth": {}}}
}

func createdOperation(operation *huma.Operation) {
	operation.DefaultStatus = http.StatusCreated
}

func handleHealth(application app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := application.Health(r.Context())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"service":"` + string(status.Service) + `","status":"` + string(status.Status) + `"}` + "\n"))
	}
}

func authenticate(ctx context.Context, application app.App, authorization string) (identity.Principal, error) {
	principal, err := application.Authenticate(ctx, authorization)
	if err != nil {
		return identity.Principal{}, toHumaError(err)
	}
	return principal, nil
}

func toHumaError(err error) error {
	switch {
	case errors.Is(err, app.ErrUnauthenticated):
		return huma.Error401Unauthorized("Authentication required.")
	case errors.Is(err, app.ErrUnauthorized):
		return huma.Error403Forbidden("Forbidden.")
	case errors.Is(err, app.ErrInvalidInput):
		return huma.Error400BadRequest("Invalid request.")
	case errors.Is(err, app.ErrNotFound):
		return huma.Error404NotFound("Resource not found.")
	default:
		return huma.Error500InternalServerError("Internal server error.")
	}
}

func errorCode(status int) string {
	switch status {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return "invalid_request"
	case http.StatusUnauthorized:
		return "authentication_required"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "resource_not_found"
	default:
		return "internal_error"
	}
}

func safeErrorMessage(status int, fallback string) string {
	switch status {
	case http.StatusUnauthorized:
		return "Authentication required."
	case http.StatusForbidden:
		return "Forbidden."
	case http.StatusNotFound:
		return "Resource not found."
	case http.StatusInternalServerError:
		return "Internal server error."
	}
	if fallback == "" {
		return "Invalid request."
	}
	return fallback
}

type meInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
}

type meOutput struct {
	Body successEnvelope[principalResponse]
}

type createTenantInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	Body          struct {
		Name string `json:"name" maxLength:"120" doc:"Tenant name"`
	}
}

type createTenantOutput struct {
	Body successEnvelope[tenantResponse]
}

type createInventoryInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Body          struct {
		Name string `json:"name" maxLength:"120" doc:"Inventory name"`
	}
}

type createInventoryOutput struct {
	Body successEnvelope[inventoryResponse]
}

type listInventoriesInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type listInventoriesOutput struct {
	Body successEnvelope[[]inventoryResponse]
}

type createTenantCustomFieldDefinitionInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Body          struct {
		Key         string   `json:"key" maxLength:"80" doc:"Stable custom field key"`
		DisplayName string   `json:"displayName" maxLength:"120" doc:"User-facing field label"`
		Type        string   `json:"type" enum:"text,number,boolean,date,url,enum" doc:"Custom field type"`
		EnumOptions []string `json:"enumOptions,omitempty" doc:"Allowed enum option keys"`
	}
}

type createInventoryCustomFieldDefinitionInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          struct {
		Key         string   `json:"key" maxLength:"80" doc:"Stable custom field key"`
		DisplayName string   `json:"displayName" maxLength:"120" doc:"User-facing field label"`
		Type        string   `json:"type" enum:"text,number,boolean,date,url,enum" doc:"Custom field type"`
		EnumOptions []string `json:"enumOptions,omitempty" doc:"Allowed enum option keys"`
	}
}

type createCustomFieldDefinitionOutput struct {
	Body successEnvelope[customFieldDefinitionResponse]
}

type listTenantCustomFieldDefinitionsInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type listInventoryCustomFieldDefinitionsInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type listCustomFieldDefinitionsOutput struct {
	Body successEnvelope[[]customFieldDefinitionResponse]
}

type createAssetInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          struct {
		Kind          string         `json:"kind" enum:"item,container,location" doc:"Asset kind"`
		Title         string         `json:"title" maxLength:"160" doc:"Asset title"`
		Description   string         `json:"description,omitempty" doc:"Asset description"`
		ParentAssetID string         `json:"parentAssetId,omitempty" doc:"Parent asset ID"`
		CustomFields  map[string]any `json:"customFields,omitempty" doc:"Custom field values"`
	}
}

type createAssetOutput struct {
	Body successEnvelope[assetResponse]
}

type listAssetsInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type listAssetsOutput struct {
	Body successEnvelope[[]assetResponse]
}

type grantInventoryAccessInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          struct {
		PrincipalID  string `json:"principalId" doc:"User principal ID to grant access to"`
		Relationship string `json:"relationship" enum:"viewer,editor" doc:"Direct inventory relationship"`
	}
}

type grantInventoryAccessOutput struct {
	Body successEnvelope[inventoryAccessGrantResponse]
}

type listInventoryAccessInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type listInventoryAccessOutput struct {
	Body successEnvelope[[]inventoryAccessGrantResponse]
}

type successEnvelope[T any] struct {
	Data T            `json:"data"`
	Meta responseMeta `json:"meta"`
}

type responseMeta struct {
	RequestID  string          `json:"requestId,omitempty"`
	TenantID   string          `json:"tenantId,omitempty"`
	Pagination *paginationMeta `json:"pagination,omitempty"`
}

type paginationMeta struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
}

type errorEnvelope struct {
	status    int
	BodyError errorBody    `json:"error"`
	Meta      responseMeta `json:"meta"`
}

func (e *errorEnvelope) Error() string {
	return e.BodyError.Message
}

func (e *errorEnvelope) GetStatus() int {
	return e.status
}

type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []errorDetail `json:"details"`
}

type errorDetail struct {
	Message string `json:"message"`
}

type indexResponse struct {
	Service string             `json:"service"`
	Links   indexLinksResponse `json:"links"`
}

type indexLinksResponse struct {
	Health  string `json:"health"`
	OpenAPI string `json:"openapi"`
	Docs    string `json:"docs"`
}

type principalResponse struct {
	ID string `json:"id"`
}

type tenantResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type inventoryResponse struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Name     string `json:"name"`
}

type customFieldDefinitionResponse struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenantId"`
	InventoryID string   `json:"inventoryId,omitempty"`
	Scope       string   `json:"scope"`
	Key         string   `json:"key"`
	DisplayName string   `json:"displayName"`
	Type        string   `json:"type"`
	EnumOptions []string `json:"enumOptions"`
}

type assetResponse struct {
	ID             string         `json:"id"`
	TenantID       string         `json:"tenantId"`
	InventoryID    string         `json:"inventoryId"`
	ParentAssetID  string         `json:"parentAssetId,omitempty"`
	Kind           string         `json:"kind"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	CustomFields   map[string]any `json:"customFields"`
	LifecycleState string         `json:"lifecycleState"`
}

type inventoryAccessGrantResponse struct {
	TenantID     string `json:"tenantId"`
	InventoryID  string `json:"inventoryId"`
	PrincipalID  string `json:"principalId"`
	Relationship string `json:"relationship"`
}

func assetToResponse(item asset.Asset) assetResponse {
	return assetResponse{
		ID:             item.ID.String(),
		TenantID:       item.TenantID.String(),
		InventoryID:    item.InventoryID.String(),
		ParentAssetID:  item.ParentAssetID.String(),
		Kind:           item.Kind.String(),
		Title:          item.Title.String(),
		Description:    item.Description.String(),
		CustomFields:   item.CustomFields.Values(),
		LifecycleState: item.LifecycleState.String(),
	}
}

func customFieldDefinitionToResponse(definition customfield.Definition) customFieldDefinitionResponse {
	options := make([]string, 0, len(definition.EnumOptions))
	for _, option := range definition.EnumOptions {
		options = append(options, option.String())
	}
	return customFieldDefinitionResponse{
		ID:          definition.ID.String(),
		TenantID:    definition.TenantID.String(),
		InventoryID: definition.InventoryID.String(),
		Scope:       definition.Scope.String(),
		Key:         definition.Key.String(),
		DisplayName: definition.DisplayName.String(),
		Type:        definition.Type.String(),
		EnumOptions: options,
	}
}

func customFieldDefinitionsOutput(tenantID string, result app.ListCustomFieldDefinitionsResult) *listCustomFieldDefinitionsOutput {
	data := make([]customFieldDefinitionResponse, 0, len(result.Items))
	for _, definition := range result.Items {
		data = append(data, customFieldDefinitionToResponse(definition))
	}
	return &listCustomFieldDefinitionsOutput{
		Body: successEnvelope[[]customFieldDefinitionResponse]{
			Data: data,
			Meta: responseMeta{
				TenantID: tenantID,
				Pagination: &paginationMeta{
					Limit:      result.Limit,
					NextCursor: result.NextCursor,
					HasMore:    result.HasMore,
				},
			},
		},
	}
}

func inventoryAccessGrantToResponse(grant ports.InventoryAccessGrant) inventoryAccessGrantResponse {
	return inventoryAccessGrantResponse{
		TenantID:     grant.TenantID.String(),
		InventoryID:  grant.InventoryID.String(),
		PrincipalID:  grant.PrincipalID.String(),
		Relationship: string(grant.Relationship),
	}
}

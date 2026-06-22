package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type CreateTenantInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	Body          CreateTenantBody
}

type CreateTenantBody struct {
	Name string `json:"name" maxLength:"120" doc:"Tenant name"`
}

type CreateTenantOutput struct {
	Body shared.SuccessEnvelope[TenantResponse]
}

type GetTenantInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
}

type GetTenantOutput struct {
	Body shared.SuccessEnvelope[TenantResponse]
}

type UpdateTenantInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	Body          UpdateTenantBody
}

type UpdateTenantBody struct {
	Name *string `json:"name,omitempty" maxLength:"120" doc:"Tenant name"`
}

type UpdateTenantOutput struct {
	Body shared.SuccessEnvelope[TenantResponse]
}

type UpdateTenantLifecycleInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
}

type UpdateTenantLifecycleOutput struct {
	Body shared.SuccessEnvelope[TenantResponse]
}

type DeleteTenantOutput struct{}

type TenantResponse struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	LifecycleState string                `json:"lifecycleState"`
	Access         shared.AccessResponse `json:"access"`
}

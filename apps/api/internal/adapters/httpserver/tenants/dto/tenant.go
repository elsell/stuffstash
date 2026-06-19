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

type TenantResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

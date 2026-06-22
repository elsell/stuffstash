package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type MeInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
}

type MeOutput struct {
	Body shared.SuccessEnvelope[PrincipalResponse]
}

type ListMyTenantsInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListMyTenantsOutput struct {
	Body shared.SuccessEnvelope[[]MyTenantResponse]
}

type PrincipalResponse struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
}

type MyTenantResponse struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	LifecycleState string                `json:"lifecycleState"`
	Access         shared.AccessResponse `json:"access"`
}

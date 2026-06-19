package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type MeInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
}

type MeOutput struct {
	Body shared.SuccessEnvelope[PrincipalResponse]
}

type PrincipalResponse struct {
	ID string `json:"id"`
}

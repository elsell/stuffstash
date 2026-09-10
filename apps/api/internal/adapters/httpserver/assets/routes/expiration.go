package routes

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/app/assets"
)

func expirationInput(value *dto.Expiration) *assets.ExpirationInput {
	if value == nil {
		return nil
	}
	return &assets.ExpirationInput{Date: value.Date, Precision: value.Precision}
}

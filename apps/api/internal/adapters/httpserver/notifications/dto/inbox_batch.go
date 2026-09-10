package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type InboxBatchInput struct {
	ScopeInput
	Cursor string `query:"cursor" maxLength:"128"`
}
type UnreadCountResponse struct {
	Count int `json:"count" minimum:"0"`
}
type UnreadCountOutput struct {
	Body shared.SuccessEnvelope[UnreadCountResponse]
}
type InboxReadAllResponse struct {
	Complete bool `json:"complete"`
}
type InboxReadAllOutput struct {
	Body shared.SuccessEnvelope[InboxReadAllResponse]
}

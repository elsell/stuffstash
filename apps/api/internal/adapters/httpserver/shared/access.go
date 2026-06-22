package shared

import "github.com/stuffstash/stuff-stash/internal/app"

type AccessResponse struct {
	Relationship string   `json:"relationship"`
	Permissions  []string `json:"permissions"`
}

func AccessToResponse(access app.AccessSummary) AccessResponse {
	return AccessResponse{
		Relationship: string(access.Relationship),
		Permissions:  append([]string(nil), access.Permissions...),
	}
}

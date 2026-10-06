package ports

import "context"

type Tag struct {
	ID          string  `json:"id"`
	TenantID    string  `json:"tenantId"`
	InventoryID string  `json:"inventoryId"`
	Key         string  `json:"key"`
	DisplayName string  `json:"displayName"`
	Color       *string `json:"color,omitempty"`
	Lifecycle   string  `json:"lifecycleState"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}
type TagAction string

const (
	CreateTag TagAction = "create"
	UpdateTag TagAction = "update"
	DeleteTag TagAction = "delete"
)

type TagsAPI interface {
	Tags(context.Context, Scope, Page) (Result[[]Tag], error)
	ChangeTag(context.Context, Scope, TagAction, string, []byte) (Result[Tag], error)
}

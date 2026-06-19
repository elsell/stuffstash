package shared

type SuccessEnvelope[T any] struct {
	Data T    `json:"data"`
	Meta Meta `json:"meta"`
}

type Meta struct {
	RequestID  string          `json:"requestId,omitempty"`
	TenantID   string          `json:"tenantId,omitempty"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

type PaginationMeta struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
}

func PaginatedMeta(tenantID string, limit int, nextCursor *string, hasMore bool) Meta {
	return Meta{
		TenantID: tenantID,
		Pagination: &PaginationMeta{
			Limit:      limit,
			NextCursor: nextCursor,
			HasMore:    hasMore,
		},
	}
}

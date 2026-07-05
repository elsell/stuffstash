package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
)

type ImportSourceRequest struct {
	SourceType          importplan.SourceType
	BaseURL             string
	Username            string
	Password            string
	IncludeImages       bool
	AllowInsecureTLS    bool
	AllowPrivateNetwork bool
	MaxAttachmentBytes  int64
	FileName            string
	Content             []byte
}

type ImportSourceReader interface {
	ReadImportPlan(ctx context.Context, request ImportSourceRequest) (importplan.Plan, error)
}

type ImportSourceUserError struct {
	Detail string
}

func (e ImportSourceUserError) Error() string {
	return e.Detail
}

func NewImportSourceUserError(detail string) error {
	return ImportSourceUserError{Detail: detail}
}

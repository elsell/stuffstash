package credentials

import (
	"context"
	"encoding/json"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ImportJobSourceSealer interface {
	SealImportJobSource(ctx context.Context, scope ports.ImportJobSourceScope, raw []byte) (ports.SealedImportJobSource, error)
	UnsealImportJobSource(ctx context.Context, scope ports.ImportJobSourceScope, sealed ports.SealedImportJobSource) ([]byte, error)
}

type DatabaseImportJobSourceVault struct {
	repository ports.ImportJobSourceRepository
	sealer     ImportJobSourceSealer
	clock      ports.Clock
}

func NewDatabaseImportJobSourceVault(repository ports.ImportJobSourceRepository, sealer ImportJobSourceSealer) DatabaseImportJobSourceVault {
	return NewDatabaseImportJobSourceVaultWithClock(repository, sealer, ports.SystemClock{})
}

func NewDatabaseImportJobSourceVaultWithClock(repository ports.ImportJobSourceRepository, sealer ImportJobSourceSealer, clock ports.Clock) DatabaseImportJobSourceVault {
	if clock == nil {
		clock = ports.SystemClock{}
	}
	return DatabaseImportJobSourceVault{repository: repository, sealer: sealer, clock: clock}
}

func (v DatabaseImportJobSourceVault) StoreImportJobSource(ctx context.Context, scope ports.ImportJobSourceScope, request ports.ImportSourceRequest, expiresAt time.Time, now time.Time) error {
	if v.repository == nil || v.sealer == nil || expiresAt.IsZero() || now.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	payload, err := json.Marshal(importSourcePayloadFromRequest(request))
	if err != nil {
		return ports.ErrInvalidProviderInput
	}
	sealed, err := v.sealer.SealImportJobSource(ctx, scope, payload)
	if err != nil {
		return ports.ErrInvalidProviderInput
	}
	return v.repository.ReplaceImportJobSource(ctx, ports.ImportJobSourceRecord{
		Scope:     scope,
		Sealed:    sealed,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (v DatabaseImportJobSourceVault) ImportJobSourceRequest(ctx context.Context, scope ports.ImportJobSourceScope) (ports.ImportSourceRequest, bool, error) {
	if v.repository == nil || v.sealer == nil {
		return ports.ImportSourceRequest{}, false, ports.ErrInvalidProviderInput
	}
	record, found, err := v.repository.ImportJobSource(ctx, scope)
	if err != nil || !found {
		return ports.ImportSourceRequest{}, found, err
	}
	if !record.ExpiresAt.After(v.clock.Now().UTC()) {
		_, _ = v.repository.DeleteImportJobSource(ctx, scope)
		return ports.ImportSourceRequest{}, false, nil
	}
	raw, err := v.sealer.UnsealImportJobSource(ctx, scope, record.Sealed)
	if err != nil {
		return ports.ImportSourceRequest{}, false, ports.ErrImportJobSourceUnreadable
	}
	var payload importSourcePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ports.ImportSourceRequest{}, false, ports.ErrImportJobSourceUnreadable
	}
	return payload.toRequest(), true, nil
}

func (v DatabaseImportJobSourceVault) DeleteImportJobSource(ctx context.Context, scope ports.ImportJobSourceScope) (bool, error) {
	if v.repository == nil {
		return false, ports.ErrInvalidProviderInput
	}
	return v.repository.DeleteImportJobSource(ctx, scope)
}

func (v DatabaseImportJobSourceVault) VacuumImportJobSources(ctx context.Context, now time.Time) ([]ports.ImportJobSourceScope, error) {
	if v.repository == nil || now.IsZero() {
		return nil, ports.ErrInvalidProviderInput
	}
	return v.repository.DeleteVacuumableImportJobSources(ctx, terminalImportJobStatuses(), now)
}

func terminalImportJobStatuses() []importjob.Status {
	return []importjob.Status{
		importjob.StatusSucceeded,
		importjob.StatusFailed,
		importjob.StatusCancelledKept,
		importjob.StatusCancelledDiscarded,
		importjob.StatusDiscardFailed,
	}
}

type importSourcePayload struct {
	SourceType           string `json:"sourceType"`
	BaseURL              string `json:"baseUrl,omitempty"`
	Username             string `json:"username,omitempty"`
	Password             string `json:"password,omitempty"`
	IncludeImages        bool   `json:"includeImages,omitempty"`
	FetchAttachmentBytes bool   `json:"fetchAttachmentBytes,omitempty"`
	AllowInsecureTLS     bool   `json:"allowInsecureTLS,omitempty"`
	AllowPrivateNetwork  bool   `json:"allowPrivateNetwork,omitempty"`
	MaxAttachmentBytes   int64  `json:"maxAttachmentBytes,omitempty"`
	FileName             string `json:"fileName,omitempty"`
	Content              []byte `json:"content,omitempty"`
}

func importSourcePayloadFromRequest(request ports.ImportSourceRequest) importSourcePayload {
	return importSourcePayload{
		SourceType:           string(request.SourceType),
		BaseURL:              request.BaseURL,
		Username:             request.Username,
		Password:             request.Password,
		IncludeImages:        request.IncludeImages,
		FetchAttachmentBytes: request.FetchAttachmentBytes,
		AllowInsecureTLS:     request.AllowInsecureTLS,
		AllowPrivateNetwork:  request.AllowPrivateNetwork,
		MaxAttachmentBytes:   request.MaxAttachmentBytes,
		FileName:             request.FileName,
		Content:              append([]byte{}, request.Content...),
	}
}

func (p importSourcePayload) toRequest() ports.ImportSourceRequest {
	return ports.ImportSourceRequest{
		SourceType:           importplan.SourceType(p.SourceType),
		BaseURL:              p.BaseURL,
		Username:             p.Username,
		Password:             p.Password,
		IncludeImages:        p.IncludeImages,
		FetchAttachmentBytes: p.FetchAttachmentBytes,
		AllowInsecureTLS:     p.AllowInsecureTLS,
		AllowPrivateNetwork:  p.AllowPrivateNetwork,
		MaxAttachmentBytes:   p.MaxAttachmentBytes,
		FileName:             p.FileName,
		Content:              append([]byte{}, p.Content...),
	}
}

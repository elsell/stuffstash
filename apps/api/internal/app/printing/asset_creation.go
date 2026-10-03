package printing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *JobService) CreateAssetAndPrint(ctx context.Context, assets ports.AssetPrintPreparation, uow ports.AssetPrintUnitOfWork, input ports.CreateAssetInput, selection JobSelection, key string) (ports.AssetPrintResult, error) {
	scope := LabelScope{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, RequestID: input.RequestID}
	if err := assets.AuthorizeAssetCreation(ctx, input); err != nil {
		return ports.AssetPrintResult{}, err
	}
	if err := s.access(ctx, scope, true); err != nil {
		return ports.AssetPrintResult{}, err
	}
	if uow == nil || len(key) < 1 || len(key) > 180 || selection.Copies < 1 || selection.Copies > s.config.MaxCopies {
		return ports.AssetPrintResult{}, apperrors.ErrInvalidInput
	}
	canonical := input
	canonical.RequestID = ""
	canonical.Principal = identity.Principal{ID: input.Principal.ID}
	canonical.Source = ""
	payload, err := json.Marshal(struct {
		Draft     ports.CreateAssetInput
		Selection JobSelection
	}{canonical, selection})
	if err != nil {
		return ports.AssetPrintResult{}, apperrors.ErrInvalidInput
	}
	hash := sha256.Sum256(payload)
	fingerprint := hex.EncodeToString(hash[:])
	key = "asset-create:" + key
	existing, prior, err := s.jobs.FindPrintJobRequest(ctx, jobScope(scope), string(input.Principal.ID), key)
	if err == nil {
		if prior != fingerprint {
			return ports.AssetPrintResult{}, apperrors.ErrConflict
		}
		item, err := s.labels.asset(ctx, scope, asset.ID(existing.AssetID))
		if err != nil {
			return ports.AssetPrintResult{}, err
		}
		return ports.AssetPrintResult{Asset: item, Job: existing}, nil
	}
	if !errors.Is(err, ports.ErrPrintJobNotFound) {
		return ports.AssetPrintResult{}, err
	}
	prepared, err := assets.PrepareCreateAssetForPrint(ctx, input)
	if err != nil {
		return ports.AssetPrintResult{}, err
	}
	printer, err := s.printers.GetPrinter(ctx, jobScope(scope), selection.PrinterID)
	if err != nil {
		return ports.AssetPrintResult{}, jobError(err)
	}
	if printer.Retired || printer.MediaFingerprint != selection.ExpectedMediaFingerprint {
		return ports.AssetPrintResult{}, apperrors.ErrConflict
	}
	instance, err := s.labels.Instance(ctx)
	if err != nil {
		return ports.AssetPrintResult{}, err
	}
	now := s.labels.deps.Clock.Now()
	label := p.Label{ID: p.LabelID(s.labels.deps.IDs.NewID()), InstanceID: instance, TenantID: scope.TenantID.String(), InventoryID: scope.InventoryID.String(), AssetID: prepared.Asset.ID.String(), CreatedAt: now}
	labelAudit, err := appsupport.NewAuditRecord(s.labels.deps.IDs, s.labels.deps.Clock, s.labels.auditInput(scope, audit.ActionLabelProvisioned, label.AssetID))
	if err != nil {
		return ports.AssetPrintResult{}, err
	}
	view := s.labels.view(label, prepared.Asset)
	request := p.RenderRequest{Content: p.ContentSnapshot{QRURL: view.URL, Title: prepared.Asset.Title.String(), Reference: string(label.ID)}, Template: selection.Template, Media: printer.Media, Format: p.FormatPNG}
	job, content, err := s.prepareJob(ctx, scope, printer, selection, request, label.AssetID, string(label.ID), key)
	if err != nil {
		return ports.AssetPrintResult{}, err
	}
	job.AssetCreationOperationID = prepared.UndoableOperation.ID
	record, err := s.audit(scope, audit.ActionPrintJobQueued, job.ID)
	if err != nil {
		return ports.AssetPrintResult{}, err
	}
	// Recheck current access after rendering, before the atomic command.
	if err = assets.AuthorizeAssetCreation(ctx, input); err != nil {
		return ports.AssetPrintResult{}, err
	}
	if err = s.access(ctx, scope, true); err != nil {
		return ports.AssetPrintResult{}, err
	}
	result, err := uow.CreateAssetWithPrint(ctx, ports.PreparedAssetPrint{Asset: prepared, Label: label, LabelAudit: labelAudit, Job: ports.PrintJobCreate{Job: job, Content: content, PrinterRevision: printer.Revision, RequestFingerprint: fingerprint, Audit: record}})
	if err != nil {
		return ports.AssetPrintResult{}, jobError(err)
	}
	if result.Created {
		assets.RecordAssetCreated(ctx, result.Asset, input.Principal.ID)
		if s.labels.deps.Observer != nil {
			s.labels.deps.Observer.Record(ctx, ports.Event{Name: ports.EventLabelProvisioned})
			s.labels.deps.Observer.Record(ctx, ports.Event{Name: ports.EventPrintJobQueued})
		}
	}
	return result, nil
}

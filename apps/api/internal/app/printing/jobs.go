package printing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	label "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type JobConfig struct {
	MaxCopies, MaxArtifactBytes                                       int
	ArtifactTTL, TerminalTTL, Lease, ReadinessMaxAge, CleanupInterval time.Duration
}
type JobService struct {
	labels   *LabelService
	jobs     ports.PrintJobRepository
	printers ports.PrinterRepository
	config   JobConfig
}

func NewJobService(labels *LabelService, jobs ports.PrintJobRepository, printers ports.PrinterRepository, config JobConfig) *JobService {
	return &JobService{labels: labels, jobs: jobs, printers: printers, config: config}
}
func jobScope(s LabelScope) label.Scope {
	return label.Scope{TenantID: s.TenantID.String(), InventoryID: s.InventoryID.String()}
}
func jobError(err error) error {
	switch {
	case errors.Is(err, ports.ErrPrintJobNotFound), errors.Is(err, ports.ErrPrintNotFound):
		return apperrors.ErrNotFound
	case errors.Is(err, ports.ErrConflict), errors.Is(err, ports.ErrPrintConflict), errors.Is(err, label.ErrJobConflict), errors.Is(err, label.ErrLeaseExpired):
		return apperrors.ErrConflict
	case errors.Is(err, label.ErrInvalidOutcome):
		return apperrors.ErrInvalidInput
	case errors.Is(err, ports.ErrPrintDenied), errors.Is(err, label.ErrAttemptOwnership):
		return ports.ErrForbidden
	}
	return err
}
func (s *JobService) access(ctx context.Context, scope LabelScope, mutate bool) error {
	if s == nil || s.labels == nil || s.jobs == nil || s.printers == nil {
		return ErrLabelsUnavailable
	}
	if err := s.labels.access(ctx, scope); err != nil {
		return err
	}
	if mutate {
		return s.labels.deps.Authorizer.CheckInventory(ctx, scope.Principal, ports.InventoryPermissionEditAsset, scope.InventoryID)
	}
	return nil
}
func (s *JobService) audit(scope LabelScope, action audit.Action, id label.JobID) (audit.Record, error) {
	target := audit.TargetPrintJob
	targetID := string(id)
	if id == "" {
		target = audit.TargetInventory
		targetID = scope.InventoryID.String()
	}
	return appsupport.NewAuditRecord(s.labels.deps.IDs, s.labels.deps.Clock, appsupport.AuditRecordInput{Principal: scope.Principal, TenantID: scope.TenantID, InventoryID: scope.InventoryID, RequestID: scope.RequestID, Source: audit.SourceAPI, Action: action, TargetType: target, TargetID: targetID})
}
func (s *JobService) Get(ctx context.Context, scope LabelScope, id label.JobID) (label.Job, error) {
	if err := s.access(ctx, scope, false); err != nil {
		return label.Job{}, err
	}
	job, err := s.jobs.GetPrintJob(ctx, jobScope(scope), id)
	if err != nil {
		return label.Job{}, jobError(err)
	}
	record, err := s.audit(scope, audit.ActionPrintJobViewed, id)
	if err != nil {
		return label.Job{}, err
	}
	if err = s.labels.deps.Audit.SaveAuditRecord(ctx, record); err != nil {
		return label.Job{}, err
	}
	return job, nil
}

type JobPage struct {
	Items      []label.Job
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (s *JobService) List(ctx context.Context, scope LabelScope, printer label.PrinterID, limit int, cursor string) (JobPage, error) {
	if err := s.access(ctx, scope, false); err != nil {
		return JobPage{}, err
	}
	if limit < 1 || limit > 100 {
		return JobPage{}, apperrors.ErrInvalidInput
	}
	cursorScope := scope.TenantID.String() + ":" + scope.InventoryID.String() + ":" + string(printer)
	after, err := appsupport.DecodePageCursor("print-jobs", cursorScope, cursor)
	if err != nil {
		return JobPage{}, err
	}
	jobs, err := s.jobs.ListPrintJobs(ctx, jobScope(scope), printer, limit+1, after)
	if err != nil {
		return JobPage{}, jobError(err)
	}
	page := JobPage{Items: jobs, Limit: limit, HasMore: len(jobs) > limit}
	if page.HasMore {
		page.Items = jobs[:limit]
		page.NextCursor = appsupport.EncodePageCursor("print-jobs", cursorScope, string(page.Items[len(page.Items)-1].ID))
	}
	record, err := s.audit(scope, audit.ActionPrintJobsListed, "")
	if err != nil {
		return JobPage{}, err
	}
	if err = s.labels.deps.Audit.SaveAuditRecord(ctx, record); err != nil {
		return JobPage{}, err
	}
	return page, nil
}
func (s *JobService) Cancel(ctx context.Context, scope LabelScope, id label.JobID, revision uint64) (label.Job, error) {
	if err := s.access(ctx, scope, true); err != nil {
		return label.Job{}, err
	}
	j, err := s.jobs.GetPrintJob(ctx, jobScope(scope), id)
	if err != nil {
		return label.Job{}, jobError(err)
	}
	result, err := s.jobs.UpdatePrintJob(ctx, ports.PrintJobUpdate{Scope: j.Scope, PrinterID: j.PrinterID, JobID: id, Now: s.labels.deps.Clock.Now(), Change: func(job *label.Job, p label.Printer) error { return job.Cancel(s.labels.deps.Clock.Now(), revision) }, Audit: func(before, after label.Job) (audit.Record, error) {
		return s.audit(scope, audit.ActionPrintJobCanceled, id)
	}})
	if err == nil && s.labels.deps.Observer != nil {
		s.labels.deps.Observer.Record(ctx, ports.Event{Name: ports.EventPrintJobCanceled})
	}
	return result, jobError(err)
}

type JobSelection struct {
	PrinterID                label.PrinterID
	ExpectedMediaFingerprint string
	Template                 label.TemplateSelection
	Copies                   int
	PreviewFingerprint       string
}
type CreateJobInput struct {
	Scope          LabelScope
	AssetID        asset.ID
	IdempotencyKey string
	Selection      JobSelection
	Predecessor    label.JobID
	Kind           label.JobKind
}

func requestFingerprint(input CreateJobInput) string {
	data, _ := json.Marshal(struct {
		AssetID     asset.ID
		Selection   JobSelection
		Predecessor label.JobID   `json:",omitempty"`
		Kind        label.JobKind `json:",omitempty"`
	}{input.AssetID, input.Selection, input.Predecessor, input.Kind})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func (s *JobService) Create(ctx context.Context, input CreateJobInput) (label.Job, bool, error) {
	if err := s.access(ctx, input.Scope, true); err != nil {
		return label.Job{}, false, err
	}
	if len(input.IdempotencyKey) < 1 || len(input.IdempotencyKey) > 200 || input.Selection.Copies < 1 || input.Selection.Copies > s.config.MaxCopies || s.config.ArtifactTTL <= 0 || s.config.MaxArtifactBytes <= 0 {
		return label.Job{}, false, apperrors.ErrInvalidInput
	}
	fingerprint := requestFingerprint(input)
	existing, prior, err := s.jobs.FindPrintJobRequest(ctx, jobScope(input.Scope), string(input.Scope.Principal.ID), input.IdempotencyKey)
	if err == nil {
		if prior != fingerprint {
			return label.Job{}, false, apperrors.ErrConflict
		}
		return existing, false, nil
	}
	if !errors.Is(err, ports.ErrPrintJobNotFound) {
		return label.Job{}, false, err
	}
	var item asset.Asset
	var content label.ContentSnapshot
	if input.Kind == label.JobPrinterTest {
		if input.AssetID != "" || input.Selection.Copies != 1 {
			return label.Job{}, false, apperrors.ErrInvalidInput
		}
		content = label.ContentSnapshot{Title: "Stuff Stash printer test", Reference: "PRINTER TEST", QRURL: "https://example.invalid/stuff-stash-printer-test"}
	} else {
		item, err = s.labels.asset(ctx, input.Scope, input.AssetID)
		if err != nil {
			return label.Job{}, false, err
		}
		if item.LifecycleState != asset.LifecycleStateActive {
			return label.Job{}, false, apperrors.ErrConflict
		}
	}
	printer, err := s.printers.GetPrinter(ctx, jobScope(input.Scope), input.Selection.PrinterID)
	if err != nil {
		return label.Job{}, false, jobError(err)
	}
	if printer.Retired || printer.MediaFingerprint != input.Selection.ExpectedMediaFingerprint {
		return label.Job{}, false, apperrors.ErrConflict
	}
	if input.Kind != label.JobPrinterTest {
		view, err := s.labels.Provision(ctx, input.Scope, input.AssetID)
		if err != nil {
			return label.Job{}, false, err
		}
		content = label.ContentSnapshot{Title: item.Title.String(), QRURL: view.URL, Reference: string(view.Label.ID)}
	}
	request := label.RenderRequest{Content: content, Template: input.Selection.Template, Media: printer.Media, Format: label.FormatPNG}
	selected, _ := json.Marshal(request)
	selectionHash := sha256.Sum256(selected)
	if input.Selection.PreviewFingerprint != "" && input.Selection.PreviewFingerprint != hex.EncodeToString(selectionHash[:]) {
		return label.Job{}, false, apperrors.ErrConflict
	}
	job, renderedContent, err := s.prepareJob(ctx, input.Scope, printer, input.Selection, request, string(input.AssetID), content.Reference, input.IdempotencyKey)
	if err != nil {
		return label.Job{}, false, err
	}
	if err = s.access(ctx, input.Scope, true); err != nil {
		return label.Job{}, false, err
	}
	if input.Kind != label.JobPrinterTest {
		current, err := s.labels.asset(ctx, input.Scope, input.AssetID)
		if err != nil {
			return label.Job{}, false, err
		}
		if current.LifecycleState != asset.LifecycleStateActive || current.Title != item.Title {
			return label.Job{}, false, apperrors.ErrConflict
		}
	}
	job.Predecessor = input.Predecessor
	if input.Kind == label.JobPrinterTest {
		job.Kind = label.JobPrinterTest
		job.LabelReference = ""
	}
	action, event := audit.ActionPrintJobQueued, ports.EventPrintJobQueued
	if job.Predecessor != "" {
		action, event = audit.ActionPrintJobReprinted, ports.EventPrintJobReprinted
	}
	record, err := s.audit(input.Scope, action, job.ID)
	if err != nil {
		return label.Job{}, false, err
	}
	result, created, err := s.jobs.CreatePrintJob(ctx, ports.PrintJobCreate{Job: job, Content: renderedContent, PrinterRevision: printer.Revision, RequestFingerprint: fingerprint, Audit: record})
	if err == nil && created && s.labels.deps.Observer != nil {
		s.labels.deps.Observer.Record(ctx, ports.Event{Name: event})
	}
	return result, created, jobError(err)
}

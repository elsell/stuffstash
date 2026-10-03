package printing

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	label "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var ErrLabelsUnavailable = errors.New("label service unavailable")

type LabelDependencies struct {
	Repository     ports.LabelRepository
	Renders        ports.LabelRenderRepository
	Assets         ports.AssetRepository
	Inventories    ports.InventoryRepository
	Tenants        ports.TenantRepository
	Authorizer     ports.Authorizer
	Audit          ports.AuditRepository
	IDs            ports.IDGenerator
	Clock          ports.Clock
	Observer       ports.Observer
	Renderer       ports.LabelRenderer
	Templates      ports.LabelTemplateCatalog
	BaseURL        string
	RenderTTL      time.Duration
	MaxRenderBytes int
}
type LabelService struct{ deps LabelDependencies }

func NewLabelService(deps LabelDependencies) *LabelService { return &LabelService{deps: deps} }

type LabelScope struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	RequestID   string
}
type LabelView = label.LabelView

func ValidateLabelBaseURL(value string) error {
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" || strings.ContainsAny(u.Path, "\\") {
		return apperrors.ErrInvalidInput
	}
	if u.RawPath != "" || strings.Contains(u.Path, "//") {
		return apperrors.ErrInvalidInput
	}
	for _, part := range strings.Split(u.Path, "/") {
		if part == "." || part == ".." {
			return apperrors.ErrInvalidInput
		}
	}
	return nil
}
func (s *LabelService) BootstrapInstance(ctx context.Context) (label.InstanceID, error) {
	if s == nil || s.deps.Repository == nil || s.deps.IDs == nil {
		return "", ErrLabelsUnavailable
	}
	id := label.InstanceID(s.deps.IDs.NewID())
	if !label.ValidOpaqueID(string(id)) {
		return "", apperrors.ErrInvalidInput
	}
	return s.deps.Repository.BootstrapLabelInstance(ctx, id)
}
func (s *LabelService) Instance(ctx context.Context) (label.InstanceID, error) {
	if s == nil || s.deps.Repository == nil {
		return "", ErrLabelsUnavailable
	}
	id, found, err := s.deps.Repository.LabelInstance(ctx)
	if err != nil {
		return "", err
	}
	if !found {
		return "", ErrLabelsUnavailable
	}
	return id, nil
}
func (s *LabelService) access(ctx context.Context, scope LabelScope) error {
	if s == nil || s.deps.Authorizer == nil || s.deps.Repository == nil || ValidateLabelBaseURL(s.deps.BaseURL) != nil {
		return ErrLabelsUnavailable
	}
	if scope.Principal.ID == "" {
		return apperrors.ErrUnauthenticated
	}
	if scope.TenantID == "" || scope.InventoryID == "" {
		return apperrors.ErrInvalidInput
	}
	if err := s.deps.Authorizer.CheckInventory(ctx, scope.Principal, ports.InventoryPermissionView, scope.InventoryID); err != nil {
		return err
	}
	inventory, found, err := s.deps.Inventories.InventoryByID(ctx, scope.TenantID, scope.InventoryID)
	if err != nil {
		return err
	}
	if !found || !inventory.IsActive() {
		return apperrors.ErrNotFound
	}
	tenant, found, err := s.deps.Tenants.TenantByID(ctx, scope.TenantID)
	if err != nil {
		return err
	}
	if !found || !tenant.IsActive() {
		return apperrors.ErrNotFound
	}
	return nil
}
func (s *LabelService) asset(ctx context.Context, scope LabelScope, id asset.ID) (asset.Asset, error) {
	if err := s.access(ctx, scope); err != nil {
		return asset.Asset{}, err
	}
	item, found, err := s.deps.Assets.AssetByID(ctx, scope.TenantID, scope.InventoryID, id)
	if err != nil {
		return item, err
	}
	if !found {
		return item, apperrors.ErrNotFound
	}
	return item, nil
}
func (s *LabelService) auditInput(scope LabelScope, action audit.Action, target string) appsupport.AuditRecordInput {
	targetType := audit.TargetAsset
	if target == "" {
		targetType = audit.TargetInventory
		target = scope.InventoryID.String()
	}
	return appsupport.AuditRecordInput{Principal: scope.Principal, TenantID: scope.TenantID, InventoryID: scope.InventoryID, RequestID: scope.RequestID, Source: audit.SourceAPI, Action: action, TargetType: targetType, TargetID: target}
}
func (s *LabelService) readAudit(ctx context.Context, scope LabelScope, action audit.Action, target string) error {
	return appsupport.SaveReadAuditRecord(ctx, s.deps.Audit, s.deps.IDs, s.deps.Clock, s.auditInput(scope, action, target))
}
func (s *LabelService) view(record label.Label, item asset.Asset) LabelView {
	return LabelView{Label: record, URL: strings.TrimRight(s.deps.BaseURL, "/") + "/l/v1/" + string(record.InstanceID) + "/" + string(record.ID), LifecycleState: item.LifecycleState.String()}
}
func (s *LabelService) Provision(ctx context.Context, scope LabelScope, assetID asset.ID) (LabelView, error) {
	item, err := s.asset(ctx, scope, assetID)
	if err != nil {
		return LabelView{}, err
	}
	instance, err := s.Instance(ctx)
	if err != nil {
		return LabelView{}, err
	}
	record := label.Label{InstanceID: instance, ID: label.LabelID(s.deps.IDs.NewID()), TenantID: scope.TenantID.String(), InventoryID: scope.InventoryID.String(), AssetID: assetID.String(), CreatedAt: s.deps.Clock.Now()}
	auditRecord, err := appsupport.NewAuditRecord(s.deps.IDs, s.deps.Clock, s.auditInput(scope, audit.ActionLabelProvisioned, assetID.String()))
	if err != nil {
		return LabelView{}, err
	}
	stored, created, err := s.deps.Repository.ProvisionLabel(ctx, record, auditRecord)
	if err != nil {
		return LabelView{}, err
	}
	if stored.Tombstoned {
		return LabelView{}, apperrors.ErrNotFound
	}
	if created && s.deps.Observer != nil {
		s.deps.Observer.Record(ctx, ports.Event{Name: ports.EventLabelProvisioned})
	}
	return s.view(stored, item), nil
}
func (s *LabelService) Get(ctx context.Context, scope LabelScope, assetID asset.ID) (LabelView, error) {
	item, err := s.asset(ctx, scope, assetID)
	if err != nil {
		return LabelView{}, err
	}
	if _, err = s.Instance(ctx); err != nil {
		return LabelView{}, err
	}
	record, found, err := s.deps.Repository.LabelForAsset(ctx, scope.TenantID, scope.InventoryID, assetID)
	if err != nil {
		return LabelView{}, err
	}
	if !found || record.Tombstoned {
		return LabelView{}, apperrors.ErrNotFound
	}
	if err := s.readAudit(ctx, scope, audit.ActionLabelViewed, assetID.String()); err != nil {
		return LabelView{}, err
	}
	return s.view(record, item), nil
}
func (s *LabelService) Resolve(ctx context.Context, principal identity.Principal, instanceID, labelID, requestID string) (LabelView, error) {
	if principal.ID == "" {
		return LabelView{}, apperrors.ErrUnauthenticated
	}
	if !label.ValidOpaqueID(instanceID) || !label.ValidOpaqueID(labelID) {
		return LabelView{}, apperrors.ErrNotFound
	}
	instance, err := s.Instance(ctx)
	if err != nil {
		return LabelView{}, err
	}
	if string(instance) != instanceID {
		return LabelView{}, apperrors.ErrNotFound
	}
	record, found, err := s.deps.Repository.LookupLabel(ctx, instance, label.LabelID(labelID))
	if err != nil {
		return LabelView{}, err
	}
	if !found || record.Tombstoned {
		return LabelView{}, apperrors.ErrNotFound
	}
	scope := LabelScope{Principal: principal, TenantID: tenant.ID(record.TenantID), InventoryID: inventory.InventoryID(record.InventoryID), RequestID: requestID}
	item, err := s.asset(ctx, scope, asset.ID(record.AssetID))
	if errors.Is(err, ports.ErrForbidden) {
		return LabelView{}, apperrors.ErrNotFound
	}
	if err != nil {
		return LabelView{}, err
	}
	if err := s.readAudit(ctx, scope, audit.ActionLabelResolved, record.AssetID); err != nil {
		return LabelView{}, err
	}
	return s.view(record, item), nil
}

package expiration

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	searchapp "github.com/stuffstash/stuff-stash/internal/app/search"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const workspaceBatchSize = 256
const workspaceMaxBatches = 1000

type WorkspaceInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Source      audit.Source
	RequestID   string
	Mode        WorkspaceMode
	Filter      WorkspaceFilter
	Limit       int
	Cursor      string
}
type WorkspaceDependencies struct {
	Checkouts   ports.AssetCheckoutRepository
	Assets      ports.AssetRepository
	Types       ports.CustomAssetTypeRepository
	Tags        ports.AssetTagRepository
	Inventories ports.InventoryRepository
	Authorizer  ports.Authorizer
	Audit       ports.AuditRepository
	IDs         ports.IDGenerator
	Clock       ports.Clock
	Observer    ports.Observer
}
type WorkspaceResult struct {
	WorkspacePage
	Records    []ports.AssetSearchResult
	Limit      int
	NextCursor *string
	Timezone   string
}
type WorkspaceService struct{ deps WorkspaceDependencies }

func NewWorkspaceService(deps WorkspaceDependencies) WorkspaceService {
	return WorkspaceService{deps: deps}
}

func (s WorkspaceService) List(ctx context.Context, input WorkspaceInput, settings notification.Settings) (WorkspaceResult, error) {
	if input.Mode == "" {
		input.Mode = WorkspaceAll
	}
	if !input.Mode.Valid() || input.Filter.Validate() != nil || settings.Validate() != nil || s.deps.Clock == nil || s.deps.Assets == nil || s.deps.Checkouts == nil || s.deps.Types == nil || s.deps.Tags == nil || s.deps.Inventories == nil || s.deps.Authorizer == nil {
		return WorkspaceResult{}, apperrors.ErrInvalidInput
	}
	if input.Principal.ID == "" || input.TenantID == "" || input.InventoryID == "" {
		return WorkspaceResult{}, apperrors.ErrInvalidInput
	}
	if err := s.deps.Authorizer.CheckInventory(ctx, input.Principal, ports.InventoryPermissionView, input.InventoryID); err != nil {
		return WorkspaceResult{}, err
	}
	inv, found, err := s.deps.Inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return WorkspaceResult{}, err
	}
	if !found || !inv.IsActive() {
		return WorkspaceResult{}, apperrors.ErrNotFound
	}
	if input.Limit == 0 {
		input.Limit = 50
	}
	if input.Limit < 1 || input.Limit > 100 {
		return WorkspaceResult{}, apperrors.ErrInvalidInput
	}
	input.Filter.Text = strings.ToLower(strings.TrimSpace(input.Filter.Text))
	input.Filter.TagIDs = slices.Clone(input.Filter.TagIDs)
	slices.Sort(input.Filter.TagIDs)
	input.Filter.TagIDs = slices.Compact(input.Filter.TagIDs)
	now := s.deps.Clock.Now()
	zone, _ := time.LoadLocation(settings.Timezone)
	scopeJSON, _ := json.Marshal(struct {
		Principal string
		Tenant    string
		Inventory string
		Mode      WorkspaceMode
		Filter    WorkspaceFilter
		Settings  notification.Settings
		Day       string
	}{string(input.Principal.ID), string(input.TenantID), string(input.InventoryID), input.Mode, input.Filter, settings, now.In(zone).Format("2006-01-02")})
	scope := string(scopeJSON)
	after, err := DecodeContinuation(input.Cursor, scope, Query{Status: QueryAll})
	if err != nil {
		return WorkspaceResult{}, apperrors.ErrInvalidInput
	}
	selection := NewWorkspaceSelection(input.Mode, input.Limit, after)
	var afterID asset.ID
	complete := false
	for batch := 0; batch < workspaceMaxBatches; batch++ {
		if err := ctx.Err(); err != nil {
			return WorkspaceResult{}, err
		}
		items, err := s.deps.Assets.ListAssetsByInventory(ctx, input.TenantID, input.InventoryID, ports.AssetListPageRequest{AfterAssetID: afterID, Limit: workspaceBatchSize, LifecycleFilter: ports.AssetLifecycleFilterActive, OnlyDated: true})
		if err != nil {
			return WorkspaceResult{}, err
		}
		if len(items) > workspaceBatchSize {
			return WorkspaceResult{}, apperrors.ErrInvalidInput
		}
		ids := make([]asset.ID, 0, len(items))
		for _, item := range items {
			if item.ID <= afterID || item.TenantID != asset.TenantID(input.TenantID) || item.InventoryID != asset.InventoryID(input.InventoryID) || item.LifecycleState != asset.LifecycleStateActive {
				return WorkspaceResult{}, apperrors.ErrInvalidInput
			}
			afterID = item.ID
			ids = append(ids, item.ID)
		}
		tags, err := s.deps.Tags.AssetTagsByAssets(ctx, input.TenantID, input.InventoryID, ids)
		if err != nil {
			return WorkspaceResult{}, err
		}
		descriptions, err := DescribeItems(ctx, items, settings, s.deps.Types, now)
		if err != nil {
			return WorkspaceResult{}, err
		}
		current := map[asset.ID]asset.Checkout{}
		if input.Filter.CheckoutState != "" && input.Filter.CheckoutState != ports.AssetCheckoutStateFilterAny {
			current, err = s.deps.Checkouts.CurrentAssetCheckouts(ctx, input.TenantID, input.InventoryID, ids)
			if err != nil {
				return WorkspaceResult{}, err
			}
		}
		pathsByID := make(map[asset.ID][]ports.AssetSearchAncestor)
		if input.Filter.LocationID != "" {
			batchRecords := make([]ports.AssetSearchResult, 0, len(items))
			for _, item := range items {
				batchRecords = append(batchRecords, ports.AssetSearchResult{TenantID: input.TenantID, Inventory: inv, Asset: item})
			}
			paths, err := searchapp.WithAncestorPaths(ctx, s.deps.Assets, batchRecords)
			if err != nil {
				return WorkspaceResult{}, err
			}
			for _, record := range paths {
				pathsByID[record.Asset.ID] = record.AncestorPath
			}
		}
		for _, item := range items {
			_, checkedOut := current[item.ID]
			candidate := WorkspaceItem{CheckedOut: checkedOut, Asset: item, Expiration: descriptions[ports.AttachmentAssetReference{InventoryID: input.InventoryID, AssetID: item.ID}]}
			for _, tag := range tags[item.ID] {
				candidate.TagIDs = append(candidate.TagIDs, tag.ID.String())
			}
			for _, parent := range pathsByID[item.ID] {
				candidate.AncestorIDs = append(candidate.AncestorIDs, parent.ID.String())
			}

			if input.Filter.Matches(candidate) {
				selection.Add(candidate)
			}
		}
		if len(items) < workspaceBatchSize {
			complete = true
			break
		}
	}
	if !complete {
		return WorkspaceResult{}, errors.New("expiration query exceeded its scan budget")
	}
	page := selection.Result()
	records := make([]ports.AssetSearchResult, 0, len(page.Items))
	ids := make([]asset.ID, 0, len(page.Items))
	for _, item := range page.Items {
		records = append(records, ports.AssetSearchResult{TenantID: input.TenantID, Inventory: inv, Asset: item.Asset})
		ids = append(ids, item.Asset.ID)
	}
	records, err = searchapp.WithAncestorPaths(ctx, s.deps.Assets, records)
	if err != nil {
		return WorkspaceResult{}, err
	}
	tags, err := s.deps.Tags.AssetTagsByAssets(ctx, input.TenantID, input.InventoryID, ids)
	if err != nil {
		return WorkspaceResult{}, err
	}
	current, err := s.deps.Checkouts.CurrentAssetCheckouts(ctx, input.TenantID, input.InventoryID, ids)
	if err != nil {
		return WorkspaceResult{}, err
	}
	for index := range records {
		if checkout, ok := current[records[index].Asset.ID]; ok {
			records[index].CurrentCheckout = &checkout
		}
		records[index].AssignedTags = tags[records[index].Asset.ID]
	}
	if err := s.deps.Authorizer.CheckInventory(ctx, input.Principal, ports.InventoryPermissionView, input.InventoryID); err != nil {
		return WorkspaceResult{}, err
	}
	if err := appsupport.SaveReadAuditRecord(ctx, s.deps.Audit, s.deps.IDs, s.deps.Clock, appsupport.AuditRecordInput{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, Source: input.Source, RequestID: input.RequestID, Action: audit.ActionAssetListed, TargetType: audit.TargetInventory, TargetID: input.InventoryID.String(), Metadata: map[string]string{"surface": "expiration", "mode": string(input.Mode), "count": strconv.Itoa(len(page.Items))}}); err != nil {
		return WorkspaceResult{}, err
	}
	if s.deps.Observer != nil {
		s.deps.Observer.Record(ctx, ports.Event{Name: ports.EventAssetsListed, Message: "expiration assets listed", Fields: map[string]string{"tenant_id": input.TenantID.String(), "inventory_id": input.InventoryID.String(), "mode": string(input.Mode)}})
	}
	result := WorkspaceResult{WorkspacePage: page, Records: records, Limit: input.Limit, Timezone: settings.Timezone}
	if page.HasMore {
		token := EncodeContinuation(WorkspacePosition(page.Items[len(page.Items)-1], input.Mode), scope, Query{Status: QueryAll})
		result.NextCursor = &token
	}
	return result, nil
}

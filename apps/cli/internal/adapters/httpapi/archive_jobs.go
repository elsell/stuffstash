package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func archiveJob(v generated.ArchiveJob) ports.ArchiveJob {
	return ports.ArchiveJob{ID: v.Id, CreatedAt: v.CreatedAt, ExpiresAt: v.ExpiresAt, DestinationInventoryID: v.DestinationInventoryId, InventoryID: v.InventoryId, Failure: v.Failure, Kind: v.Kind, OtherFiles: v.OtherFiles, Photos: v.Photos, Phase: v.Phase, State: v.State}
}
func archiveJobResult(response *http.Response, err error) (ports.Result[ports.ArchiveJob], error) {
	r, err := read[generated.SuccessEnvelopeArchiveJob](response, err)
	if err != nil {
		return ports.Result[ports.ArchiveJob]{}, err
	}
	return ports.Result[ports.ArchiveJob]{Data: archiveJob(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func archiveInventory(s ports.Scope) *string {
	if s.Inventory == "" {
		return nil
	}
	return &s.Inventory
}
func (c *Client) ArchiveJobs(ctx context.Context, s ports.Scope, p ports.Page) (ports.Result[[]ports.ArchiveJob], error) {
	r, err := read[generated.SuccessEnvelopeListArchiveJob](c.sdk.GetTenantsByTenantIdArchiveJobs(ctx, s.Tenant, &generated.GetTenantsByTenantIdArchiveJobsParams{InventoryId: archiveInventory(s), After: &p.Cursor, Limit: &p.Limit}))
	if err != nil {
		return ports.Result[[]ports.ArchiveJob]{}, err
	}
	var jobs []ports.ArchiveJob
	if r.Data.GetOrEmpty() != nil {
		jobs = make([]ports.ArchiveJob, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		jobs = append(jobs, archiveJob(v))
	}
	return ports.Result[[]ports.ArchiveJob]{Data: jobs, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) ArchiveJob(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.ArchiveJob], error) {
	return archiveJobResult(c.sdk.GetTenantsByTenantIdArchiveJobsByJobId(ctx, s.Tenant, id, &generated.GetTenantsByTenantIdArchiveJobsByJobIdParams{InventoryId: archiveInventory(s)}))
}
func (c *Client) RetryArchiveJob(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.ArchiveJob], error) {
	return archiveJobResult(c.sdk.PostTenantsByTenantIdArchiveJobsByJobIdRetry(ctx, s.Tenant, id, &generated.PostTenantsByTenantIdArchiveJobsByJobIdRetryParams{InventoryId: archiveInventory(s)}))
}
func (c *Client) DeleteArchiveJob(ctx context.Context, s ports.Scope, id string) error {
	return noContent(c.sdk.DeleteTenantsByTenantIdArchiveJobsByJobId(ctx, s.Tenant, id, &generated.DeleteTenantsByTenantIdArchiveJobsByJobIdParams{InventoryId: archiveInventory(s)}))
}
func (c *Client) ArchivePreview(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.ArchivePreview], error) {
	r, err := read[generated.SuccessEnvelopeArchivePreview](c.sdk.GetTenantsByTenantIdArchiveJobsByJobIdPreview(ctx, s.Tenant, id, &generated.GetTenantsByTenantIdArchiveJobsByJobIdPreviewParams{InventoryId: archiveInventory(s)}))
	if err != nil {
		return ports.Result[ports.ArchivePreview]{}, err
	}
	v := r.Data
	var keys []ports.ArchiveKeyRemapping
	if v.KeyRemappings.GetOrEmpty() != nil {
		keys = make([]ports.ArchiveKeyRemapping, 0, len(v.KeyRemappings.GetOrEmpty()))
	}
	for _, k := range v.KeyRemappings.GetOrEmpty() {
		keys = append(keys, ports.ArchiveKeyRemapping{DestinationKey: k.DestinationKey, SourceKey: k.SourceKey, Family: k.Family})
	}
	return ports.Result[ports.ArchivePreview]{Data: ports.ArchivePreview{Assets: v.Assets, CustomAssetTypes: v.CustomAssetTypes, CustomFields: v.CustomFields, InventoryName: v.InventoryName, KeyRemappings: keys, OmittedAttachments: v.OmittedAttachments, OtherFiles: v.OtherFiles, Photos: v.Photos, Tags: v.Tags}, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}

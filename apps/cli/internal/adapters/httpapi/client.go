package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Client struct{ sdk *generated.Client }

func New(server, token string, httpClient *http.Client) (*Client, error) {
	// API credentials must never follow a server redirect to another origin.
	safe := *httpClient
	safe.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	sdk, err := generated.NewClient(server, generated.WithHTTPClient(&safe), generated.WithRequestEditorFn(func(_ context.Context, r *http.Request) error {
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		return nil
	}))
	if err != nil {
		return nil, err
	}
	return &Client{sdk: sdk}, nil
}
func read[T any](response *http.Response, err error) (T, error) {
	var zero T
	if err != nil {
		return zero, ports.Failure("network", "could not reach Stuff Stash")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		category, message := "api", "Stuff Stash rejected the request"
		switch response.StatusCode {
		case 401:
			category, message = "authentication", "session expired or invalid; log in again"
		case 403:
			category, message = "forbidden", "you do not have permission for this action"
		case 404:
			category, message = "not_found", "resource not found"
		case 409:
			category, message = "conflict", "the resource changed or the operation conflicts; refresh and retry"
		case 400, 422:
			category, message = "validation", "request is invalid; check the supplied values"
		case 503:
			category, message = "unavailable", "service or login method unavailable"
		}
		return zero, ports.Failure(category, message)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 16<<20)).Decode(&zero); err != nil {
		return zero, ports.Failure("protocol", "invalid Stuff Stash response")
	}
	return zero, nil
}
func key(value string) generated.RequestEditorFn {
	return func(_ context.Context, r *http.Request) error { r.Header.Set("Idempotency-Key", value); return nil }
}
func page(meta generated.Meta) *ports.Pagination {
	if meta.Pagination == nil {
		return nil
	}
	p := meta.Pagination
	return &ports.Pagination{Limit: int(p.Limit), NextCursor: optionalCursor(p.NextCursor.GetOrEmpty()), HasMore: p.HasMore}
}
func asset(a generated.AssetResponse) ports.Asset {
	p := ""
	if a.ParentAssetId != nil {
		p = *a.ParentAssetId
	}
	printJobID := ""
	if a.PrintJobId != nil {
		printJobID = *a.PrintJobId
	}
	return ports.Asset{PrintJobID: printJobID, ID: a.Id, Title: a.Title, Kind: a.Kind, Parent: p, Lifecycle: a.LifecycleState}
}
func assetResult(r generated.SuccessEnvelopeAssetResponse, err error) (ports.Result[ports.Asset], error) {
	if err != nil {
		return ports.Result[ports.Asset]{}, err
	}
	return ports.Result[ports.Asset]{Data: asset(r.Data)}, nil
}
func (c *Client) Inventories(ctx context.Context, s ports.Scope, p ports.Page) (ports.Result[[]ports.Inventory], error) {
	r, err := read[generated.SuccessEnvelopeListInventoryResponse](c.sdk.GetTenantsByTenantIdInventories(ctx, s.Tenant, &generated.GetTenantsByTenantIdInventoriesParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.Inventory]{}, err
	}
	items := make([]ports.Inventory, 0)
	for _, v := range r.Data.GetOrEmpty() {
		items = append(items, inventory(v))
	}
	return ports.Result[[]ports.Inventory]{Data: items, Pagination: page(r.Meta)}, nil
}
func (c *Client) Assets(ctx context.Context, s ports.Scope, p ports.Page) (ports.Result[[]ports.Asset], error) {
	r, err := read[generated.SuccessEnvelopeListAssetResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAssets(ctx, s.Tenant, s.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdAssetsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.Asset]{}, err
	}
	items := make([]ports.Asset, 0)
	for _, v := range r.Data.GetOrEmpty() {
		items = append(items, asset(v))
	}
	return ports.Result[[]ports.Asset]{Data: items, Pagination: page(r.Meta)}, nil
}
func (c *Client) Asset(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.Asset], error) {
	return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetId(ctx, s.Tenant, s.Inventory, id, nil)))
}

func optionalCursor(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

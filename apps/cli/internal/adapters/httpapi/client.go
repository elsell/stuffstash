package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Client struct{ sdk *generated.Client }

type Options struct{ RequestID string }

func New(server, token string, httpClient *http.Client, options ...Options) (*Client, error) {
	requestID := ""
	if len(options) > 0 {
		requestID = options[0].RequestID
	}
	if strings.IndexFunc(requestID, func(r rune) bool { return r < 32 || r > 126 }) >= 0 {
		return nil, ports.Failure("usage", "The request ID contains invalid characters. Use printable ASCII characters.")
	}

	// API credentials must never follow a server redirect to another origin.
	safe := *httpClient
	safe.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	sdk, err := generated.NewClient(server, generated.WithHTTPClient(&safe), generated.WithRequestEditorFn(func(_ context.Context, r *http.Request) error {
		if requestID != "" {
			r.Header.Set("X-Request-ID", requestID)
		}
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
	decoder := json.NewDecoder(io.LimitReader(response.Body, 16<<20))
	decoder.UseNumber()
	if err := decoder.Decode(&zero); err != nil {
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
	var cursor *string
	if value, err := p.NextCursor.Get(); err == nil {
		cursor = &value
	}
	return &ports.Pagination{Limit: p.Limit, NextCursor: cursor, HasMore: p.HasMore}
}
func assetResult(r generated.SuccessEnvelopeAssetResponse, err error) (ports.Result[ports.Asset], error) {
	if err != nil {
		return ports.Result[ports.Asset]{}, err
	}
	return ports.Result[ports.Asset]{Data: asset(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) Inventories(ctx context.Context, s ports.Scope, p ports.Page) (ports.Result[[]ports.Inventory], error) {
	r, err := read[generated.SuccessEnvelopeListInventoryResponse](c.sdk.GetTenantsByTenantIdInventories(ctx, s.Tenant, &generated.GetTenantsByTenantIdInventoriesParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.Inventory]{}, err
	}
	var items []ports.Inventory
	if r.Data.GetOrEmpty() != nil {
		items = make([]ports.Inventory, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		items = append(items, inventory(v))
	}
	return ports.Result[[]ports.Inventory]{Data: items, Pagination: page(r.Meta), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) Assets(ctx context.Context, s ports.Scope, query ports.AssetQuery) (ports.Result[[]ports.Asset], error) {
	r, err := read[assetListResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAssets(ctx, s.Tenant, s.Inventory, assetQuery(query)))
	if err != nil {
		return ports.Result[[]ports.Asset]{}, err
	}
	var items []ports.Asset
	if r.Data != nil {
		items = make([]ports.Asset, 0, len(r.Data))
	}
	for _, v := range r.Data {
		items = append(items, asset(v))
	}
	return ports.Result[[]ports.Asset]{Data: items, Pagination: page(r.Meta), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
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

func metadata(meta generated.Meta) *ports.Metadata {
	return &ports.Metadata{RequestID: meta.RequestId, TenantID: meta.TenantId, Pagination: page(meta)}
}

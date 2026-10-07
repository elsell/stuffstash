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

type Client struct {
	sdk      *generated.Client
	receipts ports.ProtocolReceipts
}

type Options struct {
	RequestID string
	Receipts  ports.ProtocolReceipts
}

func New(server, token string, httpClient *http.Client, options ...Options) (*Client, error) {
	requestID := ""
	if len(options) > 0 {
		requestID = options[0].RequestID
	}
	if strings.IndexFunc(requestID, func(r rune) bool { return r < 32 || r > 126 }) >= 0 {
		return nil, ports.Failure("usage", "The request ID contains characters outside printable ASCII. Use printable ASCII characters.")
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
	var receipts ports.ProtocolReceipts
	if len(options) > 0 {
		receipts = options[0].Receipts
	}
	return &Client{sdk: sdk, receipts: receipts}, nil
}
func read[T any](response *http.Response, err error) (T, error) {
	var zero T
	if err != nil {
		return zero, ports.Failure("network", "The CLI cannot connect to Stuff Stash. Make sure that the server address is correct. Examine the current state before you make the change again.")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		category, message := "api", "Stuff Stash did not return a successful result. Examine the current state before you make the change again."
		switch response.StatusCode {
		case 401:
			category, message = "authentication", "The server did not accept your session. Run stuffstash login and try again."
		case 403:
			category, message = "forbidden", "You do not have permission for this action. Ask a household or inventory owner to examine your access."
		case 404:
			category, message = "not_found", "The server did not find the resource. Examine its ID and the selected household and inventory."
		case 409:
			category, message = "conflict", "The request conflicts with the current state. Read the resource again and review your change before you submit it."
		case 400, 422:
			category, message = "validation", "The server did not accept the request. Examine the input values against the command help and API requirements."
		case 429:
			category, message = "api", "Too many requests were sent. Wait before you send another request. Examine the current state before you make the change again."
		case 503:
			category, message = "unavailable", "The service is not available. Wait for it to recover. Examine the current state before you make the change again."
		}
		return zero, ports.Failure(category, message)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 16<<20))
	decoder.UseNumber()
	if err := decoder.Decode(&zero); err != nil {
		return zero, ports.Failure("protocol", "The server response is not correct. Examine the current state before you make the change again. Contact the server operator if this continues.")
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

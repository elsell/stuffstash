package httpserver

import (
	"net/http/httptest"
	"testing"
)

type assetResponse struct {
	ID                string                 `json:"id"`
	TenantID          string                 `json:"tenantId"`
	InventoryID       string                 `json:"inventoryId"`
	ParentAssetID     string                 `json:"parentAssetId,omitempty"`
	CustomAssetTypeID string                 `json:"customAssetTypeId,omitempty"`
	Kind              string                 `json:"kind"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	CustomFields      map[string]interface{} `json:"customFields"`
	Tags              []assetTagSummary      `json:"tags"`
	LifecycleState    string                 `json:"lifecycleState"`
	CreatedAt         string                 `json:"createdAt"`
	UpdatedAt         string                 `json:"updatedAt"`
	PrimaryPhoto      *assetPrimaryPhoto     `json:"primaryPhoto,omitempty"`
	CurrentCheckout   *currentCheckout       `json:"currentCheckout,omitempty"`
}

type assetTagSummary struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
	Color       string `json:"color,omitempty"`
}

type currentCheckout struct {
	ID                      string            `json:"id"`
	State                   string            `json:"state"`
	CheckedOutAt            string            `json:"checkedOutAt"`
	CheckedOutByPrincipalID string            `json:"checkedOutByPrincipalId"`
	CheckedOutByPrincipal   *principalSummary `json:"checkedOutByPrincipal,omitempty"`
}

type principalSummary struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
}

type assetPrimaryPhoto struct {
	ID          string               `json:"id"`
	FileName    string               `json:"fileName"`
	ContentType string               `json:"contentType"`
	SizeBytes   int64                `json:"sizeBytes"`
	Thumbnails  assetPhotoThumbnails `json:"thumbnails"`
}

type assetPhotoThumbnails struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

type assetBody struct {
	Data assetResponse `json:"data"`
	Meta responseMeta  `json:"meta"`
}

type assetListBody struct {
	Data []assetResponse `json:"data"`
	Meta responseMeta    `json:"meta"`
}

type assetCheckoutResponse struct {
	ID                      string `json:"id"`
	TenantID                string `json:"tenantId"`
	InventoryID             string `json:"inventoryId"`
	AssetID                 string `json:"assetId"`
	State                   string `json:"state"`
	CheckedOutAt            string `json:"checkedOutAt"`
	CheckedOutByPrincipalID string `json:"checkedOutByPrincipalId"`
	CheckoutDetails         string `json:"checkoutDetails,omitempty"`
	ReturnedAt              string `json:"returnedAt,omitempty"`
	ReturnedByPrincipalID   string `json:"returnedByPrincipalId,omitempty"`
	ReturnDetails           string `json:"returnDetails,omitempty"`
	CreatedAt               string `json:"createdAt"`
	UpdatedAt               string `json:"updatedAt"`
}

type assetCheckoutBody struct {
	Data assetCheckoutResponse `json:"data"`
	Meta responseMeta          `json:"meta"`
}

type assetCheckoutListBody struct {
	Data []assetCheckoutResponse `json:"data"`
	Meta responseMeta            `json:"meta"`
}

type checkedOutAssetResponse struct {
	Asset    assetResponse   `json:"asset"`
	Checkout currentCheckout `json:"checkout"`
}

type checkedOutAssetListBody struct {
	Data []checkedOutAssetResponse `json:"data"`
	Meta responseMeta              `json:"meta"`
}

func decodeAsset(t *testing.T, response *httptest.ResponseRecorder) assetBody {
	t.Helper()

	var body assetBody
	decodeBody(t, response, &body)
	return body
}

func decodeAssetList(t *testing.T, response *httptest.ResponseRecorder) assetListBody {
	t.Helper()

	var body assetListBody
	decodeBody(t, response, &body)
	return body
}

func decodeAssetCheckout(t *testing.T, response *httptest.ResponseRecorder) assetCheckoutBody {
	t.Helper()

	var body assetCheckoutBody
	decodeBody(t, response, &body)
	return body
}

func decodeAssetCheckoutList(t *testing.T, response *httptest.ResponseRecorder) assetCheckoutListBody {
	t.Helper()

	var body assetCheckoutListBody
	decodeBody(t, response, &body)
	return body
}

func decodeCheckedOutAssetList(t *testing.T, response *httptest.ResponseRecorder) checkedOutAssetListBody {
	t.Helper()

	var body checkedOutAssetListBody
	decodeBody(t, response, &body)
	return body
}

func assetListContainsID(items []assetResponse, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

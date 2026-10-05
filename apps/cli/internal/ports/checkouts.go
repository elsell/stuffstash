package ports

type CheckoutAction string

const (
	CheckOut            CheckoutAction = "checkout"
	Return              CheckoutAction = "return"
	UpdateReturnDetails CheckoutAction = "return-details"
)

type Checkout struct {
	ID                      string  `json:"id"`
	TenantID                string  `json:"tenantId"`
	InventoryID             string  `json:"inventoryId"`
	AssetID                 string  `json:"assetId"`
	CheckedOutAt            string  `json:"checkedOutAt"`
	CheckedOutByPrincipalID string  `json:"checkedOutByPrincipalId"`
	CheckoutDetails         *string `json:"checkoutDetails,omitempty"`
	ReturnDetails           *string `json:"returnDetails,omitempty"`
	ReturnedAt              *string `json:"returnedAt,omitempty"`
	ReturnedByPrincipalID   *string `json:"returnedByPrincipalId,omitempty"`
	State                   string  `json:"state"`
	CreatedAt               string  `json:"createdAt"`
	UpdatedAt               string  `json:"updatedAt"`
	UndoableOperationID     *string `json:"undoableOperationId,omitempty"`
}

type CheckedOutAsset struct {
	Asset    Asset           `json:"asset"`
	Checkout CurrentCheckout `json:"checkout"`
}

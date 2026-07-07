package asset

import (
	"strings"
	"time"
	"unicode/utf8"
)

type CheckoutID string

func NewCheckoutID(value string) (CheckoutID, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	return CheckoutID(value), true
}

func (id CheckoutID) String() string {
	return string(id)
}

type CheckoutState string

const (
	CheckoutStateOpen     CheckoutState = "open"
	CheckoutStateReturned CheckoutState = "returned"
	CheckoutStateUndone   CheckoutState = "undone"
)

func NewCheckoutState(value string) (CheckoutState, bool) {
	switch CheckoutState(strings.TrimSpace(value)) {
	case CheckoutStateOpen:
		return CheckoutStateOpen, true
	case CheckoutStateReturned:
		return CheckoutStateReturned, true
	case CheckoutStateUndone:
		return CheckoutStateUndone, true
	default:
		return "", false
	}
}

func (s CheckoutState) String() string {
	return string(s)
}

type CheckoutDetails string

const MaxCheckoutDetailsLength = 1000

func NewCheckoutDetails(value string) (CheckoutDetails, bool) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > MaxCheckoutDetailsLength {
		return "", false
	}
	return CheckoutDetails(value), true
}

func (d CheckoutDetails) String() string {
	return string(d)
}

func (d CheckoutDetails) IsEmpty() bool {
	return d.String() == ""
}

type Checkout struct {
	ID                    CheckoutID
	TenantID              TenantID
	InventoryID           InventoryID
	AssetID               ID
	State                 CheckoutState
	CheckedOutAt          time.Time
	CheckedOutByPrincipal string
	CheckoutDetails       CheckoutDetails
	ReturnedAt            time.Time
	ReturnedByPrincipal   string
	ReturnDetails         CheckoutDetails
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (c Checkout) IsOpen() bool {
	return c.State == CheckoutStateOpen
}

func CheckoutsEquivalentForStaleCheck(left Checkout, right Checkout) bool {
	return left.ID == right.ID &&
		left.TenantID == right.TenantID &&
		left.InventoryID == right.InventoryID &&
		left.AssetID == right.AssetID &&
		left.State == right.State &&
		left.CheckedOutAt.Equal(right.CheckedOutAt) &&
		left.CheckedOutByPrincipal == right.CheckedOutByPrincipal &&
		left.CheckoutDetails == right.CheckoutDetails &&
		left.ReturnedAt.Equal(right.ReturnedAt) &&
		left.ReturnedByPrincipal == right.ReturnedByPrincipal &&
		left.ReturnDetails == right.ReturnDetails &&
		left.CreatedAt.Equal(right.CreatedAt) &&
		left.UpdatedAt.Equal(right.UpdatedAt)
}

package ports

import (
	"context"
	"time"
)

type Scope struct{ Tenant, Inventory string }
type Page struct {
	Limit  int64
	Cursor string
}
type Pagination struct {
	Limit      int64   `json:"limit"`
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
}
type Inventory struct {
	TenantID  string `json:"tenantId"`
	Access    Access `json:"access"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Lifecycle string `json:"lifecycleState"`
}
type Asset struct {
	PrintJobID string `json:"printJobId,omitempty"`
	ID         string `json:"id"`
	Title      string `json:"title"`
	Kind       string `json:"kind"`
	Parent     string `json:"parentAssetId,omitempty"`
	Lifecycle  string `json:"lifecycleState"`
}
type Metadata struct {
	RequestID  *string     `json:"requestId,omitempty"`
	TenantID   *string     `json:"tenantId,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Result[T any] struct {
	Schema     *string     `json:"$schema,omitempty"`
	Meta       *Metadata   `json:"meta,omitempty"`
	Data       T           `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}
type AssetInput struct {
	Kind, Title, Parent string
	PrintLabel          *LabelPrintSelection
}
type AssetChange struct {
	Title      *string
	Parent     *string
	MoveToRoot bool
}
type AuthConfig struct {
	Issuer, ClientID                 string
	Scopes, LoginMethods             []string
	LoopbackHost, LoopbackPathPrefix string
	EphemeralPort                    bool
}
type Session struct {
	Server       string    `json:"server"`
	Issuer       string    `json:"issuer"`
	Subject      string    `json:"subject,omitempty"`
	ClientID     string    `json:"clientId"`
	IDToken      string    `json:"idToken"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	ExpiresAt    time.Time `json:"expiresAt"`
}
type API interface {
	AuthConfig(context.Context) (AuthConfig, error)
	Inventories(context.Context, Scope, Page) (Result[[]Inventory], error)
	Assets(context.Context, Scope, Page) (Result[[]Asset], error)
	Asset(context.Context, Scope, string) (Result[Asset], error)
	CreateAsset(context.Context, Scope, AssetInput, string) (Result[Asset], error)
	UpdateAsset(context.Context, Scope, string, AssetChange, string) (Result[Asset], error)
	SetArchived(context.Context, Scope, string, bool, string) (Result[Asset], error)
}
type Credentials interface {
	Load(context.Context, string) (Session, error)
	Save(context.Context, Session) error
	Delete(context.Context, string) error
}
type Clock interface{ Now() time.Time }
type Auth interface {
	Login(context.Context, string, AuthConfig, bool) (Session, error)
	Refresh(context.Context, Session) (Session, error)
}
type Output interface {
	Result(any) error
	Notice(string) error
	Error(string, string)
}
type Browser interface{ Open(string) error }
type Observer interface{ Event(context.Context, string) }

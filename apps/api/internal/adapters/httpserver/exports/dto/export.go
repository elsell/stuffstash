package dto

type ExportInput struct {
	Authorization string `header:"Authorization" doc:"Bearer access token"`
	RequestID     string `header:"X-Request-ID"`
	TenantID      string `path:"tenantId"`
	InventoryID   string `path:"inventoryId"`
	Format        string `query:"format" enum:"json,csv" default:"json"`
}
type ExportOutput struct {
	ContentType        string `header:"Content-Type"`
	ContentDisposition string `header:"Content-Disposition"`
	CacheControl       string `header:"Cache-Control"`
	Body               []byte
}

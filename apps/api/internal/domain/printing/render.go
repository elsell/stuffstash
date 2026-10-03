package printing

// ContentSnapshot contains only the content explicitly selected for a label.
type ContentSnapshot struct {
	QRURL     string `json:"qr_url"`
	Title     string `json:"title"`
	Reference string `json:"reference"`
}

type Format string

const (
	FormatPNG Format = "png"
	FormatPDF Format = "pdf"
)

type RenderRequest struct {
	Content  ContentSnapshot   `json:"content"`
	Template TemplateSelection `json:"template"`
	Media    MediaSnapshot     `json:"media"`
	Format   Format            `json:"format"`
}

type RenderedLabel struct {
	Content         []byte
	ContentType     string
	WidthPixels     int
	HeightPixels    int
	SHA256          string
	DisplayRotation int
}

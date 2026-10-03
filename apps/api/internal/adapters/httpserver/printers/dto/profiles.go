package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type ListProfilesInput struct {
	Authorization string `header:"Authorization"`
	TenantID      string `path:"tenantId"`
	InventoryID   string `path:"inventoryId"`
}
type ListProfilesOutput struct {
	Body shared.SuccessEnvelope[[]PrinterProfile]
}
type PrinterProfile struct {
	AdapterID          string         `json:"adapterId"`
	Name               string         `json:"name"`
	Transport          string         `json:"transport"`
	SupportedPlatforms []string       `json:"supportedPlatforms"`
	PhysicallyVerified bool           `json:"physicallyVerified"`
	Media              []MediaProfile `json:"media"`
}
type MediaMargins struct {
	Left   int `json:"left"`
	Right  int `json:"right"`
	Top    int `json:"top"`
	Bottom int `json:"bottom"`
}
type MediaProfile struct {
	Name               string       `json:"name"`
	PresetID           string       `json:"presetId"`
	Version            uint32       `json:"version"`
	WidthMicrometers   int          `json:"widthMicrometers"`
	HeightMicrometers  int          `json:"heightMicrometers"`
	MarginsMicrometers MediaMargins `json:"marginsMicrometers"`
	ResolutionDPI      int          `json:"resolutionDpi"`
	RasterWidth        int          `json:"rasterWidth"`
	RasterHeight       int          `json:"rasterHeight"`
	Orientation        string       `json:"orientation"`
	ColorMode          string       `json:"colorMode"`
	CutPolicy          string       `json:"cutPolicy"`
	DisplayRotation    int          `json:"displayRotation"`
}

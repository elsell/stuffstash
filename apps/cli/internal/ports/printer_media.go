package ports

type PrinterMargins struct {
	Bottom int64 `json:"bottom"`
	Left   int64 `json:"left"`
	Right  int64 `json:"right"`
	Top    int64 `json:"top"`
}
type PrinterMedia struct {
	ColorMode          string         `json:"colorMode"`
	CutPolicy          string         `json:"cutPolicy"`
	DisplayRotation    int64          `json:"displayRotation"`
	HeightMicrometers  int64          `json:"heightMicrometers"`
	MarginsMicrometers PrinterMargins `json:"marginsMicrometers"`
	Name               string         `json:"name"`
	Orientation        string         `json:"orientation"`
	PresetID           string         `json:"presetId"`
	RasterHeight       int64          `json:"rasterHeight"`
	RasterWidth        int64          `json:"rasterWidth"`
	ResolutionDpi      int64          `json:"resolutionDpi"`
	Version            uint32         `json:"version"`
	WidthMicrometers   int64          `json:"widthMicrometers"`
}

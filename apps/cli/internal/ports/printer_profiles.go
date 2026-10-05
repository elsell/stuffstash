package ports

type PrinterProfile struct {
	AdapterID          string         `json:"adapterId"`
	Name               string         `json:"name"`
	Transport          string         `json:"transport"`
	PhysicallyVerified bool           `json:"physicallyVerified"`
	SupportedPlatforms []string       `json:"supportedPlatforms"`
	Media              []PrinterMedia `json:"media"`
}

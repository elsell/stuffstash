package printing

type MediaProfile struct {
	Name  string
	Media MediaSnapshot
}
type PrinterProfile struct {
	AdapterID, Name, Transport string
	SupportedPlatforms         []string
	Media                      []MediaProfile
	PhysicallyVerified         bool
}

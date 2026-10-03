package printing

// RegisteredPrinter is the authorized API snapshot; physical identity is never
// selected from a print job or inferred from a display name.
type RegisteredPrinter struct {
	ID, AdapterID, DeviceID, MediaFingerprint string
	BindingGeneration                         uint64
	Media                                     Media
	Retired                                   bool
}

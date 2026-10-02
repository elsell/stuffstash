package ports

// ArchiveRestorePlan is private application state, not the public archive wire
// representation. Persist it before approval; retries reuse these same IDs.
type ArchiveRestorePlan struct {
	Document                 InventoryExportDocument
	KeyRemappings            []ArchiveKeyRemapping
	OmittedAttachments       int
	CheckoutSourcePrincipals map[string]string
}
type ArchiveKeyRemapping struct{ Family, SourceKey, DestinationKey string }
type ArchiveKeyReservations struct{ Fields, Types map[string]bool }
type ArchiveMediaSelection struct{ Photos, OtherFiles bool }
type ArchiveRestoreDestination struct{ TenantID, InventoryID, Name, PrincipalID string }

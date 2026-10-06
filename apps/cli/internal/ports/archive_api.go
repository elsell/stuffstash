package ports

// ArchiveAPI composes the existing archive metadata, writes and stream ports.
type ArchiveAPI interface {
	ArchiveJobsAPI
	ArchiveWrites
	ArchiveTransfers
}

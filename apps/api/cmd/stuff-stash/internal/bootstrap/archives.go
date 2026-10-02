package bootstrap

import (
	"fmt"
	"os"

	"github.com/stuffstash/stuff-stash/internal/adapters/blobstore"
	"github.com/stuffstash/stuff-stash/internal/adapters/idgen"
	"github.com/stuffstash/stuff-stash/internal/adapters/inventoryarchive"
	"github.com/stuffstash/stuff-stash/internal/app/dataportability"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func buildArchiveRuntime(cfg config.ArchiveConfig, repos repositories, authorizer ports.Authorizer, observer ports.Observer) (*dataportability.ArchiveService, *dataportability.ArchiveWorker, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	if !cfg.Enabled {
		return nil, nil, nil
	}
	storage := repos.archiveBlobs
	if storage == nil || repos.archivePublication == nil {
		return nil, nil, fmt.Errorf("archive runtime requires streaming storage and durable repositories")
	}
	if err := os.MkdirAll(cfg.ScratchDirectory, 0700); err != nil {
		return nil, nil, err
	}
	clock := ports.SystemClock{}
	service, err := dataportability.NewArchiveService(dataportability.ArchiveDependencies{Jobs: repos.archiveJobs, Artifacts: repos.archiveArtifacts, Commands: repos.archiveCommands, Audit: repos.audit, Plans: inventoryarchive.PlanCodec{}, MaxMetadataBytes: cfg.MetadataBytes, MaxRecords: cfg.MaxRecords, Authorizer: authorizer, Inventories: repos.inventories, Tenants: repos.tenants, IDs: idgen.NewULIDGenerator(), Clock: clock, Storage: storage, Scratch: blobstore.ScratchSpace{Directory: cfg.ScratchDirectory}, MaxArchiveBytes: cfg.MaxBytes, Retention: cfg.Retention, CleanupTimeout: cfg.CleanupInterval, Observer: observer})
	if err != nil {
		return nil, nil, err
	}
	worker, err := dataportability.NewArchiveWorker(dataportability.ArchiveWorkerDependencies{Service: service, Snapshots: repos.archiveSnapshots, Metadata: inventoryarchive.MetadataCodec{}, Packages: inventoryarchive.PackageCodec{}, Readers: inventoryarchive.PackageCodec{}, Plans: inventoryarchive.PlanCodec{}, Fields: repos.customFields, Types: repos.customAssetTypes, Publisher: repos.archivePublication(clock, cfg.MaxRecords), Limits: ports.ArchivePackageLimits{CompressedBytes: cfg.MaxBytes, ExpandedBytes: cfg.ExpandedBytes, MetadataBytes: cfg.MetadataBytes, EntryBytes: cfg.EntryBytes, Entries: cfg.MaxEntries}, MaxRecords: cfg.MaxRecords, LeaseDuration: cfg.Lease, HeartbeatInterval: cfg.Heartbeat})
	if err != nil {
		return nil, nil, err
	}
	return &service, &worker, nil
}

package gormstore

import "time"

type importJobModel struct {
	ID                        string         `gorm:"primaryKey;size:26"`
	TenantID                  string         `gorm:"not null;size:26;index:idx_import_jobs_inventory_created,priority:1;index"`
	Tenant                    tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID               string         `gorm:"not null;size:26;index:idx_import_jobs_inventory_created,priority:2;index"`
	Inventory                 inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	ActorID                   string         `gorm:"not null;size:128;index"`
	Status                    string         `gorm:"not null;size:64;index"`
	SourceType                string         `gorm:"not null;size:64"`
	SourceName                string         `gorm:"not null;size:128"`
	SourceBaseURL             string         `gorm:"not null;size:2048"`
	SourceVersion             string         `gorm:"not null;size:128"`
	SourceImageImport         string         `gorm:"not null;size:64"`
	SourceAllowPrivateNetwork bool           `gorm:"not null;default:false"`
	SourceAllowInsecureTLS    bool           `gorm:"not null;default:false"`
	SourceFingerprint         string         `gorm:"not null;size:128"`
	Fields                    int            `gorm:"not null"`
	Locations                 int            `gorm:"not null"`
	Assets                    int            `gorm:"not null"`
	Attachments               int            `gorm:"not null"`
	Warnings                  int            `gorm:"not null"`
	Errors                    int            `gorm:"not null"`
	FieldsCreated             int            `gorm:"not null"`
	FieldsExisting            int            `gorm:"not null"`
	LocationsCreated          int            `gorm:"not null"`
	AssetsCreated             int            `gorm:"not null"`
	AssetsSkipped             int            `gorm:"not null"`
	AttachmentsCreated        int            `gorm:"not null"`
	AttachmentsSkipped        int            `gorm:"not null"`
	RecordsDiscarded          int            `gorm:"not null"`
	SourceLinksDiscarded      int            `gorm:"not null"`
	PreviewJSON               []byte
	ProgressPhase             string `gorm:"not null;size:64"`
	ProgressDone              int    `gorm:"not null"`
	ProgressTotal             int    `gorm:"not null"`
	ProgressMessage           string `gorm:"not null;size:512"`
	ProgressUpdatedAt         *time.Time
	ProgressHistoryJSON       []byte
	CancellationMode          string `gorm:"not null;size:64"`
	CancellationRequestID     string `gorm:"not null;size:128;default:''"`
	MessagesJSON              []byte `gorm:"not null"`
	HistoryRemovedAt          *time.Time
	StartedAt                 *time.Time
	CompletedAt               *time.Time
	CreatedAt                 time.Time `gorm:"not null;index:idx_import_jobs_inventory_created,priority:3"`
	UpdatedAt                 time.Time `gorm:"not null"`
}

func (importJobModel) TableName() string {
	return "import_jobs"
}

type importJobSourceModel struct {
	JobID       string         `gorm:"primaryKey;size:26"`
	TenantID    string         `gorm:"not null;size:26;index"`
	Tenant      tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID string         `gorm:"not null;size:26;index"`
	Inventory   inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	KeyID       string         `gorm:"not null;size:128"`
	Algorithm   string         `gorm:"not null;size:64"`
	Nonce       []byte         `gorm:"not null"`
	Ciphertext  []byte         `gorm:"not null"`
	ExpiresAt   time.Time      `gorm:"not null;index"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
}

func (importJobSourceModel) TableName() string {
	return "import_job_sources"
}

type importSourceLinkModel struct {
	ID                uint           `gorm:"primaryKey"`
	TenantID          string         `gorm:"not null;size:26;index:idx_import_source_links_source,unique,priority:1;index"`
	Tenant            tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID       string         `gorm:"not null;size:26;index:idx_import_source_links_source,unique,priority:2;index"`
	Inventory         inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	JobID             string         `gorm:"not null;size:26;index"`
	Job               importJobModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:JobID;references:ID"`
	SourceType        string         `gorm:"not null;size:64;index:idx_import_source_links_source,unique,priority:3"`
	SourceInstanceKey string         `gorm:"not null;size:2048;index:idx_import_source_links_source,unique,priority:4"`
	SourceEntityType  string         `gorm:"not null;size:64;index:idx_import_source_links_source,unique,priority:5"`
	SourceEntityID    string         `gorm:"not null;size:512;index:idx_import_source_links_source,unique,priority:6"`
	ResourceType      string         `gorm:"not null;size:64"`
	ResourceID        string         `gorm:"not null;size:64;index"`
	CreatedAt         time.Time      `gorm:"not null"`
}

func (importSourceLinkModel) TableName() string {
	return "import_source_links"
}

type importJobResourceModel struct {
	ID                uint           `gorm:"primaryKey"`
	TenantID          string         `gorm:"not null;size:26;index:idx_import_job_resources_job,priority:1;index"`
	Tenant            tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID       string         `gorm:"not null;size:26;index:idx_import_job_resources_job,priority:2;index"`
	Inventory         inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:InventoryID;references:ID"`
	JobID             string         `gorm:"not null;size:26;index:idx_import_job_resources_job,priority:3"`
	Job               importJobModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:JobID;references:ID"`
	ResourceType      string         `gorm:"not null;size:64;index:idx_import_job_resources_resource,unique,priority:1"`
	ResourceID        string         `gorm:"not null;size:64;index:idx_import_job_resources_resource,unique,priority:2"`
	ResourceOwnerID   string         `gorm:"not null;size:64"`
	SourceType        string         `gorm:"not null;size:64"`
	SourceInstanceKey string         `gorm:"not null;size:2048"`
	SourceEntityType  string         `gorm:"not null;size:64"`
	SourceEntityID    string         `gorm:"not null;size:512"`
	CreatedAt         time.Time      `gorm:"not null"`
}

func (importJobResourceModel) TableName() string {
	return "import_job_resources"
}

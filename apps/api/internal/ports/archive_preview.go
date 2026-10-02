package ports

type ArchivePreview struct {
	InventoryName                                                                        string
	Assets, Tags, CustomAssetTypes, CustomFields, Photos, OtherFiles, OmittedAttachments int
	KeyRemappings                                                                        []ArchiveKeyRemapping
}

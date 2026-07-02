package shared

type AssetPrimaryPhoto struct {
	ID          string               `json:"id"`
	FileName    string               `json:"fileName"`
	ContentType string               `json:"contentType"`
	SizeBytes   int64                `json:"sizeBytes"`
	Thumbnails  AssetPhotoThumbnails `json:"thumbnails"`
}

type AssetPhotoThumbnails struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

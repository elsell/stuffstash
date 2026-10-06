package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"time"
)

func (o Output) archiveJob(v ports.ArchiveJob) error {
	value := func(p *string) string {
		if p == nil {
			return "None"
		}
		return *p
	}
	return o.details([][2]string{{"Job", v.ID}, {"Kind", v.Kind}, {"State", v.State}, {"Phase", v.Phase}, {"Inventory", value(v.InventoryID)}, {"Destination inventory", value(v.DestinationInventoryID)}, {"Created", v.CreatedAt.Format(time.RFC3339Nano)}, {"Expires", v.ExpiresAt.Format(time.RFC3339Nano)}, {"Photos included", strconv.FormatBool(v.Photos)}, {"Other files included", strconv.FormatBool(v.OtherFiles)}, {"Failure", value(v.Failure)}})
}
func (o Output) archivePreview(v ports.ArchivePreview) error {
	if err := o.details([][2]string{{"Inventory name", v.InventoryName}, {"Assets", strconv.FormatInt(v.Assets, 10)}, {"Tags", strconv.FormatInt(v.Tags, 10)}, {"Custom fields", strconv.FormatInt(v.CustomFields, 10)}, {"Custom asset types", strconv.FormatInt(v.CustomAssetTypes, 10)}, {"Photos", strconv.FormatInt(v.Photos, 10)}, {"Other files", strconv.FormatInt(v.OtherFiles, 10)}, {"Omitted attachments", strconv.FormatInt(v.OmittedAttachments, 10)}}); err != nil {
		return err
	}
	for _, key := range v.KeyRemappings {
		if _, err := fmt.Fprintf(o.Stdout, "Remap %s: %s → %s\n", strconv.Quote(key.Family), strconv.Quote(key.SourceKey), strconv.Quote(key.DestinationKey)); err != nil {
			return err
		}
	}
	return nil
}

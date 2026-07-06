package homebox

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type LegacyImporter struct {
	client              *http.Client
	maxAttachmentBytes  int64
	allowPrivateNetwork bool
}

const defaultMaxAttachmentBytes int64 = 25 * 1024 * 1024

func NewLegacyImporter(client *http.Client) LegacyImporter {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return LegacyImporter{client: client, maxAttachmentBytes: defaultMaxAttachmentBytes}
}

func (i LegacyImporter) ReadImportPlan(ctx context.Context, request ports.ImportSourceRequest) (importplan.Plan, error) {
	switch request.SourceType {
	case importplan.SourceLegacyHomeboxCSV:
		return i.readCSV(request)
	case importplan.SourceLegacyHomebox:
		return i.readLive(ctx, request)
	default:
		return importplan.Plan{}, errors.New("unsupported import source")
	}
}

func (i LegacyImporter) readCSV(request ports.ImportSourceRequest) (importplan.Plan, error) {
	reader := csv.NewReader(bytes.NewReader(request.Content))
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return importplan.Plan{}, err
	}
	if len(rows) == 0 {
		return importplan.Plan{}, errors.New("empty Homebox CSV")
	}
	header := map[string]int{}
	for index, column := range rows[0] {
		header[strings.TrimSpace(column)] = index
	}
	required := []string{"HB.location", "HB.asset_id", "HB.name"}
	var messages []importplan.Message
	for _, column := range required {
		if _, ok := header[column]; !ok {
			messages = append(messages, importplan.Message{
				Code:     "missing-column",
				Severity: importplan.SeverityError,
				Summary:  "Missing Homebox CSV column",
				Detail:   column,
			})
		}
	}
	plan := importplan.Plan{
		Source: importplan.SourceSummary{
			Type:        importplan.SourceLegacyHomeboxCSV,
			Name:        safeFileName(request.FileName, "homebox-export.csv"),
			ImageImport: "unavailable",
		},
		Fields:   homeboxFields(),
		Messages: messages,
	}
	if len(messages) > 0 {
		return plan, nil
	}

	locationIDs := map[string]string{}
	for rowIndex, row := range rows[1:] {
		name := getCSV(row, header, "HB.name")
		if strings.TrimSpace(name) == "" {
			continue
		}
		locationName := getCSV(row, header, "HB.location")
		parentSourceID := ""
		if strings.TrimSpace(locationName) != "" {
			parentSourceID = locationSourceID(locationName)
			if _, seen := locationIDs[parentSourceID]; !seen {
				locationIDs[parentSourceID] = locationName
				plan.Assets = append(plan.Assets, importplan.Asset{
					SourceID:     parentSourceID,
					Kind:         "location",
					Title:        locationName,
					CustomFields: sourceReferenceFields(parentSourceID),
				})
			}
		}
		sourceID := getCSV(row, header, "HB.import_ref")
		if strings.TrimSpace(sourceID) == "" {
			sourceID = getCSV(row, header, "HB.asset_id")
		}
		if strings.TrimSpace(sourceID) == "" {
			sourceID = fmt.Sprintf("csv-row-%d", rowIndex+2)
		}
		customFields, rowMessages := customFieldsFromLegacyValues(legacyValues{
			AssetID:          getCSV(row, header, "HB.asset_id"),
			Tags:             getCSV(row, header, "HB.tags"),
			Quantity:         getCSV(row, header, "HB.quantity"),
			Insured:          getCSV(row, header, "HB.insured"),
			Notes:            getCSV(row, header, "HB.notes"),
			PurchasePrice:    getCSV(row, header, "HB.purchase_price"),
			PurchaseFrom:     getCSV(row, header, "HB.purchase_from"),
			PurchaseTime:     getCSV(row, header, "HB.purchase_time"),
			Manufacturer:     getCSV(row, header, "HB.manufacturer"),
			ModelNumber:      getCSV(row, header, "HB.model_number"),
			SerialNumber:     getCSV(row, header, "HB.serial_number"),
			LifetimeWarranty: getCSV(row, header, "HB.lifetime_warranty"),
			WarrantyExpires:  getCSV(row, header, "HB.warranty_expires"),
			WarrantyDetails:  getCSV(row, header, "HB.warranty_details"),
			SoldTo:           getCSV(row, header, "HB.sold_to"),
			SoldPrice:        getCSV(row, header, "HB.sold_price"),
			SoldTime:         getCSV(row, header, "HB.sold_time"),
			SoldNotes:        getCSV(row, header, "HB.sold_notes"),
		}, sourceID, name)
		plan.Messages = append(plan.Messages, rowMessages...)
		plan.Assets = append(plan.Assets, importplan.Asset{
			SourceID:       "item:" + sourceID,
			SourceRef:      sourceID,
			Kind:           "item",
			Title:          name,
			Description:    getCSV(row, header, "HB.description"),
			ParentSourceID: parentSourceID,
			Archived:       parseBool(getCSV(row, header, "HB.archived")),
			CustomFields:   customFields,
		})
	}
	return plan, nil
}

func (i LegacyImporter) readLive(ctx context.Context, request ports.ImportSourceRequest) (importplan.Plan, error) {
	importer := i.withRequestOptions(request)
	importer.maxAttachmentBytes = normalizedMaxAttachmentBytes(request.MaxAttachmentBytes)
	importer.allowPrivateNetwork = request.AllowPrivateNetwork
	baseURL, err := normalizeBaseURL(request.BaseURL)
	if err != nil {
		return importplan.Plan{}, err
	}
	token, status, err := importer.login(ctx, baseURL, request.Username, request.Password)
	if err != nil {
		return importplan.Plan{}, safeLiveSourceError(err)
	}
	plan := importplan.Plan{
		Source: importplan.SourceSummary{
			Type:        importplan.SourceLegacyHomebox,
			Name:        "Homebox",
			BaseURL:     baseURL,
			Version:     status.Build.Version,
			ImageImport: imageImportLabel(request.IncludeImages),
		},
		Fields: homeboxFields(),
	}
	locations, err := importer.locations(ctx, baseURL, token)
	if err != nil {
		return importplan.Plan{}, safeLiveSourceError(err)
	}
	locationTree, err := importer.locationTree(ctx, baseURL, token)
	if err != nil {
		return importplan.Plan{}, safeLiveSourceError(err)
	}
	locationsByID := map[string]legacyLocation{}
	for _, location := range locations {
		locationsByID[location.ID] = location
	}
	seenLocations := map[string]struct{}{}
	var addLocation func(node legacyTreeNode, parentID string)
	addLocation = func(node legacyTreeNode, parentID string) {
		if strings.TrimSpace(node.ID) == "" {
			return
		}
		seenLocations[node.ID] = struct{}{}
		location := locationsByID[node.ID]
		title := firstNonEmpty(location.Name, node.Name)
		description := location.Description
		plan.Assets = append(plan.Assets, importplan.Asset{
			SourceID:       locationSourceID(node.ID),
			Kind:           "location",
			Title:          title,
			Description:    description,
			ParentSourceID: parentID,
			CustomFields:   sourceReferenceFields(locationSourceID(node.ID)),
		})
		for _, child := range node.Children {
			addLocation(child, locationSourceID(node.ID))
		}
	}
	for _, node := range locationTree {
		addLocation(node, "")
	}
	for _, location := range locations {
		if _, ok := seenLocations[location.ID]; ok {
			continue
		}
		plan.Assets = append(plan.Assets, importplan.Asset{
			SourceID:     locationSourceID(location.ID),
			Kind:         "location",
			Title:        location.Name,
			Description:  location.Description,
			CustomFields: sourceReferenceFields(locationSourceID(location.ID)),
		})
	}

	items, err := importer.items(ctx, baseURL, token)
	if err != nil {
		return importplan.Plan{}, err
	}
	sort.Slice(items, func(left, right int) bool {
		return items[left].AssetID < items[right].AssetID
	})
	for _, summary := range items {
		detail, err := importer.item(ctx, baseURL, token, summary.ID)
		if err != nil {
			plan.Messages = append(plan.Messages, importplan.Message{
				Code:       "item-detail-unavailable",
				Severity:   importplan.SeverityWarning,
				Summary:    "Item detail could not be read",
				Detail:     safeHomeboxWarningDetail(err, "item detail could not be read"),
				SourceID:   summary.ID,
				SourceName: summary.Name,
			})
			continue
		}
		values := legacyValuesFromItem(detail)
		customFields, rowMessages := customFieldsFromLegacyValues(values, detail.ID, detail.Name)
		plan.Messages = append(plan.Messages, rowMessages...)
		parentSourceID := ""
		if detail.Location.ID != "" {
			parentSourceID = locationSourceID(detail.Location.ID)
		}
		plan.Assets = append(plan.Assets, importplan.Asset{
			SourceID:       "item:" + detail.ID,
			SourceRef:      firstNonEmpty(detail.AssetID, detail.ID),
			Kind:           "item",
			Title:          detail.Name,
			Description:    detail.Description,
			ParentSourceID: parentSourceID,
			Archived:       detail.Archived,
			CustomFields:   customFields,
		})
		if request.IncludeImages && !detail.Archived {
			for _, attachment := range detail.Attachments {
				if attachment.Type != "" && attachment.Type != "photo" {
					continue
				}
				planned := importplan.Attachment{
					SourceID:      attachment.ID,
					AssetSourceID: "item:" + detail.ID,
					FileName:      safeFileName(attachment.Title, defaultImageName(attachment.MIMEType)),
					ContentType:   attachment.MIMEType,
					Primary:       attachment.Primary,
				}
				if !request.FetchAttachmentBytes {
					plan.Attachments = append(plan.Attachments, planned)
					continue
				}
				content, contentType, err := importer.attachment(ctx, baseURL, token, detail.ID, attachment.ID)
				if err != nil {
					planned.UnavailableReason = safeHomeboxWarningDetail(err, "attachment could not be downloaded")
					plan.Attachments = append(plan.Attachments, planned)
					continue
				}
				if !supportedImageType(contentType) {
					plan.Messages = append(plan.Messages, importplan.Message{
						Code:       "attachment-unsupported-type",
						Severity:   importplan.SeverityWarning,
						Summary:    "Attachment type is not supported",
						Detail:     contentType,
						SourceID:   attachment.ID,
						SourceName: detail.Name,
					})
					continue
				}
				planned.FileName = safeFileName(attachment.Title, defaultImageName(contentType))
				planned.ContentType = contentType
				planned.Content = content
				planned.SizeBytes = len(content)
				plan.Attachments = append(plan.Attachments, planned)
			}
		}
	}
	return plan, nil
}

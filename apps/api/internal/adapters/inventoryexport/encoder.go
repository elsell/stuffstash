// Package inventoryexport implements portable inventory file formats.
package inventoryexport

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"
	"unicode"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Encoder struct{}

func (Encoder) Encode(ctx context.Context, document ports.InventoryExportDocument, format ports.InventoryExportFormat, maxBytes int) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if maxBytes <= 0 {
		return nil, ports.ErrInventoryExportLimit
	}
	output := &boundedWriter{ctx: ctx, limit: maxBytes}
	var err error
	switch format {
	case ports.InventoryExportJSON:
		err = writeJSON(output, document)
	case ports.InventoryExportCSV:
		err = writeCSV(output, document)
	default:
		return nil, ports.ErrInventoryExportFormat
	}
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

type boundedWriter struct {
	bytes.Buffer
	ctx   context.Context
	limit int
}

func (w *boundedWriter) Write(p []byte) (int, error) {
	if err := w.ctx.Err(); err != nil {
		return 0, err
	}
	if len(p) > w.limit-w.Len() {
		return 0, ports.ErrInventoryExportLimit
	}
	return w.Buffer.Write(p)
}

func (w *boundedWriter) WriteString(value string) (int, error) { return w.Write([]byte(value)) }

func writeJSON(w io.Writer, d ports.InventoryExportDocument) error {
	// Encode collections independently so the JSON encoder never duplicates the
	// complete document in one temporary allocation. Publication remains all-or-none.
	header := map[string]any{"schemaVersion": d.SchemaVersion, "exportedAt": d.ExportedAt, "tenantId": d.TenantID, "inventoryId": d.InventoryID, "inventoryName": d.InventoryName}
	b, err := json.Marshal(header)
	if err != nil {
		return err
	}
	if _, err = w.Write(b[:len(b)-1]); err != nil {
		return err
	}
	collections := []struct {
		name   string
		length int
		value  func(int) any
	}{
		{"assets", len(d.Assets), func(i int) any { return assetRecord(d.Assets[i]) }},
		{"tags", len(d.Tags), func(i int) any { return tagRecord(d.Tags[i]) }},
		{"customAssetTypes", len(d.CustomAssetTypes), func(i int) any { return assetTypeRecord(d.CustomAssetTypes[i]) }},
		{"customFieldDefinitions", len(d.CustomFieldDefinitions), func(i int) any { return fieldRecord(d.CustomFieldDefinitions[i]) }},
	}
	for _, collection := range collections {
		if _, err = io.WriteString(w, ",\""+collection.name+"\":["); err != nil {
			return err
		}
		for i := 0; i < collection.length; i++ {
			if i > 0 {
				if _, err = io.WriteString(w, ","); err != nil {
					return err
				}
			}
			if err = json.NewEncoder(w).Encode(collection.value(i)); err != nil {
				return err
			}
		}
		if _, err = io.WriteString(w, "]"); err != nil {
			return err
		}
	}
	_, err = io.WriteString(w, "}\n")
	return err
}

var csvColumns = []string{"id", "title", "description", "kind", "parentAssetId", "customAssetTypeId", "lifecycleState", "createdAt", "updatedAt", "expirationDate", "expirationPrecision", "tagIds", "customFields", "currentCheckout", "attachments"}

func writeCSV(w io.Writer, d ports.InventoryExportDocument) error {
	writer := csv.NewWriter(w)
	if err := writer.Write(csvColumns); err != nil {
		return err
	}
	for _, item := range d.Assets {
		record := assetRecord(item)
		row := make([]string, len(csvColumns))
		for i, column := range csvColumns {
			if value, ok := record[column].(string); ok {
				row[i] = safeCell(value)
			} else {
				data, err := json.Marshal(record[column])
				if err != nil {
					return err
				}
				row[i] = safeCell(string(data))
			}
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
func safeCell(value string) string {
	trimmed := strings.TrimLeftFunc(value, unicode.IsSpace)
	if strings.ContainsAny(value, "\t\r") || (len(trimmed) > 0 && strings.ContainsRune("=+-@", rune(trimmed[0]))) {
		return "'" + value
	}
	return value
}

package inventoryarchive

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"unicode"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

const metadataVersion = 2

type MetadataCodec struct{}

var _ ports.ArchiveMetadataCodec = MetadataCodec{}

func (MetadataCodec) EncodeMetadata(ctx context.Context, d ports.InventoryExportDocument, maxBytes int) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, ErrLimit
	}
	if d.SchemaVersion != metadataVersion {
		return nil, ErrInvalid
	}
	var output bytes.Buffer
	w := &boundedWriter{ctx: ctx, w: &output, remaining: int64(maxBytes)}
	header := map[string]any{"schemaVersion": metadataVersion, "exportedAt": d.ExportedAt, "tenantId": d.TenantID, "inventoryId": d.InventoryID, "inventoryName": d.InventoryName}
	data, err := json.Marshal(header)
	if err != nil {
		return nil, err
	}
	if _, err = w.Write(data[:len(data)-1]); err != nil {
		return nil, err
	}
	collections := []struct {
		name  string
		size  int
		value func(int) any
	}{
		{"assets", len(d.Assets), func(i int) any { return projectAsset(d.Assets[i]) }},
		{"tags", len(d.Tags), func(i int) any { return projectmetadataTag(d.Tags[i]) }},
		{"customAssetTypes", len(d.CustomAssetTypes), func(i int) any { return projectmetadataType(d.CustomAssetTypes[i]) }},
		{"customFieldDefinitions", len(d.CustomFieldDefinitions), func(i int) any { return projectmetadataField(d.CustomFieldDefinitions[i]) }},
	}
	for _, c := range collections {
		if _, err = io.WriteString(w, ",\""+c.name+"\":["); err != nil {
			return nil, err
		}
		for i := 0; i < c.size; i++ {
			if i > 0 {
				if _, err = io.WriteString(w, ","); err != nil {
					return nil, err
				}
			}
			if err = json.NewEncoder(w).Encode(c.value(i)); err != nil {
				return nil, err
			}
		}
		if _, err = io.WriteString(w, "]"); err != nil {
			return nil, err
		}
	}
	if _, err = io.WriteString(w, "}\n"); err != nil {
		return nil, err
	}
	return output.Bytes(), ctx.Err()
}

func (MetadataCodec) DecodeMetadata(ctx context.Context, data []byte, maxBytes int) (ports.InventoryExportDocument, error) {
	empty := ports.InventoryExportDocument{}
	if maxBytes <= 0 || len(data) > maxBytes {
		return empty, ErrLimit
	}
	if err := uniqueJSON(ctx, data); err != nil {
		return empty, err
	}
	var wire metadataDocument
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(&wire); err != nil {
		return empty, ErrInvalid
	}
	if wire.SchemaVersion != metadataVersion || wire.Assets == nil || wire.Tags == nil || wire.CustomAssetTypes == nil || wire.CustomFieldDefinitions == nil {
		return empty, ErrInvalid
	}
	result := ports.InventoryExportDocument{SchemaVersion: wire.SchemaVersion, ExportedAt: wire.ExportedAt, TenantID: wire.TenantID, InventoryID: wire.InventoryID, InventoryName: wire.InventoryName}
	for _, v := range wire.Assets {
		result.Assets = append(result.Assets, readAsset(v))
	}
	for _, v := range wire.Tags {
		result.Tags = append(result.Tags, readmetadataTag(v))
	}
	for _, v := range wire.CustomAssetTypes {
		result.CustomAssetTypes = append(result.CustomAssetTypes, readmetadataType(v))
	}
	for _, v := range wire.CustomFieldDefinitions {
		result.CustomFieldDefinitions = append(result.CustomFieldDefinitions, readmetadataField(v))
	}
	return result, ctx.Err()
}

// Check ambiguity before decoding into structs: encoding/json otherwise silently
// accepts repeated keys. Depth is bounded independently of the byte limit.
func uniqueJSON(ctx context.Context, data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := uniqueValue(ctx, decoder, 0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return ErrInvalid
	}
	return nil
}
func uniqueValue(ctx context.Context, d *json.Decoder, depth int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth > 64 {
		return ErrLimit
	}
	token, err := d.Token()
	if err != nil {
		return ErrInvalid
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return ErrInvalid
			}
			name, ok := key.(string)
			if !ok || seen[foldJSONKey(name)] {
				return ErrInvalid
			}
			seen[foldJSONKey(name)] = true
			if err := uniqueValue(ctx, d, depth+1); err != nil {
				return err
			}
		}
		token, err = d.Token()
		if err != nil || token != json.Delim('}') {
			return ErrInvalid
		}
	case '[':
		for d.More() {
			if err := uniqueValue(ctx, d, depth+1); err != nil {
				return err
			}
		}
		token, err = d.Token()
		if err != nil || token != json.Delim(']') {
			return ErrInvalid
		}
	default:
		return ErrInvalid
	}
	return nil
}

// Match encoding/json's Unicode case folding, including long-s and Kelvin sign.
func foldJSONKey(key string) string {
	folded := []rune(key)
	for i, r := range folded {
		smallest := r
		for next := unicode.SimpleFold(r); next != r; next = unicode.SimpleFold(next) {
			if next < smallest {
				smallest = next
			}
		}
		folded[i] = smallest
	}
	return string(folded)
}

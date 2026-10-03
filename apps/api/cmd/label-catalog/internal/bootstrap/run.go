package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/stuffstash/stuff-stash/internal/adapters/labelrenderer"
	"github.com/stuffstash/stuff-stash/internal/app/printingcatalog"
)

// Run is intentionally offline. A media catalog is optional for descriptor-only
// export; providing one renders every advertised template/media combination.
func Run(ctx context.Context, args []string, output io.Writer) error {
	flags := flag.NewFlagSet("label-catalog", flag.ContinueOnError)
	flags.SetOutput(output)
	mediaPath := flags.String("media", "", "versioned printer media catalog JSON")
	outputPath := flags.String("output", "", "directory for generated catalog and fixtures")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *outputPath == "" {
		return errors.New("usage: label-catalog --output <directory> [--media <catalog.json>]")
	}
	var media printingcatalog.MediaCatalog
	if *mediaPath != "" {
		file, err := os.Open(*mediaPath)
		if err != nil {
			return err
		}
		defer file.Close()
		decoder := json.NewDecoder(io.LimitReader(file, 1024*1024))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&media); err != nil {
			return err
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			return errors.New("media catalog contains trailing data or exceeds one MiB")
		}
		if media.SchemaVersion != printingcatalog.SchemaVersion {
			return errors.New("unsupported media catalog schema version")
		}
	}
	renderer, err := labelrenderer.New(labelrenderer.DefaultLimits())
	if err != nil {
		return err
	}
	bundle, err := printingcatalog.Export(ctx, renderer, renderer, media.Media)
	if err != nil {
		return err
	}
	// Refuse a populated directory: generation never clobbers user files or leaves
	// stale artifacts looking current. Drift tooling should use a fresh temp tree.
	entries, err := os.ReadDir(*outputPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(entries) > 0 {
		return errors.New("output directory must be empty")
	}
	if err := os.MkdirAll(*outputPath, 0755); err != nil {
		return err
	}
	names := make([]string, 0, len(bundle.Files))
	for name := range bundle.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(*outputPath, name), bundle.Files[name], 0644); err != nil {
			return err
		}
	}
	return nil
}
